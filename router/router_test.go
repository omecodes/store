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
	"github.com/omecodes/omestore/acl"
	"github.com/omecodes/omestore/auth"
	"github.com/omecodes/omestore/oms"
	"github.com/omecodes/omestore/pb"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
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
	settingsDB oms.SettingsManager
	aclDB      acl.Store

	userA = &pb.Auth{
		Uid:    "a",
		Email:  "a@ome.ci",
		Worker: false,
		Scope:  []string{"profile", "email"},
		Group:  "users",
	}
	userB = &pb.Auth{
		Uid:    "b",
		Email:  "b@ome.ci",
		Worker: false,
		Scope:  []string{"profile", "email"},
		Group:  "users",
	}
	userC = &pb.Auth{
		Uid:    "c",
		Email:  "c@ome.ci",
		Worker: false,
		Scope:  []string{"profile", "email"},
		Group:  "users",
	}
	worker = &pb.Auth{
		Uid:    "worker",
		Email:  "",
		Worker: true,
	}
	admin = &pb.Auth{
		Uid:   "admin",
		Email: "admin@ome.ci",
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

type failureDummyAccessStore struct{}

func (f *failureDummyAccessStore) DeleteForPath(ctx context.Context, objectID string, path string) error {
	panic("implement me")
}

func (f *failureDummyAccessStore) SaveRules(ctx context.Context, objectID string, rules *pb.PathAccessRules) error {
	return errors.New("failure")
}

func (f *failureDummyAccessStore) GetRules(ctx context.Context, objectID string) (*pb.PathAccessRules, error) {
	return nil, errors.New("failure")
}

func (f *failureDummyAccessStore) GetForPath(ctx context.Context, objectID string, path string) (*pb.AccessRules, error) {
	return nil, errors.New("failure")
}

func (f *failureDummyAccessStore) Delete(ctx context.Context, objectID string) error {
	return errors.New("failure")
}

type failureDummyStorage struct{}

func (f *failureDummyStorage) Save(ctx context.Context, object *oms.Object) error {
	return errors.New("failure")
}

func (f *failureDummyStorage) Patch(ctx context.Context, patch *oms.Patch) error {
	return errors.New("failure")
}

func (f *failureDummyStorage) Delete(ctx context.Context, objectID string) error {
	return errors.New("failure")
}

func (f *failureDummyStorage) List(ctx context.Context, before int64, count int, filter oms.ObjectFilter) (*oms.ObjectList, error) {
	return nil, errors.New("failure")
}

func (f *failureDummyStorage) ListAt(ctx context.Context, path string, before int64, count int, filter oms.ObjectFilter) (*oms.ObjectList, error) {
	return nil, errors.New("failure")
}

func (f *failureDummyStorage) Get(ctx context.Context, objectID string) (*oms.Object, error) {
	return nil, errors.New("failure")
}

func (f *failureDummyStorage) GetAt(ctx context.Context, objectID string, path string) (*oms.Object, error) {
	return nil, errors.New("failure")
}

func (f *failureDummyStorage) Info(ctx context.Context, objectID string) (*pb.Header, error) {
	return nil, errors.New("failure")
}

func (f *failureDummyStorage) Clear() error {
	return errors.New("failure")
}

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
		workersDB, err = bome.NewJSONMap(db, testDialect, "test_users")
		So(err, ShouldBeNil)
	}

	if aclDB == nil {
		aclDB, err = acl.NewSQLStore(db, testDialect, "test_access")
		So(err, ShouldBeNil)
	}

	if settingsDB == nil {
		settingsDB, err = oms.NewSQLSettings(db, testDialect, "test_settings")
		So(err, ShouldBeNil)
	}

	if objectsDB == nil {
		objectsDB, err = oms.NewSQLObjects(db, testDialect, "test_objects")
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
		xSearchEnv, err = cel.NewEnv(
			cel.Declarations(
				decls.NewVar("o", decls.NewMapType(decls.String, decls.Dyn)),
			),
		)
		So(err, ShouldBeNil)
	}
}

func fullConfiguredContext() context.Context {
	initDBs()
	ctx := WithWorkers(workersDB)(context.Background())
	ctx = WithRouterProvider(ctx, ProviderFunc(func(ctx context.Context) Router {
		return DefaultRouter()
	}))
	ctx = WithSettings(settingsDB)(ctx)
	ctx = acl.ContextWithStore(ctx, aclDB)
	ctx = oms.ContextWithStore(ctx, objectsDB)
	ctx = WithCelSearchEnv(xSearchEnv)(ctx)
	ctx = WithCelPolicyEnv(xPolicyEnv)(ctx)
	return ctx
}

