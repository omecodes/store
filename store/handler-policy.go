package store

import (
	"context"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	ome "github.com/omecodes/libome"
	"github.com/omecodes/omestore/ent"
	"github.com/omecodes/omestore/pb"
	"time"
)

type CELParams struct {
	auth  *pb.AuthCEL
	data  *pb.DataCEL
	graft *pb.GraftCEL
	at    int64
}

type policyHandler struct {
	base
}

func (p *policyHandler) evaluate(ctx *context.Context, state *CELParams, rule string) (bool, error) {
	if rule == "" || rule == "false" {
		return false, nil
	}

	if rule == "true" {
		return true, nil
	}

	prg, err := getProgram(ctx, rule)
	if err != nil {
		return false, err
	}

	vars := map[string]interface{}{
		"auth":  state.auth,
		"data":  state.data,
		"graft": state.graft,
		"at":    state.at,
	}
	out, details, err := prg.Eval(vars)

	if err != nil {
		log.Error("cel execution", log.Field("details", details))
		return false, err
	}

	return out.Value().(bool), nil
}

func (p *policyHandler) actionRuleForData(ctx *context.Context, action pb.Action) (string, error) {
	db := getDB(*ctx)
	if db == nil {
		return "", errors.Internal
	}

	route := getRoute(SkipPoliciesCheck(), SkipParamsCheck())
	settings, err := route.GetSettings(*ctx, pb.SettingsOptions{Path: settingsDataAccessSecurityRulesPath})
	if err != nil {
		return "", errors.Internal
	}

	rule, err := settings.String(action.String())
	if err != nil {
		log.Error("could not get 'read' rule for access security rules settings", log.Err(err))
		return "", errors.Internal
	}
	return rule, nil
}

func (p *policyHandler) actionRuleForGraft(ctx *context.Context, action pb.Action) (string, error) {
	db := getDB(*ctx)
	if db == nil {
		return "", errors.Internal
	}

	route := getRoute(SkipPoliciesCheck(), SkipParamsCheck())
	settings, err := route.GetSettings(*ctx, pb.SettingsOptions{Path: settingsGraftAccessSecurityRulesPath})
	if err != nil {
		return "", errors.Internal
	}

	rule, err := settings.String(action.String())
	if err != nil {
		log.Error("could not get 'read' rule for access security rules settings", log.Err(err))
		return "", errors.Internal
	}
	return rule, nil
}

func (p *policyHandler) assertIsAllowedOnData(ctx *context.Context, action pb.Action, collection string, id string) error {
	s := &CELParams{
		at:    time.Now().Unix(),
		data:  &pb.DataCEL{},
		auth:  &pb.AuthCEL{},
		graft: &pb.GraftCEL{},
	}

	// load default 'action' rule for data
	rule, err := p.actionRuleForData(ctx, action)
	if err != nil {
		return err
	}

	if action != pb.Action_create {
		route := getRoute(SkipPoliciesCheck(), SkipParamsCheck())
		info, err := route.Info(*ctx, collection, id)
		if err != nil {
			return err
		}
		*ctx = contextWithDataInfo(*ctx, collection, id, info)
		s.data.Creator = info.CreatedBy
		s.data.Col = collection
		s.data.Id = id
	}

	// loading jwt
	cred := ome.CredentialsFromContext(*ctx)

	if cred != nil {
		route := getRoute(SkipPoliciesCheck())
		info, err := route.UserInfo(*ctx, cred.Username, pb.UserOptions{WithGroups: true})
		if err != nil {
			return err
		}

		s.auth.Uid = info.ID
		s.auth.Email = info.Email
		s.auth.Validated = info.Validated
		if info.Edges.Group != nil {
			s.auth.Group = info.Edges.Group.ID
		}
	}

	allowed, err := p.evaluate(ctx, s, rule)
	if err != nil {
		log.Error("failed to evaluate access rule", log.Err(err))
		return errors.Internal
	}

	if !allowed {
		return errors.Unauthorized
	}

	return nil
}

func (p *policyHandler) assertIsAllowedOnGraft(ctx *context.Context, action pb.Action, collection string, dataID string, id string) error {
	s := &CELParams{
		at:    time.Now().Unix(),
		data:  &pb.DataCEL{},
		auth:  &pb.AuthCEL{},
		graft: &pb.GraftCEL{},
	}

	// load default 'action' rule for data
	rule, err := p.actionRuleForGraft(ctx, action)
	if err != nil {
		return err
	}

	route := getRoute(SkipPoliciesCheck(), SkipParamsCheck())

	info, err := route.Info(*ctx, collection, dataID)
	if err != nil {
		return err
	}
	*ctx = contextWithDataInfo(*ctx, collection, id, info)
	s.data.Creator = info.CreatedBy
	s.data.Col = collection
	s.data.Id = id

	graftInfo, err := route.GraftInfo(*ctx, collection, dataID, id)
	if err != nil {
		return err
	}

	s.graft.Id = id
	s.graft.Creator = graftInfo.CreatedBy
	s.graft.CreatedAt = graftInfo.CreatedAt

	// loading jwt
	cred := ome.CredentialsFromContext(*ctx)

	if cred != nil {
		route := getRoute(SkipPoliciesCheck())
		info, err := route.UserInfo(*ctx, cred.Username, pb.UserOptions{WithGroups: true})
		if err != nil {
			return err
		}

		s.auth.Uid = info.ID
		s.auth.Email = info.Email
		s.auth.Validated = info.Validated
		if info.Edges.Group != nil {
			s.auth.Group = info.Edges.Group.ID
		}
	}

	allowed, err := p.evaluate(ctx, s, rule)
	if err != nil {
		log.Error("failed to evaluate access rule", log.Err(err))
		return errors.Internal
	}

	if !allowed {
		return errors.Unauthorized
	}

	return nil
}

