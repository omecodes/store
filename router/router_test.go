package router

import (
	"bytes"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/omecodes/bome"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/omestore/oms"
	"github.com/omecodes/omestore/pb"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"strings"
	"testing"
)

var (
	testDBUri       string
	jsonTestEnabled bool
	testDialect     string

	xPolicyEnv *cel.Env
	xSearchEnv *cel.Env
	objectsDB  oms.Objects
	workersDB  *bome.JSONMap
	settingsDB *bome.Map
	aclDB      oms.AccessStore

	userA = &pb.Auth{
		Uid:       "a",
		Email:     "a@ome.ci",
		Worker:    false,
		Validated: true,
		Scope:     []string{"profile", "email"},
		Group:     "users",
	}
	userB = &pb.Auth{
		Uid:       "b",
		Email:     "b@ome.ci",
		Worker:    false,
		Validated: true,
		Scope:     []string{"profile", "email"},
		Group:     "users",
	}
	userC = &pb.Auth{
		Uid:       "c",
		Email:     "c@ome.ci",
		Worker:    false,
		Validated: false,
		Scope:     []string{"profile", "email"},
		Group:     "users",
	}
	worker = &pb.Auth{
		Uid:       "worker",
		Email:     "",
		Worker:    true,
		Validated: true,
	}
	admin = &pb.Auth{
		Uid:       "admin",
		Email:     "admin@ome.ci",
		Validated: true,
	}

	userAObjects []*oms.Object
	userBObjects []*oms.Object

	object1 = `{
	"project": "ome",
	"private": true,
	"git": "https://github.com/omecodes/ome.git",
	"description": "Service Authority. Generates and signs certificates for services."
}`
	object2 = `{
	"project": "accounts",
	"private": true,
	"git": "https://github.com/omecodes/accounts.git",
	"description": "Account manager application. Supports OAUTH2"
}`
	object3 = `{
	"project": "tdb",
	"private": true,
	"git": "https://github.com/omecodes/tdb.git",
	"description": "Token store app"
}`
	object4 = `{
	"project": "libome",
	"private": true,
	"git": "https://github.com/omecodes/libome.git",
	"description": "Base library for all service definition"
}`
)

func init() {
	testDBUri = os.Getenv("OMS_TESTS_DB")
	if testDBUri == "" {
		testDBUri = "objects.db"
	}
	testDialect = os.Getenv("OMS_TESTS_DIALECT")
	if testDialect == "" {
		testDialect = bome.SQLite3
	}

	if flag.Lookup("test.v") != nil || strings.HasSuffix(os.Args[0], ".test") || strings.Contains(os.Args[0], "/_test/") {
		fmt.Println()
		fmt.Println()
		fmt.Println("TESTS_DIALECT: ", testDialect)
		fmt.Println("TESTS_DB     : ", testDBUri)
		fmt.Println("TESTS_ENABLED: ", jsonTestEnabled)
		fmt.Println()
		fmt.Println()
	}
}

func initDBs() {
	var db, err = sql.Open(testDialect, testDBUri)
	So(err, ShouldBeNil)

	if workersDB == nil {
		workersDB, err = bome.NewJSONMap(db, testDialect, "users")
		So(err, ShouldBeNil)
	}

	if aclDB == nil {
		aclDB, err = oms.NewSQLAccessStore(db, testDialect, "accesses")
		So(err, ShouldBeNil)
	}

	if settingsDB == nil {
		settingsDB, err = bome.NewMap(db, testDialect, "settings")
		So(err, ShouldBeNil)
	}

	if objectsDB == nil {
		objectsDB, err = oms.NewSQLObjects(db, testDialect)
		So(err, ShouldBeNil)
	}

	if xPolicyEnv == nil {
		xPolicyEnv, err = cel.NewEnv(
			cel.Declarations(
				decls.NewVar("auth", decls.NewMapType(decls.String, decls.Dyn)),
				decls.NewVar("data", decls.NewMapType(decls.String, decls.Dyn)),
			),
		)
		So(err, ShouldBeNil)
	}

	if xSearchEnv == nil {
		xSearchEnv, err = cel.NewEnv()
		So(err, ShouldBeNil)
	}
}

