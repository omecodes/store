package objects

import (
	"context"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/omecodes/bome"
	"github.com/omecodes/store/acl"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/common/utime"
	pb "github.com/omecodes/store/gen/go/proto"
	"github.com/omecodes/store/settings"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"os"
	"testing"
	"time"
)

var (
	db DB

	sm settings.Manager

	nsConfigStore acl.NamespaceConfigStore
	tupleStore    acl.TupleStore
	adminApp      *pb.ClientApp
	clientApp     *pb.ClientApp

	juveTeam = &pb.Collection{
		Id:          "juventus",
		Label:       "Team of Juventus",
		Description: "List of Juventus FC players",
		AclConfig: &pb.ACLConfig{
			Namespace:           "object",
			RelationWithCreated: "owner",
		},
		ActionAuthorizedUsers: &pb.PathAccessRules{
			AccessRules: map[string]*pb.ObjectActionsUsers{
				"$": {
					View:   &pb.SubjectSet{Relation: "viewer"},
					Edit:   &pb.SubjectSet{Relation: "editor"},
					Delete: &pb.SubjectSet{Relation: "owner"},
				},
			},
		},
	}
	barcaTeam = &pb.Collection{
		Id:          "barcelona",
		Label:       "Team of Barcelona",
		Description: "List of Barcelona FC players",
		AclConfig: &pb.ACLConfig{
			Namespace:           "object",
			RelationWithCreated: "owner",
		},
		ActionAuthorizedUsers: &pb.PathAccessRules{
			AccessRules: map[string]*pb.ObjectActionsUsers{
				"$": {
					View:   &pb.SubjectSet{Relation: "viewer"},
					Edit:   &pb.SubjectSet{Relation: "editor"},
					Delete: &pb.SubjectSet{Relation: "owner"},
				},
			},
		},
	}
	psgTeam = &pb.Collection{
		Id:          "paris-sg",
		Label:       "Team of PSG",
		Description: "List of Paris Saint-Germain players",
		AclConfig: &pb.ACLConfig{
			Namespace:           "object",
			RelationWithCreated: "owner",
		},
		ActionAuthorizedUsers: &pb.PathAccessRules{
			AccessRules: map[string]*pb.ObjectActionsUsers{
				"$": {
					View:   &pb.SubjectSet{Relation: "viewer"},
					Edit:   &pb.SubjectSet{Relation: "editor"},
					Delete: &pb.SubjectSet{Relation: "owner"},
				},
			},
		},
	}
)

func setupDatabases() {

	if db == nil {
		conn, err := sql.Open("sqlite3", ":memory:")
		So(err, ShouldBeNil)

		db, err = NewSqlDB(conn, bome.SQLite3, "store")
		So(err, ShouldBeNil)
	}

	if sm == nil {
		conn, err := sql.Open("sqlite3", ":memory:")
		So(err, ShouldBeNil)

		sm, err = settings.NewSQLManager(conn, bome.SQLite3, "settings")
		So(err, ShouldBeNil)
	}

	if nsConfigStore == nil {
		conn, err := sql.Open(bome.SQLite3, ":memory:")
		So(err, ShouldBeNil)

		nsConfigStore, err = acl.NewNamespaceSQLStore(conn, bome.SQLite3, "nc")
		So(err, ShouldBeNil)
	}

	if tupleStore == nil {
		conn, err := sql.Open(bome.SQLite3, ":memory:")
		So(err, ShouldBeNil)

		tupleStore, err = acl.NewTupleSQLStore(conn, bome.SQLite3, "ts")
		So(err, ShouldBeNil)

		setupNamespaceConfigs()
	}
}

func setupNamespaceConfigs() {
	groupNamespaceConfig := &pb.NamespaceConfig{
		Sid:       1,
		Namespace: "group",
		Relations: map[string]*pb.RelationDefinition{
			"owner": {
				Name: "owner",
				SubjectSetRewrite: []*pb.SubjectSetDefinition{
					{
						Type: pb.SubjectSetType_This,
					},
				},
			},
			"member": {
				Name: "member",
				SubjectSetRewrite: []*pb.SubjectSetDefinition{
					{
						Type: pb.SubjectSetType_This,
					},
					{
						Type:  pb.SubjectSetType_Computed,
						Value: "owner",
					},
				},
			},
		},
	}
	err := nsConfigStore.SaveNamespace(groupNamespaceConfig)
	So(err, ShouldBeNil)

	objectNamespaceConfig := &pb.NamespaceConfig{
		Sid:       1,
		Namespace: "object",
		Relations: map[string]*pb.RelationDefinition{
			"owner": {
				Name: "owner",
				SubjectSetRewrite: []*pb.SubjectSetDefinition{
					{
						Type: pb.SubjectSetType_This,
					},
				},
			},
			"editor": {
				Name: "editor",
				SubjectSetRewrite: []*pb.SubjectSetDefinition{
					{
						Type: pb.SubjectSetType_This,
					},
					{
						Type:  pb.SubjectSetType_Computed,
						Value: "owner",
					},
				},
			},
			"viewer": {
				Name: "viewer",
				SubjectSetRewrite: []*pb.SubjectSetDefinition{
					{
						Type: pb.SubjectSetType_This,
					},
					{
						Type:  pb.SubjectSetType_Computed,
						Value: "editor",
					},
					{
						Type:  pb.SubjectSetType_Computed,
						Value: "owner",
					},
				},
			},
		},
	}
	err = nsConfigStore.SaveNamespace(objectNamespaceConfig)
	So(err, ShouldBeNil)

	err = tupleStore.Save(context.Background(), &pb.DBEntry{
		Sid:      1,
		Object:   "group:admins",
		Relation: "member",
		Subject:  "admin",
	})
	So(err, ShouldBeNil)
}