func (p *policyHandler) isAdmin(ctx context.Context) (bool, error) {
	cred := ome.CredentialsFromContext(ctx)
	if cred == nil {
		return false, nil
	}

	adminPwd, ok := getAdminPassword(ctx)
	if !ok {
		log.Info("no admin password in context")
		return false, errors.Internal
	}

	return cred.Username == "admin" && cred.Password == adminPwd, nil
}

// Route Handler methods
func (p *policyHandler) SetSettings(ctx context.Context, value *JSON, opts pb.SettingsOptions) error {
	isAdmin, err := p.isAdmin(ctx)
	if err != nil {
		return err
	}

	if !isAdmin {
		return errors.Forbidden
	}

	env := getCelEnv(ctx)
	if env == nil {
		return errors.Internal
	}

	/*if settings.Data.AccessRules.Default.Objects != nil {
		for _, rule := range settings.Data.AccessRules.Default.Objects {
			_, issues := env.Compile(rule)
			if issues != nil && issues.Err() != nil {
				log.Error("incorrect authorization security rule", log.Field("rule", rule), log.Err(issues.Err()))
				return errors.BadInput
			}
		}
	}

	if settings.Data.AccessRules.Default.Grafts != nil {
		for _, rule := range settings.Data.AccessRules.Default.Grafts {
			_, issues := env.Compile(rule)
			if issues != nil && issues.Err() != nil {
				log.Error("incorrect authoritzation security rule", log.Field("rule", rule), log.Err(issues.Err()))
				return errors.BadInput
			}
		}
	} */
	return p.base.SetSettings(ctx, value, opts)
}

func (p *policyHandler) GetSettings(ctx context.Context, opts pb.SettingsOptions) (*JSON, error) {
	isAdmin, err := p.isAdmin(ctx)
	if err != nil {
		return nil, err
	}

	if !isAdmin {
		return nil, errors.Forbidden
	}

	return p.base.GetSettings(ctx, opts)
}

func (p *policyHandler) RegisterUser(ctx context.Context, user *ent.User, opts pb.UserOptions) error {
	user.Validated = false
	return p.base.next.RegisterUser(ctx, user, opts)
}

func (p *policyHandler) CreateUser(ctx context.Context, user *ent.User) error {
	isAdmin, err := p.isAdmin(ctx)
	if err != nil {
		return err
	}

	if !isAdmin {
		return errors.Forbidden
	}
	return p.base.CreateUser(ctx, user)
}

func (p *policyHandler) UserInfo(ctx context.Context, username string, opts pb.UserOptions) (*ent.User, error) {
	opts.WithPassword = false
	isAdmin, err := p.isAdmin(ctx)
	if err != nil {
		log.Error("could not determine if context is admin", log.Err(err))
		return nil, errors.Internal
	}

	if isAdmin {
		opts.WithAccessList = true
		opts.WithGroups = true
		opts.WithPermissions = true
	} else {
		cred := ome.CredentialsFromContext(ctx)
		if cred == nil {
			return nil, errors.Forbidden
		}

		opts.WithAccessList = cred.Username == username
		opts.WithGroups = cred.Username == username
		opts.WithPermissions = cred.Username == username
	}

	return p.base.UserInfo(ctx, username, opts)
}

func (p *policyHandler) ValidateUser(ctx context.Context, userId string, opts pb.UserOptions) error {
	isAdmin, err := p.isAdmin(ctx)
	if err != nil {
		return err
	}

	if !isAdmin {
		return errors.Forbidden
	}

	return p.base.ValidateUser(ctx, userId, opts)
}

func (p *policyHandler) ListUsers(ctx context.Context, opts pb.UserOptions) ([]*ent.User, error) {
	isAdmin, err := p.isAdmin(ctx)
	if err != nil {
		return nil, err
	}

	if !isAdmin {
		cred := ome.CredentialsFromContext(ctx)
		if cred == nil {
			return nil, errors.Forbidden
		}
	} else {
		opts.WithPermissions = true
		opts.WithGroups = true
		opts.WithAccessList = true
	}

	return p.base.ListUsers(ctx, opts)
}