func fullConfiguredContext() context.Context {
	initDBs()
	ctx := WithWorkers(workersDB)(context.Background())
	ctx = WithAccessStore(aclDB)(ctx)
	ctx = WithSettings(settingsDB)(ctx)
	ctx = WithObjectsStore(objectsDB)(ctx)
	ctx = WithCelSearchEnv(xSearchEnv)(ctx)
	ctx = WithCelPolicyEnv(xPolicyEnv)(ctx)
	return ctx
}

func contextWithoutStore() context.Context {
	initDBs()
	ctx := WithWorkers(workersDB)(context.Background())
	ctx = WithAccessStore(aclDB)(ctx)
	ctx = WithSettings(settingsDB)(ctx)
	//ctx = WithObjectsStore(objectsDB)(ctx)
	ctx = WithCelSearchEnv(xSearchEnv)(ctx)
	ctx = WithCelPolicyEnv(xPolicyEnv)(ctx)
	return ctx
}

func contextWithoutACLStore() context.Context {
	initDBs()
	ctx := WithWorkers(workersDB)(context.Background())
	//ctx = WithAccessStore(aclDB)(ctx)
	ctx = WithSettings(settingsDB)(ctx)
	ctx = WithObjectsStore(objectsDB)(ctx)
	ctx = WithCelSearchEnv(xSearchEnv)(ctx)
	ctx = WithCelPolicyEnv(xPolicyEnv)(ctx)
	return ctx
}

func contextWithoutSettings() context.Context {
	initDBs()
	ctx := WithWorkers(workersDB)(context.Background())
	ctx = WithAccessStore(aclDB)(ctx)
	//ctx = WithSettings(settingsDB)(ctx)
	ctx = WithObjectsStore(objectsDB)(ctx)
	ctx = WithCelSearchEnv(xSearchEnv)(ctx)
	ctx = WithCelPolicyEnv(xPolicyEnv)(ctx)
	return ctx
}

func Test_SetSettings(t *testing.T) {
	Convey("Set settings with wrong parameters", t, func() {
		ctx := fullConfiguredContext()
		route := Route()

		userACtx := WithUserInfo(ctx, userA)
		err := route.SetSettings(userACtx, "", "1024", oms.SettingsOptions{})
		So(err, ShouldEqual, errors.BadInput)

		err = route.SetSettings(userACtx, "something", "", oms.SettingsOptions{})
		So(err, ShouldEqual, errors.BadInput)
	})
}

func Test_SetSettings1(t *testing.T) {
	Convey("Non admin cannot set/get settings", t, func() {
		ctx := fullConfiguredContext()
		route := Route()

		userACtx := WithUserInfo(ctx, userA)
		err := route.SetSettings(userACtx, oms.SettingsDataMaxSizePath, "1024", oms.SettingsOptions{})
		So(err, ShouldEqual, errors.Forbidden)
	})
}

func Test_SetSettings2(t *testing.T) {
	Convey("Writing settings always requires settings DB in context", t, func() {
		route := Route()
		adminCtxWoSettings := WithUserInfo(contextWithoutSettings(), admin)

		err := route.SetSettings(adminCtxWoSettings, oms.SettingsDataMaxSizePath, "1024", oms.SettingsOptions{})
		So(err, ShouldEqual, errors.Internal)

		value, err := route.GetSettings(adminCtxWoSettings, oms.SettingsDataMaxSizePath)
		So(err, ShouldEqual, errors.Internal)
		So(value, ShouldEqual, "")

	})
}

func Test_SetSettings3(t *testing.T) {
	Convey("Only Admin can set settings with properly configured context", t, func() {
		route := Route()

		adminCtx := WithUserInfo(fullConfiguredContext(), admin)

		err := route.SetSettings(adminCtx, oms.SettingsDataMaxSizePath, "1024", oms.SettingsOptions{})
		So(err, ShouldBeNil)

		err = route.SetSettings(adminCtx, oms.SettingsCreateDataSecurityRule, "auth.worker || auth.validated", oms.SettingsOptions{})
		So(err, ShouldBeNil)
	})
}

func Test_GetSetting(t *testing.T) {
	Convey("Get settings with wrong parameters", t, func() {
		route := Route()
		userACtx := WithUserInfo(fullConfiguredContext(), userA)

		value, err := route.GetSettings(userACtx, "")
		So(err, ShouldEqual, errors.BadInput)
		So(value, ShouldEqual, "")
	})
}