func contextWithFailureDummyStore() context.Context {
	initDBs()
	ctx := WithWorkers(workersDB)(context.Background())
	ctx = WithRouterProvider(ctx, ProviderFunc(func(ctx context.Context) Router {
		return DefaultRouter()
	}))
	ctx = acl.ContextWithStore(ctx, aclDB)
	ctx = oms.ContextWithStore(ctx, &failureDummyStorage{})
	ctx = WithSettings(settingsDB)(ctx)
	ctx = WithCelSearchEnv(xSearchEnv)(ctx)
	ctx = WithCelPolicyEnv(xPolicyEnv)(ctx)
	return ctx
}

func contextWithoutStore() context.Context {
	initDBs()
	ctx := WithWorkers(workersDB)(context.Background())
	ctx = WithRouterProvider(ctx, ProviderFunc(func(ctx context.Context) Router {
		return DefaultRouter()
	}))
	ctx = acl.ContextWithStore(ctx, aclDB)
	ctx = WithSettings(settingsDB)(ctx)
	//ctx = WithObjectsStore(objectsDB)(ctx)
	ctx = WithCelSearchEnv(xSearchEnv)(ctx)
	ctx = WithCelPolicyEnv(xPolicyEnv)(ctx)
	return ctx
}

func contextWithoutACLStore() context.Context {
	initDBs()
	ctx := WithWorkers(workersDB)(context.Background())
	ctx = WithRouterProvider(ctx, ProviderFunc(func(ctx context.Context) Router {
		return DefaultRouter()
	}))
	//ctx = WithAccessStore(aclDB)(ctx)
	ctx = WithSettings(settingsDB)(ctx)
	ctx = oms.ContextWithStore(ctx, objectsDB)
	ctx = WithCelSearchEnv(xSearchEnv)(ctx)
	ctx = WithCelPolicyEnv(xPolicyEnv)(ctx)
	return ctx
}

func contextFailureDummyACLStore() context.Context {
	initDBs()
	ctx := WithWorkers(workersDB)(context.Background())
	ctx = WithRouterProvider(ctx, ProviderFunc(func(ctx context.Context) Router {
		return DefaultRouter()
	}))
	ctx = acl.ContextWithStore(ctx, &failureDummyAccessStore{})
	ctx = WithSettings(settingsDB)(ctx)
	ctx = oms.ContextWithStore(ctx, objectsDB)
	ctx = WithCelSearchEnv(xSearchEnv)(ctx)
	ctx = WithCelPolicyEnv(xPolicyEnv)(ctx)
	return ctx
}

func contextWithoutSettings() context.Context {
	initDBs()
	ctx := WithWorkers(workersDB)(context.Background())
	ctx = WithRouterProvider(ctx, ProviderFunc(func(ctx context.Context) Router {
		return DefaultRouter()
	}))
	ctx = acl.ContextWithStore(ctx, aclDB)
	//ctx = WithSettings(settingsDB)(ctx)
	ctx = oms.ContextWithStore(ctx, objectsDB)
	ctx = WithCelSearchEnv(xSearchEnv)(ctx)
	ctx = WithCelPolicyEnv(xPolicyEnv)(ctx)
	return ctx
}

func Test_PutObject(t *testing.T) {
	Convey("Cannot put object without having set default settings", t, func() {
		route := getRoute()
		ctx := fullConfiguredContext()
		settings := Settings(ctx)

		o := new(oms.Object)
		o.SetHeader(&pb.Header{
			CreatedBy: "ome",
			Size:      int64(len(object1)),
		})
		o.SetContent(bytes.NewBufferString(object1))

		id, err := route.PutObject(ctx, nil, nil, oms.PutDataOptions{})
		So(err, ShouldEqual, errors.BadInput)
		So(id, ShouldEqual, "")

		o.SetSize(0)
		id, err = route.PutObject(ctx, o, nil, oms.PutDataOptions{})
		So(err, ShouldEqual, errors.BadInput)
		So(id, ShouldEqual, "")

		o.SetSize(1220)
		id, err = route.PutObject(ctx, o, nil, oms.PutDataOptions{})
		So(err, ShouldEqual, errors.BadInput)
		So(id, ShouldEqual, "")

		err = settings.Delete(oms.SettingsDataMaxSizePath)
		So(err, ShouldBeNil)

		o.SetSize(int64(len(object1)))
		id, err = route.PutObject(ctx, o, nil, oms.PutDataOptions{})
		So(err, ShouldEqual, errors.Internal)
		So(id, ShouldEqual, "")

		err = settings.Set(oms.SettingsDataMaxSizePath, "ekhfs")
		So(err, ShouldBeNil)

		id, err = route.PutObject(ctx, o, nil, oms.PutDataOptions{})
		So(err, ShouldEqual, errors.Internal)
		So(id, ShouldEqual, "")

		err = settings.Set(oms.SettingsDataMaxSizePath, "1024")
		So(err, ShouldBeNil)
	})
}