func setupClientApps() {
	adminApp = &pb.ClientApp{
		Key:      "admin-app",
		Secret:   "secret",
		Type:     pb.ClientType_desktop,
		AdminApp: true,
	}
	clientApp = &pb.ClientApp{
		Key:      "client-app",
		Secret:   "secret",
		Type:     pb.ClientType_web,
		AdminApp: false,
	}
}

func setup() {
	setupDatabases()
	setupClientApps()
}

func tearDown() {
	_ = os.Remove("objects.db")
	_ = os.Remove("settings.db")
	_ = os.Remove("namespaces.db")
	_ = os.Remove("tuples.db")
}

func baseContext() context.Context {
	ctx := context.Background()

	ctx = acl.ContextWithManager(ctx, &acl.DefaultManager{})
	ctx = acl.ContextWithNamespaceConfigStore(ctx, nsConfigStore)
	ctx = acl.ContextWithTupleStore(ctx, tupleStore)

	ctx = settings.ContextWithManager(ctx, sm)

	ctx = ContextWithStore(ctx, db)
	return ctx
}

func missingSettingsContext() context.Context {
	ctx := context.Background()

	ctx = acl.ContextWithManager(ctx, &acl.DefaultManager{})
	ctx = acl.ContextWithNamespaceConfigStore(ctx, nsConfigStore)
	ctx = acl.ContextWithTupleStore(ctx, tupleStore)

	ctx = acl.ContextWithManager(ctx, &acl.DefaultManager{})
	ctx = ContextWithStore(ctx, db)
	return ctx
}

func userContext(ctx context.Context, name string) context.Context {
	return auth.ContextWithUser(ctx, &pb.User{Name: name})
}

// this create a context as if it was created by the authentication interceptor when
// receiving a request from a user by the means of a registered app client
func userContextFromRegisteredApplication(ctx context.Context, name string) context.Context {
	return userContext(clientAppContext(ctx), name)
}

// this create a context as if it was created by the authentication interceptor when
// receiving a request from a user by the means of a registered app client
func fullAdminContext() context.Context {
	return userContext(adminAppContext(baseContext()), "admin")
}

func clientAppContext(ctx context.Context) context.Context {
	return auth.ContextWithApp(ctx, clientApp)
}

func adminAppContext(ctx context.Context) context.Context {
	return auth.ContextWithApp(ctx, adminApp)
}

func Test_DBInitialization(t *testing.T) {
	Convey("Database initialization should be executed with no errors", t, func() {
		tearDown()
		setup()
	})
}