func Test_GetSetting1(t *testing.T) {
	Convey("Non admin cannot read settings", t, func() {
		route := Route()
		userACtx := WithUserInfo(fullConfiguredContext(), userA)

		value, err := route.GetSettings(userACtx, oms.SettingsDataMaxSizePath)
		So(err, ShouldEqual, errors.Forbidden)
		So(value, ShouldEqual, "")

		value, err = route.GetSettings(userACtx, oms.SettingsCreateDataSecurityRule)
		So(err, ShouldEqual, errors.Forbidden)
		So(value, ShouldEqual, "")
	})
}

func Test_GetSetting2(t *testing.T) {
	Convey("Admin is allowed to get settings", t, func() {

		ctx := fullConfiguredContext()
		route := Route()

		adminCtx := WithUserInfo(ctx, admin)

		value, err := route.GetSettings(adminCtx, oms.SettingsDataMaxSizePath)
		So(err, ShouldBeNil)
		So(value, ShouldEqual, "1024")

		value, err = route.GetSettings(adminCtx, oms.SettingsCreateDataSecurityRule)
		So(err, ShouldBeNil)
		So(value, ShouldEqual, "auth.worker || auth.validated")
	})
}

func Test_DeleteSettings(t *testing.T) {
	Convey("Delete settings with wrong parameters", t, func() {
		route := Route()
		err := route.DeleteSettings(context.Background(), "")
		So(err, ShouldEqual, errors.BadInput)
	})
}

func Test_DeleteSettings0(t *testing.T) {
	Convey("Non authenticated context cannot delete settings", t, func() {
		route := Route()
		err := route.DeleteSettings(context.Background(), "something")
		So(err, ShouldEqual, errors.Forbidden)
	})
}

func Test_DeleteSettings1(t *testing.T) {
	Convey("Delete settings with incomplete context", t, func() {
		route := Route()
		adminCtx := WithUserInfo(contextWithoutSettings(), admin)
		err := route.DeleteSettings(adminCtx, "hello")
		So(err, ShouldEqual, errors.Internal)
	})
}

func Test_DeleteSettings2(t *testing.T) {
	Convey("Delete settings with wrong parameters", t, func() {
		route := Route()
		adminCtx := WithUserInfo(fullConfiguredContext(), admin)
		err := route.DeleteSettings(adminCtx, "hello")
		So(err, ShouldBeNil)
	})
}

func Test_ClearSettings(t *testing.T) {
	Convey("Non authenticated user cannot clear settings", t, func() {
		route := Route()
		err := route.ClearSettings(context.Background())
		So(err, ShouldEqual, errors.Forbidden)
	})
}

func Test_ClearSettings0(t *testing.T) {
	Convey("Non admin user cannot clear settings", t, func() {
		route := Route()
		userACtx := WithUserInfo(fullConfiguredContext(), userA)
		err := route.ClearSettings(userACtx)
		So(err, ShouldEqual, errors.Forbidden)
	})
}

func Test_ClearSettings1(t *testing.T) {
	Convey("Non admin user cannot clear settings", t, func() {
		route := Route()
		userACtx := WithUserInfo(fullConfiguredContext(), userA)
		err := route.ClearSettings(userACtx)
		So(err, ShouldEqual, errors.Forbidden)
	})
}

func Test_ClearSettings2(t *testing.T) {
	Convey("Admin user can clear settings", t, func() {
		route := Route()
		adminCtx := WithUserInfo(fullConfiguredContext(), admin)

		err := route.ClearSettings(adminCtx)
		So(err, ShouldBeNil)

		err = route.SetSettings(adminCtx, oms.SettingsDataMaxSizePath, "1024", oms.SettingsOptions{})
		So(err, ShouldBeNil)

		err = route.SetSettings(adminCtx, oms.SettingsCreateDataSecurityRule, "auth.worker || auth.validated", oms.SettingsOptions{})
		So(err, ShouldBeNil)
	})
}