func (p *policyHandler) PutData(ctx context.Context, data *pb.Data, opts pb.PutDataOptions) error {
	isAdmin, err := p.isAdmin(ctx)
	if err != nil {
		return err
	}

	if isAdmin {
		return errors.NotFound
	}

	err = p.assertIsAllowedOnData(&ctx, pb.Action_create, data.Collection, data.ID)
	if err != nil {
		return err
	}

	return p.base.PutData(ctx, data, opts)
}

func (p *policyHandler) GetData(ctx context.Context, collection string, id string, opts pb.GetDataOptions) (*pb.Data, error) {
	isAdmin, err := p.isAdmin(ctx)
	if err != nil {
		return nil, err
	}

	if !isAdmin {
		err := p.assertIsAllowedOnData(&ctx, pb.Action_read, collection, id)
		if err != nil {
			return nil, err
		}
	}
	return p.base.GetData(ctx, collection, id, opts)
}

func (p *policyHandler) Info(ctx context.Context, collection string, id string) (*pb.Info, error) {
	isAdmin, err := p.isAdmin(ctx)
	if err != nil {
		return nil, err
	}

	if isAdmin {
		return p.base.Info(ctx, collection, id)
	}

	err = p.assertIsAllowedOnData(&ctx, pb.Action_read, collection, id)
	if err != nil {
		return nil, err
	}
	return getDataInfo(ctx, collection, id), nil
}

func (p *policyHandler) Delete(ctx context.Context, collection string, id string) error {
	isAdmin, err := p.isAdmin(ctx)
	if err != nil {
		return err
	}

	if !isAdmin {
		err := p.assertIsAllowedOnData(&ctx, pb.Action_read, collection, id)
		if err != nil {
			return err
		}
	}
	return p.base.Delete(ctx, collection, id)
}

func (p *policyHandler) List(ctx context.Context, collection string, opts pb.ListOptions) (*pb.DataList, error) {
	isAdmin, err := p.isAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		opts.IDFilter = pb.IDFilterFunc(func(id string) (bool, error) {
			err = p.assertIsAllowedOnData(&ctx, pb.Action_read, collection, id)
			if err != nil {
				if err == errors.Unauthorized || err == errors.Forbidden {
					return false, nil
				}
				return false, err
			}
			return true, nil
		})
	}
	return p.base.List(ctx, collection, opts)
}

func (p *policyHandler) SaveGraft(ctx context.Context, graft *pb.Graft) (string, error) {
	isAdmin, err := p.isAdmin(ctx)
	if err != nil {
		return "", err
	}
	if !isAdmin {
		err = p.assertIsAllowedOnData(&ctx, pb.Action_graft, graft.Collection, graft.DataID)
		if err != nil {
			return "", err
		}
	}
	return p.next.SaveGraft(ctx, graft)
}

func (p *policyHandler) GetGraft(ctx context.Context, collection string, dataID string, id string) (*pb.Graft, error) {
	isAdmin, err := p.isAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		err = p.assertIsAllowedOnGraft(&ctx, pb.Action_read, collection, dataID, id)
		if err != nil {
			return nil, err
		}
	}
	return p.base.GetGraft(ctx, collection, dataID, id)
}

func (p *policyHandler) GraftInfo(ctx context.Context, collection string, dataID string, id string) (*pb.GraftInfo, error) {
	isAdmin, err := p.isAdmin(ctx)
	if err != nil {
		return nil, err
	}

	if isAdmin {
		return p.base.GraftInfo(ctx, collection, dataID, id)
	}

	err = p.assertIsAllowedOnGraft(&ctx, pb.Action_read, collection, dataID, id)
	if err != nil {
		return nil, err
	}
	return getGraftInfo(ctx, collection, dataID, id), nil
}

func (p *policyHandler) ListGrafts(ctx context.Context, collection string, dataID string, opts pb.ListOptions) (*pb.GraftList, error) {
	isAdmin, err := p.isAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		opts.IDFilter = pb.IDFilterFunc(func(id string) (bool, error) {
			err = p.assertIsAllowedOnGraft(&ctx, pb.Action_read, collection, dataID, id)
			if err != nil {
				if err == errors.Unauthorized || err == errors.Forbidden {
					return false, nil
				}
				return false, err
			}
			return true, nil
		})
	}
	return p.base.ListGrafts(ctx, collection, dataID, opts)
}

func (p *policyHandler) DeleteGraft(ctx context.Context, collection string, dataID string, id string) error {
	isAdmin, err := p.isAdmin(ctx)
	if err != nil {
		return err
	}
	if !isAdmin {
		err = p.assertIsAllowedOnGraft(&ctx, pb.Action_delete, collection, dataID, id)
		if err != nil {
			return err
		}
	}
	return p.base.DeleteGraft(ctx, collection, dataID, id)
}