func TestHandler_CreateCollection1(t *testing.T) {
	Convey("COLLECTION - CREATE: cannot create a collection if id or security rules are not provided", t, func() {
		setup()

		router := DefaultRouter()
		h := router.GetHandler()

		col := &pb.Collection{
			Id:          "",
			Label:       "Objects",
			Description: "List of random object created for the sake of test",
			NumberIndex: nil,
			TextIndexes: nil,
			FieldsIndex: nil,
			ActionAuthorizedUsers: &pb.PathAccessRules{
				AccessRules: map[string]*pb.ObjectActionsUsers{
					"$": {
						View:   &pb.SubjectSet{},
						Edit:   &pb.SubjectSet{},
						Delete: &pb.SubjectSet{},
					},
				},
			},
		}

		// Try to create collection as admin
		adminContext := userContext(baseContext(), "admin")
		err := h.CreateCollection(adminContext, col, CreateCollectionOptions{})
		So(err, ShouldNotBeNil)

		col.Id = "objects"
		//col.ActionAccessSecurityRules = nil
		err = h.CreateCollection(adminContext, col, CreateCollectionOptions{})
		So(err, ShouldNotBeNil)

		// Try to create collection as unauthenticated user
		err = h.CreateCollection(baseContext(), col, CreateCollectionOptions{})
		So(err, ShouldNotBeNil)

		// Try to create collection as user1
		user1Context := userContext(baseContext(), "user1")
		err = h.CreateCollection(user1Context, col, CreateCollectionOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_CreateCollection2(t *testing.T) {
	Convey("COLLECTION - CREATE: cannot create a collection if user is not admin", t, func() {
		setup()
		router := DefaultRouter()
		h := router.GetHandler()

		// Try to create collection as admin
		adminContext := userContext(baseContext(), "user1")
		err := h.CreateCollection(adminContext, juveTeam, CreateCollectionOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_CreateCollection(t *testing.T) {
	Convey("COLLECTION - CREATE: can create a collection if user is admin", t, func() {
		setup()

		h := DefaultRouter().GetHandler()
		ctx := adminAppContext(baseContext())

		// Try to create collection as admin
		adminContext := userContext(ctx, "admin")
		err := h.CreateCollection(adminContext, juveTeam, CreateCollectionOptions{})
		So(err, ShouldBeNil)
		err = h.CreateCollection(adminContext, barcaTeam, CreateCollectionOptions{})
		So(err, ShouldBeNil)
		err = h.CreateCollection(adminContext, psgTeam, CreateCollectionOptions{})
		So(err, ShouldBeNil)

		// Retrieve new created collection from admin context
		col, err := h.GetCollection(adminContext, "juventus", GetCollectionOptions{})
		So(err, ShouldBeNil)
		So(col.Id, ShouldEqual, "juventus")
	})
}

func TestHandler_GetCollection1(t *testing.T) {
	Convey("COLLECTION - GET: cannot get a collection info if id is not provided", t, func() {
		setup()
		router := DefaultRouter()
		h := router.GetHandler()

		// Retrieve new created collection from admin context
		adminContext := userContext(baseContext(), "admin")
		col, err := h.GetCollection(adminContext, "", GetCollectionOptions{})
		So(err, ShouldNotBeNil)
		So(col, ShouldBeNil)
	})
}

func TestHandler_GetCollection2(t *testing.T) {
	Convey("COLLECTION - GET: cannot get collection info if user is not admin or not authenticated from a registered client application", t, func() {
		setup()

		router := DefaultRouter()
		h := router.GetHandler()

		// Try to retrieve collection info from non authenticated context
		_, err := h.GetCollection(baseContext(), "paris-sg", GetCollectionOptions{})
		So(err, ShouldNotBeNil)

		// Getting collection using a user from non registered client application
		user1Context := userContext(baseContext(), "user1")
		_, err = h.GetCollection(user1Context, "paris-sg", GetCollectionOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_GetCollection(t *testing.T) {
	Convey("COLLECTION - GET: can get collection info if user is admin or authenticated from a registered client application", t, func() {
		setup()

		router := DefaultRouter()
		h := router.GetHandler()

		// Try to retrieve collection info from non authenticated user
		col, err := h.GetCollection(baseContext(), "objects", GetCollectionOptions{})
		So(err, ShouldNotBeNil)
		So(col, ShouldBeNil)

		// Retrieve new created collection from user1 context
		user1Context := userContextFromRegisteredApplication(baseContext(), "user1")
		col, err = h.GetCollection(user1Context, "objects", GetCollectionOptions{})
		So(err, ShouldNotBeNil)
		So(col, ShouldBeNil)

		// Retrieve new created collection from user1 context
		col, err = h.GetCollection(user1Context, "juventus", GetCollectionOptions{})
		So(err, ShouldBeNil)
		So(col.Id, ShouldEqual, "juventus")

		// Retrieve new created collection from admin context
		col, err = h.GetCollection(fullAdminContext(), "juventus", GetCollectionOptions{})
		So(err, ShouldBeNil)
		So(col.Id, ShouldEqual, "juventus")

		// Retrieve all the created collection from admin context
		cols, err := h.ListCollections(fullAdminContext(), ListCollectionOptions{})
		So(err, ShouldBeNil)
		So(cols, ShouldHaveLength, 3)
	})
}

func TestHandler_ListCollection1(t *testing.T) {
	Convey("COLLECTION - LIST: cannot list collections if non authenticated or user is authenticated from an unregistered client application", t, func() {
		setup()

		router := DefaultRouter()
		h := router.GetHandler()

		// Try to retrieve collection info from non authenticated context
		_, err := h.ListCollections(baseContext(), ListCollectionOptions{})
		So(err, ShouldNotBeNil)

		// Retrieve new created collection from user1 context from non registered client application
		user1Context := userContext(baseContext(), "user1")
		_, err = h.ListCollections(user1Context, ListCollectionOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_ListCollection(t *testing.T) {
	Convey("COLLECTION - LIST: can list collections if user is admin or authenticated a registered client application", t, func() {
		setup()

		router := DefaultRouter()
		h := router.GetHandler()

		// Try to retrieve collection info from non authenticated user
		col, err := h.GetCollection(baseContext(), "objects", GetCollectionOptions{})
		So(err, ShouldNotBeNil)
		So(col, ShouldBeNil)

		// Retrieve new created collection from user1 context
		user1Context := userContext(baseContext(), "user1")
		col, err = h.GetCollection(user1Context, "objects", GetCollectionOptions{})
		So(err, ShouldNotBeNil)
		So(col, ShouldBeNil)

		// Retrieve new created collection from user1 context
		user1Context = userContextFromRegisteredApplication(baseContext(), "user1")
		col, err = h.GetCollection(user1Context, "juventus", GetCollectionOptions{})
		So(err, ShouldBeNil)
		So(col.Id, ShouldEqual, "juventus")

		// Retrieve new created collection from admin context
		col, err = h.GetCollection(fullAdminContext(), "juventus", GetCollectionOptions{})
		So(err, ShouldBeNil)
		So(col.Id, ShouldEqual, "juventus")

		// Retrieve all the created collection from admin context
		cols, err := h.ListCollections(fullAdminContext(), ListCollectionOptions{})
		So(err, ShouldBeNil)
		So(cols, ShouldHaveLength, 3)
	})
}

func TestHandler_PutObject1(t *testing.T) {
	Convey("OBJECTS - PUT: cannot create objects if context has no settings manager", t, func() {
		setup()

		router := DefaultRouter()
		h := router.GetHandler()
		data := `{"name": "user1", "age": 30, "city": "Paris"}`
		object := &pb.Object{
			Header: &pb.Header{
				Id:        "object1",
				CreatedBy: "user1",
				CreatedAt: time.Now().UnixNano(),
				Size:      int64(len(data)),
			},
			Data: data,
		}
		var err error

		user1Context := userContextFromRegisteredApplication(missingSettingsContext(), "user1")
		object.Header.Id, err = h.PutObject(user1Context, "objects", object, nil, nil, PutOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_PutObject2(t *testing.T) {
	Convey("OBJECTS - PUT: cannot put object if one of the items is not specified: collection, header, object data", t, func() {
		setup()

		router := DefaultRouter()
		h := router.GetHandler()
		data := `{"name": "user1", "age": 30, "city": "Paris"}`
		header := &pb.Header{
			Id:        "object1",
			CreatedBy: "user1",
			CreatedAt: time.Now().UnixNano(),
		}
		object := &pb.Object{
			Header: header,
			Data:   data,
		}
		var err error

		user1Context := userContextFromRegisteredApplication(baseContext(), "user1")

		_, err = h.PutObject(user1Context, "", object, nil, nil, PutOptions{})
		So(err, ShouldNotBeNil)

		_, err = h.PutObject(user1Context, "objects", nil, nil, nil, PutOptions{})
		So(err, ShouldNotBeNil)

		object.Header = nil
		_, err = h.PutObject(user1Context, "objects", object, nil, nil, PutOptions{})
		So(err, ShouldNotBeNil)

		object.Header = header
		object.Data = ""
		_, err = h.PutObject(user1Context, "objects", object, nil, nil, PutOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_PutObject3(t *testing.T) {
	Convey("OBJECTS - PUT: cannot create object if there is no 'SettingsDataMaxSizePath' is settings or has non numeric value", t, func() {
		setup()

		router := DefaultRouter()
		h := router.GetHandler()
		data := `{"name": "user1", "age": 30, "city": "Paris"}`
		object := &pb.Object{
			Header: &pb.Header{
				Id:        "object1",
				CreatedBy: "user1",
				CreatedAt: time.Now().UnixNano(),
				Size:      int64(len(data)),
			},
			Data: data,
		}

		// saving current value
		value, err := sm.Get(settings.DataMaxSizePath)
		So(err, ShouldBeNil)

		user1Context := userContextFromRegisteredApplication(missingSettingsContext(), "user1")
		object.Header.Id, err = h.PutObject(user1Context, "objects", object, nil, nil, PutOptions{})
		So(err, ShouldNotBeNil)

		err = sm.Delete(settings.DataMaxSizePath)
		So(err, ShouldBeNil)
		user1Context = userContext(baseContext(), "user1")
		object.Header.Id, err = h.PutObject(user1Context, "objects", object, nil, nil, PutOptions{})
		So(err, ShouldNotBeNil)

		err = sm.Set(settings.DataMaxSizePath, "no-number")
		So(err, ShouldBeNil)

		object.Header.Id, err = h.PutObject(user1Context, "objects", object, nil, nil, PutOptions{})
		So(err, ShouldNotBeNil)

		err = sm.Set(settings.DataMaxSizePath, value)
		So(err, ShouldBeNil)
	})
}

func TestHandler_PutObject4(t *testing.T) {
	Convey("OBJECTS - CREATE: cannot create object with data size greater than 'SettingsDataMaxSizePath' value", t, func() {
		setup()

		router := DefaultRouter()
		h := router.GetHandler()
		data := `{"name": "user1", "age": 30, "city": "Paris"}`
		object := &pb.Object{
			Header: &pb.Header{
				Id:        "object1",
				CreatedBy: "user1",
				CreatedAt: time.Now().UnixNano(),
				Size:      int64(len(data)),
			},
			Data: data,
		}

		// saving current value
		value, err := sm.Get(settings.DataMaxSizePath)
		So(err, ShouldBeNil)

		err = sm.Set(settings.DataMaxSizePath, "5")
		So(err, ShouldBeNil)

		user1Context := userContextFromRegisteredApplication(baseContext(), "user1")
		object.Header.Id, err = h.PutObject(user1Context, "objects", object, nil, nil, PutOptions{})
		So(err, ShouldNotBeNil)

		err = sm.Set(settings.DataMaxSizePath, value)
		So(err, ShouldBeNil)
	})
}

func TestHandler_PutObject5(t *testing.T) {
	Convey("OBJECTS - CREATE: cannot create object if context is not authenticated", t, func() {
		setup()

		router := DefaultRouter()
		h := router.GetHandler()

		data := `{"name": "Cristiano Ronaldo", "age": 35, "city": "Turin"}`
		object := &pb.Object{
			Header: &pb.Header{
				Id:        "cr7",
				CreatedAt: time.Now().UnixNano(),
				Size:      int64(len(data)),
			},
			Data: data,
		}

		_, err := h.PutObject(baseContext(), "juventus", object, nil, nil, PutOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_PutObject(t *testing.T) {
	Convey("OBJECTS - CREATE: can create object if user is authenticated from registered client application", t, func() {
		setup()

		router := DefaultRouter()
		h := router.GetHandler()

		data := `{"name": "Cristiano Ronaldo", "age": 35, "city": "Turin"}`
		object := &pb.Object{
			Header: &pb.Header{
				Id:        "cr7",
				CreatedAt: time.Now().UnixNano(),
				Size:      int64(len(data)),
			},
			Data: data,
		}

		var err error
		juveCtx := userContextFromRegisteredApplication(baseContext(), "pirlo")
		object.Header.Id, err = h.PutObject(juveCtx, "juventus", object, nil, nil, PutOptions{})
		So(err, ShouldBeNil)

		data = `{"name": "Lionel Messi", "age": 32, "city": "Barcelona"}`
		object = &pb.Object{
			Header: &pb.Header{
				Id:        "m10",
				CreatedAt: time.Now().UnixNano(),
				Size:      int64(len(data)),
			},
			Data: data,
		}
		barcaCtx := userContextFromRegisteredApplication(baseContext(), "koemann")
		object.Header.Id, err = h.PutObject(barcaCtx, "barcelona", object, nil, nil, PutOptions{})
		So(err, ShouldBeNil)

		data = `{"name": "Neymar Dos Santos", "age": 29, "city": "Paris"}`
		object = &pb.Object{
			Header: &pb.Header{
				Id:        "n10",
				CreatedAt: time.Now().UnixNano(),
				Size:      int64(len(data)),
			},
			Data: data,
		}
		psgCtx := userContextFromRegisteredApplication(baseContext(), "pochettino")
		object.Header.Id, err = h.PutObject(psgCtx, "paris-sg", object, nil, nil, PutOptions{})
		So(err, ShouldBeNil)
	})
}

func TestHandler_PatchObject1(t *testing.T) {
	Convey("OBJECTS - PATCH: cannot patch object if context has no settings manager", t, func() {
		setup()

		router := DefaultRouter()
		h := router.GetHandler()
		patch := &pb.Patch{
			ObjectId: "some-object-id",
			At:       "$.city",
			Data:     "bangkok",
		}

		user1Context := userContextFromRegisteredApplication(missingSettingsContext(), "user1")
		err := h.PatchObject(user1Context, "objects", patch, PatchOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_PatchObject2(t *testing.T) {
	Convey("OBJECTS - PATCH: cannot patch object if one of the items is not specified: collection, header, object data", t, func() {
		setup()

		router := DefaultRouter()
		h := router.GetHandler()
		patch := &pb.Patch{
			ObjectId: "some-object-id",
			At:       "$.city",
			Data:     "bangkok",
		}

		user1Context := userContextFromRegisteredApplication(baseContext(), "user1")

		err := h.PatchObject(user1Context, "", patch, PatchOptions{})
		So(err, ShouldNotBeNil)

		err = h.PatchObject(user1Context, "objects", patch, PatchOptions{})
		So(err, ShouldNotBeNil)

		patch.ObjectId = ""
		err = h.PatchObject(user1Context, "objects", patch, PatchOptions{})
		So(err, ShouldNotBeNil)

		patch.ObjectId = "some-object-id"
		patch.Data = ""
		err = h.PatchObject(user1Context, "objects", patch, PatchOptions{})
		So(err, ShouldNotBeNil)

		patch.ObjectId = "some-object-id"
		patch.Data = "bangkok"
		patch.At = ""
		err = h.PatchObject(user1Context, "objects", patch, PatchOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_PatchObject3(t *testing.T) {
	Convey("OBJECTS - PATCH: cannot create object if there is no 'SettingsDataMaxSizePath' is settings or has non numeric value", t, func() {
		setup()

		router := DefaultRouter()
		h := router.GetHandler()
		patch := &pb.Patch{
			ObjectId: "some-object-id",
			At:       "$.city",
			Data:     "bangkok",
		}

		// saving current value
		value, err := sm.Get(settings.DataMaxSizePath)
		So(err, ShouldBeNil)

		user1Context := userContextFromRegisteredApplication(missingSettingsContext(), "user1")
		err = h.PatchObject(user1Context, "objects", patch, PatchOptions{})
		So(err, ShouldNotBeNil)

		err = sm.Delete(settings.DataMaxSizePath)
		So(err, ShouldBeNil)
		user1Context = userContext(baseContext(), "user1")
		err = h.PatchObject(user1Context, "objects", patch, PatchOptions{})
		So(err, ShouldNotBeNil)

		err = sm.Set(settings.DataMaxSizePath, "no-number")
		So(err, ShouldBeNil)

		err = h.PatchObject(user1Context, "objects", patch, PatchOptions{})
		So(err, ShouldNotBeNil)

		err = sm.Set(settings.DataMaxSizePath, value)
		So(err, ShouldBeNil)
	})
}

func TestHandler_PatchObject4(t *testing.T) {
	Convey("OBJECTS - PATCH: cannot patch object with data size greater than 'SettingsDataMaxSizePath' value", t, func() {
		setup()

		router := DefaultRouter()
		h := router.GetHandler()
		patch := &pb.Patch{
			ObjectId: "some-object-id",
			At:       "$.city",
			Data:     "return to sender! return to sender",
		}

		// saving current value
		value, err := sm.Get(settings.DataMaxSizePath)
		So(err, ShouldBeNil)

		err = sm.Set(settings.DataMaxSizePath, "5")
		So(err, ShouldBeNil)

		user1Context := userContextFromRegisteredApplication(baseContext(), "user1")
		err = h.PatchObject(user1Context, "objects", patch, PatchOptions{})
		So(err, ShouldNotBeNil)

		err = sm.Set(settings.DataMaxSizePath, value)
		So(err, ShouldBeNil)
	})
}

func TestHandler_PatchObject(t *testing.T) {
	Convey("OBJECTS - PATCH: can patch object if context satisfies one of the object WRITE access rules", t, func() {
		setup()

		h := DefaultRouter().GetHandler()
		patch := &pb.Patch{
			ObjectId: "n10",
			At:       "$.age",
			Data:     "30",
		}

		// Try to create collection as admin
		psgCtx := userContextFromRegisteredApplication(baseContext(), "pochettino")

		err := h.PatchObject(psgCtx, "paris-sg", patch, PatchOptions{})
		So(err, ShouldBeNil)

		object, err := h.GetObject(psgCtx, "paris-sg", "n10", GetObjectOptions{At: "$.age"})
		So(err, ShouldBeNil)
		So(object.Data, ShouldEqual, "30")
	})
}

func TestHandler_MoveObject1(t *testing.T) {
	Convey("OBJECTS - MOVE: cannot move object if one of the items is not provided: collection-id, object-id, target-collection-id", t, func() {
		setup()
		router := DefaultRouter()
		h := router.GetHandler()

		err := h.MoveObject(baseContext(), "", "some-object-id", "paris-sg", nil, MoveOptions{})
		So(err, ShouldNotBeNil)

		err = h.MoveObject(baseContext(), "source-collection", "", "paris-sg", nil, MoveOptions{})
		So(err, ShouldNotBeNil)

		err = h.MoveObject(baseContext(), "source-collection", "some-object-id", "", nil, MoveOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_MoveObject(t *testing.T) {
	Convey("OBJECTS - MOVE: can move object if user is admin or is an authenticated user from a registered application and can create object in target collection", t, func() {
		setup()
		router := DefaultRouter()
		h := router.GetHandler()

		adminContext := userContextFromRegisteredApplication(baseContext(), "koemann")

		err := h.MoveObject(adminContext, "barcelona", "m10", "paris-sg", nil, MoveOptions{})
		So(err, ShouldBeNil)
	})
}

func TestHandler_GetObject1(t *testing.T) {
	Convey("OBJECTS - GET: cannot get object if one of the items is not provided: collection-id, object-id", t, func() {
		setup()
		router := DefaultRouter()
		h := router.GetHandler()

		_, err := h.GetObject(baseContext(), "", "some-object", GetObjectOptions{})
		So(err, ShouldNotBeNil)

		_, err = h.GetObject(baseContext(), "juventus", "", GetObjectOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_GetObject2(t *testing.T) {
	Convey("OBJECTS - GET: cannot get object info if user is not admin or the one who put it (according to collection rules)", t, func() {
		setup()
		router := DefaultRouter()
		h := router.GetHandler()

		psgCtx := userContextFromRegisteredApplication(baseContext(), "pochettino")
		_, err := h.GetObject(psgCtx, "juventus", "cr7", GetObjectOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_GetObject(t *testing.T) {
	Convey("OBJECTS - GET: can get object if user is admin or context satisfies one of the READ rule of the object", t, func() {
		setup()
		router := DefaultRouter()
		h := router.GetHandler()

		psgCtx := userContextFromRegisteredApplication(baseContext(), "pirlo")
		object, err := h.GetObject(psgCtx, "juventus", "cr7", GetObjectOptions{})
		So(err, ShouldBeNil)
		So(object.Header.Id, ShouldEqual, "cr7")
	})
}

func TestHandler_GetObjectHeader1(t *testing.T) {
	Convey("OBJECTS - HEADER: cannot get object header if one of the following items is not provided: collection-d, object-id", t, func() {
		setup()
		router := DefaultRouter()
		h := router.GetHandler()

		_, err := h.GetObjectHeader(baseContext(), "", "some-object", GetHeaderOptions{})
		So(err, ShouldNotBeNil)

		_, err = h.GetObjectHeader(baseContext(), "juventus", "", GetHeaderOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_GetObjectHeader2(t *testing.T) {
	Convey("OBJECTS - HEADER: cannot get object header if user is not admin and the user is not the one who put the data", t, func() {
		setup()
		router := DefaultRouter()
		h := router.GetHandler()

		psgCtx := userContextFromRegisteredApplication(baseContext(), "pochettino")
		header, err := h.GetObjectHeader(psgCtx, "juventus", "cr7", GetHeaderOptions{})
		So(err, ShouldNotBeNil)
		So(header, ShouldBeNil)
	})
}

func TestHandler_GetObjectHeader(t *testing.T) {
	Convey("OBJECTS - HEADER: can get object header if user is admin or context satisfies one of the READ rule of the object", t, func() {
		setup()
		router := DefaultRouter()
		h := router.GetHandler()

		psgCtx := userContextFromRegisteredApplication(baseContext(), "pirlo")

		header, err := h.GetObjectHeader(psgCtx, "juventus", "cr7", GetHeaderOptions{})
		So(err, ShouldBeNil)
		So(header.Id, ShouldEqual, "cr7")
	})
}

func TestHandler_ListObjects1(t *testing.T) {
	Convey("OBJECTS - LIST: cannot get a collection objects if no id is specified", t, func() {
		setup()
		router := DefaultRouter()
		h := router.GetHandler()

		_, err := h.ListObjects(baseContext(), "", ListOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_ListObjects2(t *testing.T) {
	Convey("OBJECTS - LIST: cannot get a collection objects if context has no settings manager", t, func() {
		setup()
		router := DefaultRouter()
		h := router.GetHandler()

		_, err := h.ListObjects(missingSettingsContext(), "juventus", ListOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_ListObjects3(t *testing.T) {
	Convey("OBJECTS - LIST: cannot get a collection objects if 'SettingsObjectListMaxCount' value is not in settings or has non numeric value", t, func() {
		setup()
		router := DefaultRouter()
		h := router.GetHandler()

		// saving current value
		value, err := sm.Get(settings.ObjectListMaxCount)
		So(err, ShouldBeNil)

		err = sm.Delete(settings.ObjectListMaxCount)
		So(err, ShouldBeNil)

		_, err = h.ListObjects(baseContext(), "juventus", ListOptions{})
		So(err, ShouldNotBeNil)

		err = sm.Set(settings.ObjectListMaxCount, "no-number")
		So(err, ShouldBeNil)

		_, err = h.ListObjects(baseContext(), "juventus", ListOptions{})
		So(err, ShouldNotBeNil)

		err = sm.Set(settings.ObjectListMaxCount, value)
		So(err, ShouldBeNil)
	})
}

func TestHandler_ListObjects4(t *testing.T) {
	Convey("OBJECTS - LIST: cannot get a collection objects if user is not authenticated", t, func() {
		setup()
		router := DefaultRouter()
		h := router.GetHandler()

		c, err := h.ListObjects(baseContext(), "juventus", ListOptions{})
		So(err, ShouldBeNil)

		defer func() {
			_ = c.Close()
		}()

		count := 0
		for {
			_, err = c.Browse()
			if err != nil {
				So(err, ShouldEqual, io.EOF)
				break
			}
			count++
		}
		So(count, ShouldEqual, 0)
	})
}

func TestHandler_ListObjects5(t *testing.T) {
	Convey("OBJECTS - LIST: cannot list objects if no collection matches the provided id", t, func() {
		setup()
		router := DefaultRouter()
		h := router.GetHandler()

		_, err := h.ListObjects(baseContext(), "some-collection-id", ListOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_ListObjects(t *testing.T) {
	Convey("OBJECTS - LIST: can list a collection objects if user is authenticated from a registered client application", t, func() {
		setup()
		router := DefaultRouter()
		h := router.GetHandler()

		userCtx := userContextFromRegisteredApplication(baseContext(), "pirlo")
		c, err := h.ListObjects(userCtx, "juventus", ListOptions{Offset: utime.Now()})
		So(err, ShouldBeNil)
		defer func() {
			_ = c.Close()
		}()
	})
}

func TestHandler_SearchObjects(t *testing.T) {
	Convey("OBJECTS - SEARCH: cannot search if one the following parameters is not provided: collection-id, query", t, func() {
		setup()
		router := DefaultRouter()
		h := router.GetHandler()

		_, err := h.SearchObjects(baseContext(), "", &pb.SearchQuery{}, SearchObjectsOptions{})
		So(err, ShouldNotBeNil)

		_, err = h.SearchObjects(baseContext(), "some-collection-id", nil, SearchObjectsOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_DeleteObject1(t *testing.T) {
	Convey("OBJECTS - DELETE: cannot delete if one the followings parameters is not provided: collection-id, object-id", t, func() {
		setup()
		router := DefaultRouter()
		h := router.GetHandler()

		err := h.DeleteObject(baseContext(), "", "some-object", DeleteObjectOptions{})
		So(err, ShouldNotBeNil)

		err = h.DeleteObject(baseContext(), "juventus", "", DeleteObjectOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_DeleteObject(t *testing.T) {
	Convey("OBJECTS - HEADER: can delete an object if user is admin or or the context satisfies one of DELETE rules if the object", t, func() {
		setup()
		router := DefaultRouter()
		h := router.GetHandler()

		psgCtx := userContextFromRegisteredApplication(baseContext(), "pochettino")

		err := h.DeleteObject(psgCtx, "paris-sg", "n10", DeleteObjectOptions{})
		So(err, ShouldBeNil)
	})
}

func TestHandler_DeleteCollection1(t *testing.T) {
	Convey("COLLECTION - DELETE: cannot delete a collection if id is not provided", t, func() {
		setup()

		router := DefaultRouter()
		h := router.GetHandler()

		// Retrieve new created collection from admin context
		adminContext := userContext(baseContext(), "admin")
		err := h.DeleteCollection(adminContext, "", DeleteCollectionOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_DeleteCollection2(t *testing.T) {
	Convey("COLLECTION - DELETE: cannot delete a collection if user is not admin", t, func() {
		setup()

		router := DefaultRouter()
		h := router.GetHandler()

		err := h.DeleteCollection(baseContext(), "juventus", DeleteCollectionOptions{})
		So(err, ShouldNotBeNil)

		user1Context := userContext(baseContext(), "pirlo")
		err = h.DeleteCollection(user1Context, "juventus", DeleteCollectionOptions{})
		So(err, ShouldNotBeNil)

		user1Context = userContextFromRegisteredApplication(baseContext(), "pirlo")
		err = h.DeleteCollection(user1Context, "juventus", DeleteCollectionOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_DeleteCollection(t *testing.T) {
	Convey("COLLECTION - DELETE: can collection if the user is admin", t, func() {
		setup()

		router := DefaultRouter()
		h := router.GetHandler()

		// Try to create collection as admin
		adminContext := userContext(adminAppContext(baseContext()), "admin")

		err := h.DeleteCollection(adminContext, "juventus", DeleteCollectionOptions{})
		So(err, ShouldBeNil)

		err = h.DeleteCollection(adminContext, "barcelona", DeleteCollectionOptions{})
		So(err, ShouldBeNil)

		err = h.DeleteCollection(adminContext, "paris-sg", DeleteCollectionOptions{})
		So(err, ShouldBeNil)

		// Retrieve all the created collection from admin context
		cols, err := h.ListCollections(adminContext, ListCollectionOptions{})
		So(err, ShouldBeNil)
		So(cols, ShouldHaveLength, 0)
	})
}