func Test_PutObject(t *testing.T) {
	Convey("Cannot put object without having set default settings", t, func() {
		route := Route()

		o := new(oms.Object)
		o.SetHeader(&pb.Header{
			Id:        "ome-ca",
			CreatedBy: "ome",
			Size:      int64(len(object1)),
		})
		o.SetContent(bytes.NewBufferString(object1))

		id, err := route.PutObject(fullConfiguredContext(), nil, nil, oms.PutDataOptions{})
		So(err, ShouldEqual, errors.BadInput)
		So(id, ShouldEqual, "")

		o.SetSize(0)
		id, err = route.PutObject(fullConfiguredContext(), o, nil, oms.PutDataOptions{})
		So(err, ShouldEqual, errors.BadInput)
		So(id, ShouldEqual, "")

		o.SetSize(1220)
		id, err = route.PutObject(fullConfiguredContext(), o, nil, oms.PutDataOptions{})
		So(err, ShouldEqual, errors.BadInput)
		So(id, ShouldEqual, "")

		adminCtx := WithUserInfo(fullConfiguredContext(), admin)
		err = route.DeleteSettings(adminCtx, oms.SettingsDataMaxSizePath)
		So(err, ShouldBeNil)

		o.SetSize(int64(len(object1)))
		id, err = route.PutObject(fullConfiguredContext(), o, nil, oms.PutDataOptions{})
		So(err, ShouldEqual, errors.Internal)
		So(id, ShouldEqual, "")

		err = route.SetSettings(adminCtx, oms.SettingsDataMaxSizePath, "ekhfs", oms.SettingsOptions{})
		So(err, ShouldBeNil)

		id, err = route.PutObject(fullConfiguredContext(), o, nil, oms.PutDataOptions{})
		So(err, ShouldEqual, errors.Internal)
		So(id, ShouldEqual, "")

		err = route.SetSettings(adminCtx, oms.SettingsDataMaxSizePath, "1024", oms.SettingsOptions{})
		So(err, ShouldBeNil)
	})
}

func Test_PutObject0(t *testing.T) {
	Convey("Cannot put object with wrong parameters", t, func() {

		route := Route()

		o := new(oms.Object)
		o.SetHeader(&pb.Header{
			Id:        "ome-ca",
			CreatedBy: "ome",
			Size:      int64(len(object1)),
		})
		o.SetContent(bytes.NewBufferString(object1))

		id, err := route.PutObject(fullConfiguredContext(), nil, nil, oms.PutDataOptions{})
		So(err, ShouldEqual, errors.BadInput)
		So(id, ShouldEqual, "")

		o.SetSize(0)
		id, err = route.PutObject(fullConfiguredContext(), o, nil, oms.PutDataOptions{})
		So(err, ShouldEqual, errors.BadInput)
		So(id, ShouldEqual, "")

	})
}

func Test_PutObject1(t *testing.T) {
	Convey("Non authenticated user cannot put object", t, func() {
		route := Route()

		o := new(oms.Object)
		o.SetHeader(&pb.Header{
			Id:        "ome-ca",
			CreatedBy: "ome",
			Size:      int64(len(object1)),
		})
		o.SetContent(bytes.NewBufferString(object1))

		id, err := route.PutObject(fullConfiguredContext(), o, nil, oms.PutDataOptions{})
		So(err, ShouldEqual, errors.Forbidden)
		So(id, ShouldEqual, "")
	})
}

func Test_PutObject2(t *testing.T) {
	Convey("Non complete context cannot put object", t, func() {
		route := Route()

		o := new(oms.Object)
		o.SetHeader(&pb.Header{
			Id:        "ome-ca",
			CreatedBy: "ome",
			Size:      int64(len(object1)),
		})
		o.SetContent(bytes.NewBufferString(object1))

		userACtx := WithUserInfo(contextWithoutStore(), userA)
		id, err := route.PutObject(userACtx, o, nil, oms.PutDataOptions{})
		So(err, ShouldEqual, errors.Internal)
		So(id, ShouldEqual, "")

		userACtx = WithUserInfo(contextWithoutACLStore(), userA)
		id, err = route.PutObject(userACtx, o, nil, oms.PutDataOptions{})
		So(err, ShouldEqual, errors.Internal)
		So(id, ShouldEqual, "")
	})
}

func Test_PutObject3(t *testing.T) {
	Convey("Non validated user cannot put objects", t, func() {
		route := Route()

		o := new(oms.Object)
		o.SetHeader(&pb.Header{
			Id:        "ome-ca",
			CreatedBy: "ome",
			Size:      int64(len(object1)),
		})
		o.SetContent(bytes.NewBufferString(object1))

		userCCtx := WithUserInfo(contextWithoutStore(), userC)
		id, err := route.PutObject(userCCtx, o, nil, oms.PutDataOptions{})
		So(err, ShouldEqual, errors.Unauthorized)
		So(id, ShouldEqual, "")
	})
}