func Test_PutObject0(t *testing.T) {
	Convey("Cannot put object with wrong parameters", t, func() {

		route := getRoute()

		o := new(oms.Object)
		o.SetHeader(&pb.Header{
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
		route := getRoute()

		o := new(oms.Object)
		o.SetHeader(&pb.Header{
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
		route := getRoute()

		o := new(oms.Object)
		o.SetHeader(&pb.Header{
			CreatedBy: "ome",
			Size:      int64(len(object1)),
		})
		o.SetContent(bytes.NewBufferString(object1))

		userACtx := auth.Context(contextWithoutStore(), userA)
		id, err := route.PutObject(userACtx, o, nil, oms.PutDataOptions{})
		So(err, ShouldEqual, errors.Internal)
		So(id, ShouldEqual, "")

		userACtx = auth.Context(contextWithoutACLStore(), userA)
		id, err = route.PutObject(userACtx, o, nil, oms.PutDataOptions{})
		So(err, ShouldEqual, errors.Internal)
		So(id, ShouldEqual, "")
	})
}

func Test_PutObject4(t *testing.T) {
	Convey("Authenticated and validated user can put object in a well configured context", t, func() {
		route := getRoute()

		o := new(oms.Object)
		o.SetHeader(&pb.Header{
			CreatedBy: "ome",
			Size:      int64(len(object1)),
		})
		o.SetContent(bytes.NewBufferString(object1))
		userACtx := auth.Context(fullConfiguredContext(), userA)
		id, err := route.PutObject(userACtx, o, nil, oms.PutDataOptions{})
		So(err, ShouldBeNil)
		So(id, ShouldNotBeNil)

		o2 := new(oms.Object)
		o2.SetHeader(&pb.Header{
			CreatedBy: "ome",
			Size:      int64(len(object1)),
		})
		o2.SetContent(bytes.NewBufferString(object2))
		userBCtx := auth.Context(fullConfiguredContext(), userB)
		id, err = route.PutObject(userBCtx, o2, nil, oms.PutDataOptions{})
		So(err, ShouldBeNil)
		So(id, ShouldNotEqual, "")

		o3 := new(oms.Object)
		o3.SetHeader(&pb.Header{
			CreatedBy: "ome",
			Size:      int64(len(object1)),
		})
		o3.SetContent(bytes.NewBufferString(object2))
		security := &pb.PathAccessRules{AccessRules: map[string]*pb.AccessRules{
			"$": {
				Read:  []string{"auth.uid=='b'"},
				Write: []string{"auth.uid=='b'"},
			},
		}}
		id, err = route.PutObject(userBCtx, o3, security, oms.PutDataOptions{})
		So(err, ShouldBeNil)
		So(id, ShouldNotEqual, "")
	})
}

func Test_PutObject5(t *testing.T) {
	Convey("Cannot save object with broken access store", t, func() {
		route := getRoute()

		o := new(oms.Object)
		o.SetHeader(&pb.Header{
			CreatedBy: "ome",
			Size:      int64(len(object1)),
		})
		o.SetContent(bytes.NewBufferString(object1))

		userACtx := auth.Context(contextFailureDummyACLStore(), userA)
		id, err := route.PutObject(userACtx, o, nil, oms.PutDataOptions{})
		So(err, ShouldEqual, errors.Internal)
		So(id, ShouldEqual, "")
	})
}

func Test_ListObjects(t *testing.T) {
	Convey("Cannot list objects without storage in context", t, func() {
		route := getRoute()

		userACtx := auth.Context(contextWithoutStore(), userA)
		objects, err := route.ListObjects(userACtx, oms.ListOptions{})
		So(err, ShouldEqual, errors.Internal)
		So(objects, ShouldBeNil)
	})
}

func Test_ListObjects0(t *testing.T) {
	Convey("List user objects", t, func() {
		route := getRoute()

		userACtx := auth.Context(fullConfiguredContext(), userA)
		objects, err := route.ListObjects(userACtx, oms.ListOptions{})
		So(err, ShouldBeNil)
		userAObjects = objects.Objects

		userBCtx := auth.Context(fullConfiguredContext(), userB)
		objects, err = route.ListObjects(userBCtx, oms.ListOptions{})
		So(err, ShouldBeNil)
		userBObjects = objects.Objects
	})
}

func Test_GetObjectHeader(t *testing.T) {
	Convey("Get object header with wrong parameters", t, func() {
		route := getRoute()
		h, err := route.GetObjectHeader(fullConfiguredContext(), "")
		So(err, ShouldEqual, errors.BadInput)
		So(h, ShouldBeNil)
	})
}

func Test_GetObjectHeader0(t *testing.T) {
	Convey("Non admin user cannot read other user objects", t, func() {
		route := getRoute()
		userBCtx := auth.Context(fullConfiguredContext(), userB)
		id := userAObjects[0].ID()
		o, err := route.GetObjectHeader(userBCtx, id)
		So(err, ShouldEqual, errors.Unauthorized)
		So(o, ShouldBeNil)
	})
}

func Test_GetObjectHeader1(t *testing.T) {
	Convey("Cannot read object header without storage in context", t, func() {
		route := getRoute(SkipPoliciesCheck())
		userBCtx := auth.Context(contextWithoutStore(), userB)

		o, err := route.GetObjectHeader(userBCtx, "some-id")
		So(err, ShouldEqual, errors.Internal)
		So(o, ShouldBeNil)
	})
}

func Test_GetObject(t *testing.T) {
	Convey("Get object with wrong parameters", t, func() {
		route := getRoute()
		h, err := route.GetObject(fullConfiguredContext(), "", oms.GetObjectOptions{})
		So(err, ShouldEqual, errors.BadInput)
		So(h, ShouldBeNil)
	})
}

func Test_GetObject1(t *testing.T) {
	Convey("Cannot get object without storage", t, func() {
		route := getRoute(SkipPoliciesCheck())

		userBCtx := auth.Context(contextWithoutStore(), userB)
		o, err := route.GetObject(userBCtx, "some-id", oms.GetObjectOptions{})
		So(err, ShouldEqual, errors.Internal)
		So(o, ShouldBeNil)
	})
}

func Test_GetObject2(t *testing.T) {
	Convey("Admin can read other user objects", t, func() {
		route := getRoute()
		adminCtx := auth.Context(fullConfiguredContext(), admin)

		id := userAObjects[0].ID()
		h, err := route.GetObjectHeader(adminCtx, id)
		So(err, ShouldBeNil)
		So(h.Id, ShouldEqual, id)
	})
}

func Test_GetObject3(t *testing.T) {
	Convey("Non admin user cannot read other user objects", t, func() {
		route := getRoute()
		userBCtx := auth.Context(fullConfiguredContext(), userB)

		id := userAObjects[0].ID()
		o, err := route.GetObject(userBCtx, id, oms.GetObjectOptions{})
		So(err, ShouldEqual, errors.Unauthorized)
		So(o, ShouldBeNil)
	})
}

func Test_GetObject4(t *testing.T) {
	Convey("User reads his own objects", t, func() {
		route := getRoute()
		userBCtx := auth.Context(fullConfiguredContext(), userB)
		o, err := route.GetObject(userBCtx, userBObjects[0].ID(), oms.GetObjectOptions{})
		So(err, ShouldBeNil)
		So(o.ID(), ShouldEqual, userBObjects[0].ID())
	})
}

func Test_GetObject5(t *testing.T) {
	Convey("User reads his own objects: using Path", t, func() {
		route := getRoute(SkipPoliciesCheck())
		userBCtx := auth.Context(fullConfiguredContext(), userB)
		o, err := route.GetObject(userBCtx, userBObjects[0].ID(), oms.GetObjectOptions{Path: "$.private"})
		So(err, ShouldBeNil)
		So(o.ID(), ShouldEqual, userBObjects[0].ID())
		data, err := ioutil.ReadAll(o.GetContent())
		So(err, ShouldBeNil)
		So(string(data), ShouldBeIn, "true", "false")
	})
}

func Test_Patch(t *testing.T) {
	Convey("Patch object with wrong parameters", t, func() {
		p := oms.NewPatch("", "")
		ctx := context.Background()
		route := getRoute()
		err := route.PatchObject(ctx, p, oms.PatchOptions{})
		So(err, ShouldEqual, errors.BadInput)
	})
}

func Test_Patch0(t *testing.T) {
	Convey("Patch object without defined default settings", t, func() {
		route := getRoute()
		ctx := fullConfiguredContext()

		settings := Settings(ctx)

		err := settings.Delete(oms.SettingsDataMaxSizePath)
		So(err, ShouldBeNil)

		p := oms.NewPatch("id", "$.user.name")
		p.SetContent(bytes.NewBufferString("What are you doing?"))
		p.SetSize(42)

		err = route.PatchObject(ctx, p, oms.PatchOptions{})
		So(err, ShouldEqual, errors.Internal)

		err = settings.Set(oms.SettingsDataMaxSizePath, "1024")
		So(err, ShouldBeNil)

		p.SetSize(1025)
		err = route.PatchObject(ctx, p, oms.PatchOptions{})
		So(err, ShouldEqual, errors.BadInput)

		err = settings.Set(oms.SettingsDataMaxSizePath, "1024")
		So(err, ShouldBeNil)

		err = settings.Set(oms.SettingsDataMaxSizePath, "ekhfs")
		So(err, ShouldBeNil)

		p.SetSize(100)
		err = route.PatchObject(ctx, p, oms.PatchOptions{})
		So(err, ShouldEqual, errors.Internal)

		err = settings.Set(oms.SettingsDataMaxSizePath, "1024")
		So(err, ShouldBeNil)
	})
}

func Test_Patch1(t *testing.T) {
	Convey("Cannot patch other user object", t, func() {
		buf := bytes.NewBufferString("[\"" + userA.Uid + "\"]")
		p := oms.NewPatch(userAObjects[0].ID(), "$.followers")
		p.SetContent(buf)
		p.SetSize(int64(buf.Len()))

		userBCtx := auth.Context(fullConfiguredContext(), userB)
		route := getRoute()
		err := route.PatchObject(userBCtx, p, oms.PatchOptions{})
		So(err, ShouldEqual, errors.Unauthorized)
	})
}

func Test_Patch2(t *testing.T) {
	Convey("Cannot patch object without storage in context", t, func() {
		buf := bytes.NewBufferString("[\"" + userA.Uid + "\"]")
		p := oms.NewPatch(userBObjects[0].ID(), "$.followers")
		p.SetContent(buf)
		p.SetSize(int64(buf.Len()))

		userBCtx := auth.Context(contextWithoutStore(), userB)
		route := getRoute(SkipPoliciesCheck())
		err := route.PatchObject(userBCtx, p, oms.PatchOptions{})
		So(err, ShouldEqual, errors.Internal)
	})
}

func Test_Patch3(t *testing.T) {
	Convey("User can only patch object they own", t, func() {
		buf := bytes.NewBufferString("[\"" + userA.Uid + "\"]")
		p := oms.NewPatch(userBObjects[0].ID(), "$.followers")
		p.SetContent(buf)
		p.SetSize(int64(buf.Len()))

		userBCtx := auth.Context(fullConfiguredContext(), userB)
		route := getRoute()
		err := route.PatchObject(userBCtx, p, oms.PatchOptions{})
		So(err, ShouldBeNil)
	})
}

func Test_SearchObjects(t *testing.T) {
	Convey("Cannot perform search with empty expression parameters", t, func() {
		route := getRoute(SkipExec(), SkipPoliciesCheck())
		objects, err := route.SearchObjects(fullConfiguredContext(), oms.SearchParams{MatchedExpression: ""}, oms.SearchOptions{})
		So(err, ShouldEqual, errors.BadInput)
		So(objects, ShouldBeNil)

	})
}

func Test_SearchObjects0(t *testing.T) {
	Convey("Searching with 'false' expression returns empty object list", t, func() {
		route := getRoute(SkipExec())
		objects, err := route.SearchObjects(fullConfiguredContext(), oms.SearchParams{MatchedExpression: "false"}, oms.SearchOptions{})
		So(err, ShouldBeNil)
		So(objects.Objects, ShouldHaveLength, 0)
	})
}

func Test_SearchObjects1(t *testing.T) {
	Convey("Searching with 'true' expression returns the same result as list", t, func() {
		route := getRoute()
		userACtx := auth.Context(fullConfiguredContext(), userA)
		objects, err := route.SearchObjects(userACtx, oms.SearchParams{MatchedExpression: "true"}, oms.SearchOptions{})
		So(err, ShouldBeNil)
		So(objects.Objects, ShouldNotBeEmpty)
	})
}

func Test_SearchObjects2(t *testing.T) {
	Convey("User can search on objects he created", t, func() {
		route := getRoute()
		userACtx := auth.Context(fullConfiguredContext(), userA)
		objects, err := route.SearchObjects(userACtx, oms.SearchParams{MatchedExpression: "o.private"}, oms.SearchOptions{})
		So(err, ShouldBeNil)
		So(objects.Objects, ShouldNotBeEmpty)

		for _, o := range objects.Objects {
			data, err := ioutil.ReadAll(o.GetContent())
			So(err, ShouldBeNil)

			content := string(data)

			js, err := oms.JSONParse(content)
			So(err, ShouldBeNil)

			val, err := js.BoolAt("$.private")
			So(err, ShouldBeNil)
			So(val, ShouldBeTrue)
		}
	})
}

func Test_SearchObjects3(t *testing.T) {
	Convey("Cannot perform search on broken store", t, func() {
		route := getRoute()
		userACtx := auth.Context(contextWithFailureDummyStore(), userA)
		objects, err := route.SearchObjects(userACtx, oms.SearchParams{MatchedExpression: "o.private "}, oms.SearchOptions{})
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "failure")
		So(objects, ShouldBeNil)
	})
}

func Test_SearchObjects4(t *testing.T) {
	Convey("Cannot perform search on wrong syntax cel-expression", t, func() {
		route := getRoute()
		userACtx := auth.Context(contextWithFailureDummyStore(), userA)
		objects, err := route.SearchObjects(userACtx, oms.SearchParams{MatchedExpression: "o.private && a.public"}, oms.SearchOptions{})
		So(err, ShouldEqual, errors.BadInput)
		So(objects, ShouldBeNil)
	})
}

func Test_Delete(t *testing.T) {
	Convey("Cannot perform delete object with wrong parameters", t, func() {
		ctx := fullConfiguredContext()
		route := getRoute()
		err := route.DeleteObject(ctx, "")
		So(err, ShouldEqual, errors.BadInput)
	})
}

func Test_Delete0(t *testing.T) {
	Convey("Cannot delete other user object", t, func() {
		route := getRoute()
		userACtx := auth.Context(fullConfiguredContext(), userA)

		err := route.DeleteObject(userACtx, userBObjects[0].ID())
		So(err, ShouldEqual, errors.Unauthorized)
	})
}

func Test_Delete1(t *testing.T) {
	Convey("Cannot delete object without storage in context", t, func() {
		route := getRoute(SkipPoliciesCheck())
		userACtx := auth.Context(contextWithoutStore(), userA)

		err := route.DeleteObject(userACtx, userAObjects[0].ID())
		So(err, ShouldEqual, errors.Internal)
	})
}

func Test_Delete2(t *testing.T) {
	Convey("Cannot delete object without acl in context", t, func() {
		route := getRoute(SkipPoliciesCheck())
		userACtx := auth.Context(contextWithoutACLStore(), userA)

		err := route.DeleteObject(userACtx, userAObjects[0].ID())
		So(err, ShouldEqual, errors.Internal)
	})
}

func Test_Delete3(t *testing.T) {
	Convey("Clear delete their objects", t, func() {
		route := getRoute(SkipPoliciesCheck())

		adminCtx := auth.Context(fullConfiguredContext(), admin)
		objectList, err := route.ListObjects(adminCtx, oms.ListOptions{})
		So(err, ShouldBeNil)

		for _, o := range objectList.Objects {
			err = route.DeleteObject(adminCtx, o.ID())
			So(err, ShouldBeNil)
		}
	})
}

func Test_Delete4(t *testing.T) {
	Convey("Cannot delete object using broken store", t, func() {
		route := getRoute(SkipPoliciesCheck())
		userACtx := auth.Context(contextWithFailureDummyStore(), userA)
		err := route.DeleteObject(userACtx, "object")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "failure")
	})
}