func Test_PutObject4(t *testing.T) {
	Convey("Authenticated and validated user can put object in a well configured context", t, func() {
		route := Route()

		o := new(oms.Object)
		o.SetHeader(&pb.Header{
			Id:        "ome-ca",
			CreatedBy: "ome",
			Size:      int64(len(object1)),
		})
		o.SetContent(bytes.NewBufferString(object1))
		userACtx := WithUserInfo(fullConfiguredContext(), userA)
		id, err := route.PutObject(userACtx, o, nil, oms.PutDataOptions{})
		So(err, ShouldBeNil)
		So(id, ShouldNotBeNil)

		o2 := new(oms.Object)
		o2.SetHeader(&pb.Header{
			Id:        "ome-ca",
			CreatedBy: "ome",
			Size:      int64(len(object1)),
		})
		o2.SetContent(bytes.NewBufferString(object2))
		userBCtx := WithUserInfo(fullConfiguredContext(), userB)
		id, err = route.PutObject(userBCtx, o2, nil, oms.PutDataOptions{})
		So(err, ShouldBeNil)
		So(id, ShouldNotEqual, "")

		o3 := new(oms.Object)
		o3.SetHeader(&pb.Header{
			Id:        "ome-ca",
			CreatedBy: "ome",
			Size:      int64(len(object1)),
		})
		o3.SetContent(bytes.NewBufferString(object2))
		security := &pb.PathAccessRules{AccessRules: map[string]*pb.AccessRules{
			"$": {
				Read:  []string{"auth.validated && auth.uid=='b'"},
				Write: []string{"auth.validated && auth.uid=='b'"},
			},
		}}
		id, err = route.PutObject(userBCtx, o3, security, oms.PutDataOptions{})
		So(err, ShouldBeNil)
		So(id, ShouldNotEqual, "")
	})
}

func Test_ListObjects(t *testing.T) {
	Convey("List user objects", t, func() {
		route := Route()

		userACtx := WithUserInfo(fullConfiguredContext(), userA)
		objects, err := route.ListObjects(userACtx, oms.ListOptions{})
		So(err, ShouldBeNil)
		userAObjects = objects.Objects

		userBCtx := WithUserInfo(fullConfiguredContext(), userB)
		objects, err = route.ListObjects(userBCtx, oms.ListOptions{})
		So(err, ShouldBeNil)
		userBObjects = objects.Objects
	})
}

func Test_GetObject(t *testing.T) {
	Convey("Non admin user cannot read other user objects", t, func() {
		route := Route()
		userBCtx := WithUserInfo(fullConfiguredContext(), userB)

		id := userAObjects[0].ID()
		o, err := route.GetObject(userBCtx, id, oms.GetDataOptions{})
		So(err, ShouldEqual, errors.Unauthorized)
		So(o, ShouldBeNil)

		h, err := route.GetObjectHeader(userBCtx, "")
		So(err, ShouldEqual, errors.BadInput)
		So(h, ShouldBeNil)

		h, err = route.GetObjectHeader(userBCtx, id)
		So(err, ShouldEqual, errors.Unauthorized)
		So(h, ShouldBeNil)
	})
}

func Test_GetObject1(t *testing.T) {
	Convey("Getting object without specifying ID", t, func() {
		route := Route()
		userBCtx := WithUserInfo(fullConfiguredContext(), userB)
		o, err := route.GetObject(userBCtx, "", oms.GetDataOptions{})
		So(err, ShouldEqual, errors.BadInput)
		So(o, ShouldBeNil)
	})
}

func Test_GetObject2(t *testing.T) {
	Convey("Admin can read other user objects", t, func() {
		route := Route()
		adminCtx := WithUserInfo(fullConfiguredContext(), admin)

		id := userAObjects[0].ID()
		h, err := route.GetObjectHeader(adminCtx, id)
		So(err, ShouldBeNil)
		So(h.Id, ShouldEqual, id)
	})
}

func Test_GetObject3(t *testing.T) {
	Convey("User reads his own objects", t, func() {
		route := Route()
		userBCtx := WithUserInfo(fullConfiguredContext(), userB)
		o, err := route.GetObject(userBCtx, userBObjects[0].ID(), oms.GetDataOptions{})
		So(err, ShouldBeNil)
		So(o.ID(), ShouldEqual, userBObjects[0].ID())
	})
}

func Test_Patch(t *testing.T) {
	Convey("Patch object with wrong parameters", t, func() {
		p := oms.NewPatch("", "")
		ctx := context.Background()
		route := Route()
		err := route.PatchObject(ctx, p, oms.PatchOptions{})
		So(err, ShouldEqual, errors.BadInput)
	})
}

func Test_Patch0(t *testing.T) {
	Convey("Patch object without defined default settings", t, func() {
		route := Route()

		adminCtx := WithUserInfo(fullConfiguredContext(), admin)
		err := route.DeleteSettings(adminCtx, oms.SettingsDataMaxSizePath)
		So(err, ShouldBeNil)

		p := oms.NewPatch("id", "$.user.name")
		p.SetContent(bytes.NewBufferString("What are you doing?"))
		p.SetSize(42)

		err = route.PatchObject(fullConfiguredContext(), p, oms.PatchOptions{})
		So(err, ShouldEqual, errors.Internal)

		err = route.SetSettings(adminCtx, oms.SettingsDataMaxSizePath, "1024", oms.SettingsOptions{})
		So(err, ShouldBeNil)

		p.SetSize(1025)
		err = route.PatchObject(fullConfiguredContext(), p, oms.PatchOptions{})
		So(err, ShouldEqual, errors.BadInput)

		err = route.SetSettings(adminCtx, oms.SettingsDataMaxSizePath, "1024", oms.SettingsOptions{})
		So(err, ShouldBeNil)

		err = route.SetSettings(adminCtx, oms.SettingsDataMaxSizePath, "ekhfs", oms.SettingsOptions{})
		So(err, ShouldBeNil)

		p.SetSize(100)
		err = route.PatchObject(fullConfiguredContext(), p, oms.PatchOptions{})
		So(err, ShouldEqual, errors.Internal)

		err = route.SetSettings(adminCtx, oms.SettingsDataMaxSizePath, "1024", oms.SettingsOptions{})
		So(err, ShouldBeNil)
	})
}

func Test_Patch1(t *testing.T) {
	Convey("Cannot patch other user object", t, func() {
		buf := bytes.NewBufferString("[\"" + userA.Uid + "\"]")
		p := oms.NewPatch(userAObjects[0].ID(), "$.followers")
		p.SetContent(buf)
		p.SetSize(int64(buf.Len()))

		userBCtx := WithUserInfo(fullConfiguredContext(), userB)
		route := Route()
		err := route.PatchObject(userBCtx, p, oms.PatchOptions{})
		So(err, ShouldEqual, errors.Unauthorized)
	})
}

func Test_Patch2(t *testing.T) {
	Convey("Cannot patch other user object", t, func() {
		buf := bytes.NewBufferString("[\"" + userA.Uid + "\"]")
		p := oms.NewPatch(userBObjects[0].ID(), "$.followers")
		p.SetContent(buf)
		p.SetSize(int64(buf.Len()))

		userBCtx := WithUserInfo(fullConfiguredContext(), userB)
		route := Route()
		err := route.PatchObject(userBCtx, p, oms.PatchOptions{})
		So(err, ShouldBeNil)
	})
}

func Test_Delete(t *testing.T) {
	Convey("Delete all items", t, func() {
		ctx := fullConfiguredContext()
		route := Route()

		err := route.DeleteObject(ctx, "")
		So(err, ShouldEqual, errors.BadInput)

		objectList, err := route.ListObjects(ctx, oms.ListOptions{})
		So(err, ShouldBeNil)
		So(objectList.Count, ShouldEqual, 0)

		userACtx := WithUserInfo(fullConfiguredContext(), userA)
		objectList, err = route.ListObjects(userACtx, oms.ListOptions{})
		So(err, ShouldBeNil)

		for _, o := range objectList.Objects {
			err = route.DeleteObject(userACtx, o.ID())
			So(err, ShouldBeNil)
		}

		userBCtx := WithUserInfo(fullConfiguredContext(), userB)
		objectList, err = route.ListObjects(userBCtx, oms.ListOptions{})
		So(err, ShouldBeNil)

		for _, o := range objectList.Objects {
			err = route.DeleteObject(userBCtx, o.ID())
			So(err, ShouldBeNil)
		}
	})
}
