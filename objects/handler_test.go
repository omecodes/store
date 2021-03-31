package objects

import (
	"context"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/omecodes/bome"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/common/utime"
	se "github.com/omecodes/store/search-engine"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"testing"
	"time"
)

var (
	db       DB
	settings SettingsManager
	acl      ACLManager

	juveTeam = &Collection{
		Id:          "juventus",
		Label:       "Team of Juventus",
		Description: "List of Juventus FC players",
		NumberIndex: nil,
		TextIndexes: nil,
		FieldsIndex: nil,
		DefaultAccessSecurityRules: &PathAccessRules{
			AccessRules: map[string]*AccessRules{
				"$": {
					Label:       "Default",
					Description: "Only owner can edit and delete, everybody can read. Admin can do everything",
					Read: []*auth.Permission{
						{
							Name:        "public",
							Label:       "Readable",
							Description: "Readable by everybody",
							Rule:        "user.name==data.created_by",
						},
					},
					Write: []*auth.Permission{
						{
							Name:        "restricted-write",
							Label:       "Restricted Write",
							Description: "Only creator can write",
							Rule:        "user.name==data.created_by",
						},
					},
					Delete: []*auth.Permission{
						{
							Name:        "restricted-delete",
							Label:       "Restricted Delete",
							Description: "Only creator can delete",
							Rule:        "user.name==data.created_by",
						},
					},
				},
			},
		},
	}

	barcaTeam = &Collection{
		Id:          "barcelona",
		Label:       "Team of Barcelona",
		Description: "List of Barcelona FC players",
		NumberIndex: nil,
		TextIndexes: nil,
		FieldsIndex: nil,
		DefaultAccessSecurityRules: &PathAccessRules{
			AccessRules: map[string]*AccessRules{
				"$": {
					Label:       "Default",
					Description: "Only owner can edit and delete, everybody can read. Admin can do everything",
					Read: []*auth.Permission{
						{
							Name:        "public",
							Label:       "Readable",
							Description: "Readable by everybody",
							Rule:        "user.name==data.created_by && app.key!=''",
						},
					},
					Write: []*auth.Permission{
						{
							Name:        "restricted-write",
							Label:       "Restricted Write",
							Description: "Only creator can write",
							Rule:        "user.name==data.created_by && app.key!=''",
						},
					},
					Delete: []*auth.Permission{
						{
							Name:        "restricted-delete",
							Label:       "Restricted Delete",
							Description: "Only creator can delete",
							Rule:        "user.name==data.created_by && app.key!=''",
						},
					},
				},
			},
		},
	}

	psgTeam = &Collection{
		Id:          "paris-sg",
		Label:       "Team of PSG",
		Description: "List of Paris Saint-Germain players",
		NumberIndex: nil,
		TextIndexes: nil,
		FieldsIndex: nil,
		DefaultAccessSecurityRules: &PathAccessRules{
			AccessRules: map[string]*AccessRules{
				"$": {
					Label:       "Default",
					Description: "Only owner can edit and delete, everybody can read. Admin can do everything",
					Read: []*auth.Permission{
						{
							Name:        "public",
							Label:       "Readable",
							Description: "Readable by everybody",
							Rule:        "user.name==data.created_by && app.key!=''",
						},
					},
					Write: []*auth.Permission{
						{
							Name:        "restricted-write",
							Label:       "Restricted Write",
							Description: "Only creator can write",
							Rule:        "user.name==data.created_by && app.key!=''",
						},
					},
					Delete: []*auth.Permission{
						{
							Name:        "restricted-delete",
							Label:       "Restricted Delete",
							Description: "Only creator can delete",
							Rule:        "user.name==data.created_by && app.key!=''",
						},
					},
				},
			},
		},
	}
)

func getContext() context.Context {
	ctx := context.Background()
	ctx = ContextWithACLStore(ctx, acl)
	ctx = ContextWithStore(ctx, db)
	ctx = ContextWithSettings(ctx, settings)
	return ctx
}

func getContextWithNoSettings() context.Context {
	ctx := context.Background()
	ctx = ContextWithACLStore(ctx, acl)
	ctx = ContextWithStore(ctx, db)
	return ctx
}

func userContext(ctx context.Context, name string) context.Context {
	return auth.ContextWithUser(ctx, &auth.User{Name: name})
}

func userContextInRegisteredClient(ctx context.Context, name string) context.Context {
	// this create a context as if it was created by the authentication interceptor when
	// receiving a request from a user by the means of a registered app client
	ctx = auth.ContextWithUser(ctx, &auth.User{Name: name})
	return auth.ContextWithApp(ctx, &auth.ClientApp{
		Key:    "some-key",
		Secret: "some-secret",
	})
}

func initDB() {
	if db == nil {
		conn, err := sql.Open("sqlite3", ":memory:")
		So(err, ShouldBeNil)

		db, err = NewSqlDB(conn, bome.SQLite3, "store")
		So(err, ShouldBeNil)
	}

	if settings == nil {
		conn, err := sql.Open("sqlite3", ":memory:")
		So(err, ShouldBeNil)

		settings, err = NewSQLSettings(conn, bome.SQLite3, "settings")
		So(err, ShouldBeNil)
	}

	if acl == nil {
		conn, err := sql.Open("sqlite3", ":memory:")
		So(err, ShouldBeNil)

		acl, err = NewSQLACLStore(conn, bome.SQLite3, "access_rules")
		So(err, ShouldBeNil)
	}
}

func Test_DBInitialization(t *testing.T) {
	Convey("Database initialization should be executed with no errors", t, func() {
		initDB()
	})
}

func TestHandler_CreateCollection1(t *testing.T) {
	Convey("COLLECTION - CREATE: cannot create a collection if id or security rules are not provided", t, func() {
		initDB()

		router := DefaultRouter()
		handler := router.GetHandler()

		col := &Collection{
			Id:          "",
			Label:       "Objects",
			Description: "List of random object created for the sake of test",
			NumberIndex: nil,
			TextIndexes: nil,
			FieldsIndex: nil,
			DefaultAccessSecurityRules: &PathAccessRules{
				AccessRules: map[string]*AccessRules{
					"$": {
						Label:       "Default",
						Description: "Only owner can edit and delete, everybody can read. Admin can do everything",
						Read: []*auth.Permission{
							{
								Name:        "public",
								Label:       "Readable",
								Description: "Readable by everybody",
								Rule:        "true",
							},
						},
						Write: []*auth.Permission{
							{
								Name:        "restricted-write",
								Label:       "Restricted Write",
								Description: "Only creator can write",
								Rule:        "user.name==data.created_by",
							},
						},
						Delete: []*auth.Permission{
							{
								Name:        "restricted-delete",
								Label:       "Restricted Delete",
								Description: "Only creator can delete",
								Rule:        "user.name==data.created_by",
							},
						},
					},
				},
			},
		}

		// Try to create collection as admin
		adminContext := userContext(getContext(), "admin")
		err := handler.CreateCollection(adminContext, col)
		So(err, ShouldNotBeNil)

		col.Id = "objects"
		col.DefaultAccessSecurityRules = nil
		err = handler.CreateCollection(adminContext, col)
		So(err, ShouldNotBeNil)

		// Try to create collection as unauthenticated user
		err = handler.CreateCollection(getContext(), col)
		So(err, ShouldNotBeNil)

		// Try to create collection as user1
		user1Context := userContext(getContext(), "user1")
		err = handler.CreateCollection(user1Context, col)
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_CreateCollection2(t *testing.T) {
	Convey("COLLECTION - CREATE: cannot create a collection if user is not admin", t, func() {
		initDB()
		router := DefaultRouter()
		handler := router.GetHandler()

		// Try to create collection as admin
		adminContext := userContext(getContext(), "user1")
		err := handler.CreateCollection(adminContext, juveTeam)
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_CreateCollection(t *testing.T) {
	Convey("COLLECTION - CREATE: can create a collection if user is admin", t, func() {
		initDB()

		router := DefaultRouter()
		handler := router.GetHandler()

		// Try to create collection as admin
		adminContext := userContext(getContext(), "admin")
		err := handler.CreateCollection(adminContext, juveTeam)
		So(err, ShouldBeNil)
		err = handler.CreateCollection(adminContext, barcaTeam)
		So(err, ShouldBeNil)
		err = handler.CreateCollection(adminContext, psgTeam)
		So(err, ShouldBeNil)

		// Retrieve new created collection from admin context
		col, err := handler.GetCollection(adminContext, "juventus")
		So(err, ShouldBeNil)
		So(col.Id, ShouldEqual, "juventus")
	})
}

func TestHandler_GetCollection1(t *testing.T) {
	Convey("COLLECTION - GET: cannot get a collection info if id is not provided", t, func() {
		initDB()
		router := DefaultRouter()
		handler := router.GetHandler()

		// Retrieve new created collection from admin context
		adminContext := userContext(getContext(), "admin")
		col, err := handler.GetCollection(adminContext, "")
		So(err, ShouldNotBeNil)
		So(col, ShouldBeNil)
	})
}

func TestHandler_GetCollection2(t *testing.T) {
	Convey("COLLECTION - GET: cannot get collection info if user is not admin or not authenticated from a registered client application", t, func() {
		initDB()

		router := DefaultRouter()
		handler := router.GetHandler()

		// Try to retrieve collection info from non authenticated context
		_, err := handler.GetCollection(getContext(), "paris-sg")
		So(err, ShouldNotBeNil)

		// Getting collection using a user from non registered client application
		user1Context := userContext(getContext(), "user1")
		_, err = handler.GetCollection(user1Context, "paris-sg")
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_GetCollection(t *testing.T) {
	Convey("COLLECTION - GET: can get collection info if user is admin or authenticated from a registered client application", t, func() {
		initDB()

		router := DefaultRouter()
		handler := router.GetHandler()

		// Try to retrieve collection info from non authenticated user
		col, err := handler.GetCollection(getContext(), "objects")
		So(err, ShouldNotBeNil)
		So(col, ShouldBeNil)

		// Retrieve new created collection from user1 context
		user1Context := userContext(getContext(), "user1")
		col, err = handler.GetCollection(user1Context, "objects")
		So(err, ShouldNotBeNil)
		So(col, ShouldBeNil)

		// Retrieve new created collection from user1 context
		user1Context = userContextInRegisteredClient(getContext(), "user1")
		col, err = handler.GetCollection(user1Context, "juventus")
		So(err, ShouldBeNil)
		So(col.Id, ShouldEqual, "juventus")

		// Retrieve new created collection from admin context
		adminContext := userContext(getContext(), "admin")
		col, err = handler.GetCollection(adminContext, "juventus")
		So(err, ShouldBeNil)
		So(col.Id, ShouldEqual, "juventus")

		// Retrieve all the created collection from admin context
		cols, err := handler.ListCollections(adminContext)
		So(err, ShouldBeNil)
		So(cols, ShouldHaveLength, 3)
	})
}

func TestHandler_ListCollection1(t *testing.T) {
	Convey("COLLECTION - LIST: cannot list collections if non authenticated or user is authenticated from an unregistered client application", t, func() {
		initDB()

		router := DefaultRouter()
		handler := router.GetHandler()

		// Try to retrieve collection info from non authenticated context
		_, err := handler.ListCollections(getContext())
		So(err, ShouldNotBeNil)

		// Retrieve new created collection from user1 context from non registered client application
		user1Context := userContext(getContext(), "user1")
		_, err = handler.ListCollections(user1Context)
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_ListCollection(t *testing.T) {
	Convey("COLLECTION - LIST: can list collections if user is admin or authenticated a registered client application", t, func() {
		initDB()

		router := DefaultRouter()
		handler := router.GetHandler()

		// Try to retrieve collection info from non authenticated user
		col, err := handler.GetCollection(getContext(), "objects")
		So(err, ShouldNotBeNil)
		So(col, ShouldBeNil)

		// Retrieve new created collection from user1 context
		user1Context := userContext(getContext(), "user1")
		col, err = handler.GetCollection(user1Context, "objects")
		So(err, ShouldNotBeNil)
		So(col, ShouldBeNil)

		// Retrieve new created collection from user1 context
		user1Context = userContextInRegisteredClient(getContext(), "user1")
		col, err = handler.GetCollection(user1Context, "juventus")
		So(err, ShouldBeNil)
		So(col.Id, ShouldEqual, "juventus")

		// Retrieve new created collection from admin context
		adminContext := userContext(getContext(), "admin")
		col, err = handler.GetCollection(adminContext, "juventus")
		So(err, ShouldBeNil)
		So(col.Id, ShouldEqual, "juventus")

		// Retrieve all the created collection from admin context
		cols, err := handler.ListCollections(adminContext)
		So(err, ShouldBeNil)
		So(cols, ShouldHaveLength, 3)
	})
}

func TestHandler_PutObject1(t *testing.T) {
	Convey("OBJECTS - PUT: cannot create objects if context has no settings manager", t, func() {
		initDB()

		router := DefaultRouter()
		handler := router.GetHandler()
		data := `{"name": "user1", "age": 30, "city": "Paris"}`
		object := &Object{
			Header: &Header{
				Id:        "object1",
				CreatedBy: "user1",
				CreatedAt: time.Now().UnixNano(),
				Size:      int64(len(data)),
			},
			Data: data,
		}
		var err error

		user1Context := userContextInRegisteredClient(getContextWithNoSettings(), "user1")
		object.Header.Id, err = handler.PutObject(user1Context, "objects", object, nil, nil, PutOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_PutObject2(t *testing.T) {
	Convey("OBJECTS - PUT: cannot put object if one of the items is not specified: collection, header, object data", t, func() {
		initDB()

		router := DefaultRouter()
		handler := router.GetHandler()
		data := `{"name": "user1", "age": 30, "city": "Paris"}`
		header := &Header{
			Id:        "object1",
			CreatedBy: "user1",
			CreatedAt: time.Now().UnixNano(),
		}
		object := &Object{
			Header: header,
			Data:   data,
		}
		var err error

		user1Context := userContextInRegisteredClient(getContext(), "user1")

		_, err = handler.PutObject(user1Context, "", object, nil, nil, PutOptions{})
		So(err, ShouldNotBeNil)

		_, err = handler.PutObject(user1Context, "objects", nil, nil, nil, PutOptions{})
		So(err, ShouldNotBeNil)

		object.Header = nil
		_, err = handler.PutObject(user1Context, "objects", object, nil, nil, PutOptions{})
		So(err, ShouldNotBeNil)

		object.Header = header
		object.Data = ""
		_, err = handler.PutObject(user1Context, "objects", object, nil, nil, PutOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_PutObject3(t *testing.T) {
	Convey("OBJECTS - PUT: cannot create object if there is no 'SettingsDataMaxSizePath' is settings or has non numeric value", t, func() {
		initDB()

		router := DefaultRouter()
		handler := router.GetHandler()
		data := `{"name": "user1", "age": 30, "city": "Paris"}`
		object := &Object{
			Header: &Header{
				Id:        "object1",
				CreatedBy: "user1",
				CreatedAt: time.Now().UnixNano(),
				Size:      int64(len(data)),
			},
			Data: data,
		}

		// saving current value
		value, err := settings.Get(SettingsDataMaxSizePath)
		So(err, ShouldBeNil)

		user1Context := userContextInRegisteredClient(getContextWithNoSettings(), "user1")
		object.Header.Id, err = handler.PutObject(user1Context, "objects", object, nil, nil, PutOptions{})
		So(err, ShouldNotBeNil)

		err = settings.Delete(SettingsDataMaxSizePath)
		So(err, ShouldBeNil)
		user1Context = userContext(getContext(), "user1")
		object.Header.Id, err = handler.PutObject(user1Context, "objects", object, nil, nil, PutOptions{})
		So(err, ShouldNotBeNil)

		err = settings.Set(SettingsDataMaxSizePath, "no-number")
		So(err, ShouldBeNil)

		object.Header.Id, err = handler.PutObject(user1Context, "objects", object, nil, nil, PutOptions{})
		So(err, ShouldNotBeNil)

		err = settings.Set(SettingsDataMaxSizePath, value)
		So(err, ShouldBeNil)
	})
}

func TestHandler_PutObject4(t *testing.T) {
	Convey("OBJECTS - CREATE: cannot create object with data size greater than 'SettingsDataMaxSizePath' value", t, func() {
		initDB()

		router := DefaultRouter()
		handler := router.GetHandler()
		data := `{"name": "user1", "age": 30, "city": "Paris"}`
		object := &Object{
			Header: &Header{
				Id:        "object1",
				CreatedBy: "user1",
				CreatedAt: time.Now().UnixNano(),
				Size:      int64(len(data)),
			},
			Data: data,
		}

		// saving current value
		value, err := settings.Get(SettingsDataMaxSizePath)
		So(err, ShouldBeNil)

		err = settings.Set(SettingsDataMaxSizePath, "5")
		So(err, ShouldBeNil)

		user1Context := userContextInRegisteredClient(getContext(), "user1")
		object.Header.Id, err = handler.PutObject(user1Context, "objects", object, nil, nil, PutOptions{})
		So(err, ShouldNotBeNil)

		err = settings.Set(SettingsDataMaxSizePath, value)
		So(err, ShouldBeNil)
	})
}

func TestHandler_PutObject5(t *testing.T) {
	Convey("OBJECTS - CREATE: cannot create object if context is not authenticated", t, func() {
		initDB()

		router := DefaultRouter()
		handler := router.GetHandler()

		data := `{"name": "Cristiano Ronaldo", "age": 35, "city": "Turin"}`
		object := &Object{
			Header: &Header{
				Id:        "cr7",
				CreatedAt: time.Now().UnixNano(),
				Size:      int64(len(data)),
			},
			Data: data,
		}

		_, err := handler.PutObject(getContext(), "juventus", object, nil, nil, PutOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_PutObject(t *testing.T) {
	Convey("OBJECTS - CREATE: can create object if user is authenticated from registered client application", t, func() {
		initDB()

		router := DefaultRouter()
		handler := router.GetHandler()

		data := `{"name": "Cristiano Ronaldo", "age": 35, "city": "Turin"}`
		object := &Object{
			Header: &Header{
				Id:        "cr7",
				CreatedAt: time.Now().UnixNano(),
				Size:      int64(len(data)),
			},
			Data: data,
		}

		var err error
		juveCtx := userContextInRegisteredClient(getContext(), "pirlo")
		object.Header.Id, err = handler.PutObject(juveCtx, "juventus", object, nil, nil, PutOptions{})
		So(err, ShouldBeNil)

		data = `{"name": "Lionel Messi", "age": 32, "city": "Barcelona"}`
		object = &Object{
			Header: &Header{
				Id:        "m10",
				CreatedAt: time.Now().UnixNano(),
				Size:      int64(len(data)),
			},
			Data: data,
		}
		barcaCtx := userContextInRegisteredClient(getContext(), "koemann")
		object.Header.Id, err = handler.PutObject(barcaCtx, "barcelona", object, nil, nil, PutOptions{})
		So(err, ShouldBeNil)

		data = `{"name": "Neymar Dos Santos", "age": 29, "city": "Paris"}`
		object = &Object{
			Header: &Header{
				Id:        "n10",
				CreatedAt: time.Now().UnixNano(),
				Size:      int64(len(data)),
			},
			Data: data,
		}
		psgCtx := userContextInRegisteredClient(getContext(), "pochettino")
		object.Header.Id, err = handler.PutObject(psgCtx, "paris-sg", object, nil, nil, PutOptions{})
		So(err, ShouldBeNil)
	})
}

func TestHandler_PatchObject1(t *testing.T) {
	Convey("OBJECTS - PATCH: cannot patch object if context has no settings manager", t, func() {
		initDB()

		router := DefaultRouter()
		handler := router.GetHandler()
		patch := &Patch{
			ObjectId: "some-object-id",
			At:       "$.city",
			Data:     "bangkok",
		}

		user1Context := userContextInRegisteredClient(getContextWithNoSettings(), "user1")
		err := handler.PatchObject(user1Context, "objects", patch, PatchOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_PatchObject2(t *testing.T) {
	Convey("OBJECTS - PATCH: cannot patch object if one of the items is not specified: collection, header, object data", t, func() {
		initDB()

		router := DefaultRouter()
		handler := router.GetHandler()
		patch := &Patch{
			ObjectId: "some-object-id",
			At:       "$.city",
			Data:     "bangkok",
		}

		user1Context := userContextInRegisteredClient(getContext(), "user1")

		err := handler.PatchObject(user1Context, "", patch, PatchOptions{})
		So(err, ShouldNotBeNil)

		err = handler.PatchObject(user1Context, "objects", patch, PatchOptions{})
		So(err, ShouldNotBeNil)

		patch.ObjectId = ""
		err = handler.PatchObject(user1Context, "objects", patch, PatchOptions{})
		So(err, ShouldNotBeNil)

		patch.ObjectId = "some-object-id"
		patch.Data = ""
		err = handler.PatchObject(user1Context, "objects", patch, PatchOptions{})
		So(err, ShouldNotBeNil)

		patch.ObjectId = "some-object-id"
		patch.Data = "bangkok"
		patch.At = ""
		err = handler.PatchObject(user1Context, "objects", patch, PatchOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_PatchObject3(t *testing.T) {
	Convey("OBJECTS - PATCH: cannot create object if there is no 'SettingsDataMaxSizePath' is settings or has non numeric value", t, func() {
		initDB()

		router := DefaultRouter()
		handler := router.GetHandler()
		patch := &Patch{
			ObjectId: "some-object-id",
			At:       "$.city",
			Data:     "bangkok",
		}

		// saving current value
		value, err := settings.Get(SettingsDataMaxSizePath)
		So(err, ShouldBeNil)

		user1Context := userContextInRegisteredClient(getContextWithNoSettings(), "user1")
		err = handler.PatchObject(user1Context, "objects", patch, PatchOptions{})
		So(err, ShouldNotBeNil)

		err = settings.Delete(SettingsDataMaxSizePath)
		So(err, ShouldBeNil)
		user1Context = userContext(getContext(), "user1")
		err = handler.PatchObject(user1Context, "objects", patch, PatchOptions{})
		So(err, ShouldNotBeNil)

		err = settings.Set(SettingsDataMaxSizePath, "no-number")
		So(err, ShouldBeNil)

		err = handler.PatchObject(user1Context, "objects", patch, PatchOptions{})
		So(err, ShouldNotBeNil)

		err = settings.Set(SettingsDataMaxSizePath, value)
		So(err, ShouldBeNil)
	})
}

func TestHandler_PatchObject4(t *testing.T) {
	Convey("OBJECTS - PATCH: cannot patch object with data size greater than 'SettingsDataMaxSizePath' value", t, func() {
		initDB()

		router := DefaultRouter()
		handler := router.GetHandler()
		patch := &Patch{
			ObjectId: "some-object-id",
			At:       "$.city",
			Data:     "return to sender! return to sender",
		}

		// saving current value
		value, err := settings.Get(SettingsDataMaxSizePath)
		So(err, ShouldBeNil)

		err = settings.Set(SettingsDataMaxSizePath, "5")
		So(err, ShouldBeNil)

		user1Context := userContextInRegisteredClient(getContext(), "user1")
		err = handler.PatchObject(user1Context, "objects", patch, PatchOptions{})
		So(err, ShouldNotBeNil)

		err = settings.Set(SettingsDataMaxSizePath, value)
		So(err, ShouldBeNil)
	})
}

func TestHandler_PatchObject(t *testing.T) {
	Convey("OBJECTS - PATCH: can patch object if context satisfies one of the object WRITE access rules", t, func() {
		initDB()

		router := DefaultRouter()
		handler := router.GetHandler()
		patch := &Patch{
			ObjectId: "n10",
			At:       "$.age",
			Data:     "30",
		}

		psgCtx := userContextInRegisteredClient(getContext(), "pochettino")
		err := handler.PatchObject(psgCtx, "paris-sg", patch, PatchOptions{})
		So(err, ShouldBeNil)

		object, err := handler.GetObject(psgCtx, "paris-sg", "n10", GetOptions{At: "$.age"})
		So(err, ShouldBeNil)
		So(object.Data, ShouldEqual, "30")
	})
}

func TestHandler_MoveObject1(t *testing.T) {
	Convey("OBJECTS - MOVE: cannot move object if one of the items is not provided: collection-id, object-id, target-collection-id", t, func() {
		initDB()
		router := DefaultRouter()
		handler := router.GetHandler()

		err := handler.MoveObject(getContext(), "", "some-object-id", "paris-sg", nil, MoveOptions{})
		So(err, ShouldNotBeNil)

		err = handler.MoveObject(getContext(), "source-collection", "", "paris-sg", nil, MoveOptions{})
		So(err, ShouldNotBeNil)

		err = handler.MoveObject(getContext(), "source-collection", "some-object-id", "", nil, MoveOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_MoveObject(t *testing.T) {
	Convey("OBJECTS - MOVE: can move object if user is admin or is an authenticated user from a registered application and can create object in target collection", t, func() {
		initDB()
		router := DefaultRouter()
		handler := router.GetHandler()

		adminContext := userContextInRegisteredClient(getContext(), "admin")

		err := handler.MoveObject(adminContext, "barcelona", "m10", "paris-sg", nil, MoveOptions{})
		So(err, ShouldBeNil)
	})
}

func TestHandler_GetObject1(t *testing.T) {
	Convey("OBJECTS - GET: cannot get object if one of the items is not provided: collection-id, object-id", t, func() {
		initDB()
		router := DefaultRouter()
		handler := router.GetHandler()

		_, err := handler.GetObject(getContext(), "", "some-object", GetOptions{})
		So(err, ShouldNotBeNil)

		_, err = handler.GetObject(getContext(), "juventus", "", GetOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_GetObject2(t *testing.T) {
	Convey("OBJECTS - GET: cannot get object info if user is not admin or the one who put it (according to collection rules)", t, func() {
		initDB()
		router := DefaultRouter()
		handler := router.GetHandler()

		psgCtx := userContextInRegisteredClient(getContext(), "pochettino")
		_, err := handler.GetObject(psgCtx, "juventus", "cr7", GetOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_GetObject(t *testing.T) {
	Convey("OBJECTS - GET: can get object if user is admin or context satisfies one of the READ rule of the object", t, func() {
		initDB()
		router := DefaultRouter()
		handler := router.GetHandler()

		psgCtx := userContextInRegisteredClient(getContext(), "pirlo")
		object, err := handler.GetObject(psgCtx, "juventus", "cr7", GetOptions{})
		So(err, ShouldBeNil)
		So(object.Header.Id, ShouldEqual, "cr7")
	})
}

func TestHandler_GetObjectHeader1(t *testing.T) {
	Convey("OBJECTS - HEADER: cannot get object header if one of the following items is not provided: collection-d, object-id", t, func() {
		initDB()
		router := DefaultRouter()
		handler := router.GetHandler()

		_, err := handler.GetObjectHeader(getContext(), "", "some-object")
		So(err, ShouldNotBeNil)

		_, err = handler.GetObjectHeader(getContext(), "juventus", "")
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_GetObjectHeader2(t *testing.T) {
	Convey("OBJECTS - HEADER: cannot get object header if user is not admin and the user is not the one who put the data", t, func() {
		initDB()
		router := DefaultRouter()
		handler := router.GetHandler()

		barcelonaCtx := userContextInRegisteredClient(getContext(), "koemann")
		header, err := handler.GetObjectHeader(barcelonaCtx, "juventus", "cr7")
		So(err, ShouldNotBeNil)
		So(header, ShouldBeNil)
	})
}

func TestHandler_GetObjectHeader(t *testing.T) {
	Convey("OBJECTS - HEADER: can get object header if user is admin or context satisfies one of the READ rule of the object", t, func() {
		initDB()
		router := DefaultRouter()
		handler := router.GetHandler()

		psgCtx := userContextInRegisteredClient(getContext(), "pirlo")

		header, err := handler.GetObjectHeader(psgCtx, "juventus", "cr7")
		So(err, ShouldBeNil)
		So(header.Id, ShouldEqual, "cr7")
	})
}

func TestHandler_ListObjects1(t *testing.T) {
	Convey("OBJECTS - LIST: cannot get a collection objects if no id is specified", t, func() {
		initDB()
		router := DefaultRouter()
		handler := router.GetHandler()

		_, err := handler.ListObjects(getContext(), "", ListOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_ListObjects2(t *testing.T) {
	Convey("OBJECTS - LIST: cannot get a collection objects if context has no settings manager", t, func() {
		initDB()
		router := DefaultRouter()
		handler := router.GetHandler()

		_, err := handler.ListObjects(getContextWithNoSettings(), "juventus", ListOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_ListObjects3(t *testing.T) {
	Convey("OBJECTS - LIST: cannot get a collection objects if 'SettingsObjectListMaxCount' value is not in settings or has non numeric value", t, func() {
		initDB()
		router := DefaultRouter()
		handler := router.GetHandler()

		// saving current value
		value, err := settings.Get(SettingsObjectListMaxCount)
		So(err, ShouldBeNil)

		err = settings.Delete(SettingsObjectListMaxCount)
		So(err, ShouldBeNil)

		_, err = handler.ListObjects(getContext(), "juventus", ListOptions{})
		So(err, ShouldNotBeNil)

		err = settings.Set(SettingsObjectListMaxCount, "no-number")
		So(err, ShouldBeNil)

		_, err = handler.ListObjects(getContext(), "juventus", ListOptions{})
		So(err, ShouldNotBeNil)

		err = settings.Set(SettingsObjectListMaxCount, value)
		So(err, ShouldBeNil)
	})
}

func TestHandler_ListObjects4(t *testing.T) {
	Convey("OBJECTS - LIST: cannot get a collection objects if user is not authenticated", t, func() {
		initDB()
		router := DefaultRouter()
		handler := router.GetHandler()

		c, err := handler.ListObjects(getContext(), "juventus", ListOptions{})
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
		initDB()
		router := DefaultRouter()
		handler := router.GetHandler()

		_, err := handler.ListObjects(getContext(), "some-collection-id", ListOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_ListObjects(t *testing.T) {
	Convey("OBJECTS - LIST: can list a collection objects if user is authenticated from a registered client application", t, func() {
		initDB()
		router := DefaultRouter()
		handler := router.GetHandler()

		userCtx := userContextInRegisteredClient(getContext(), "pirlo")
		c, err := handler.ListObjects(userCtx, "juventus", ListOptions{Offset: utime.Now()})
		So(err, ShouldBeNil)
		defer func() {
			_ = c.Close()
		}()
	})
}

func TestHandler_SearchObjects(t *testing.T) {
	Convey("OBJECTS - SEARCH: cannot search if one the following parameters is not provided: collection-id, query", t, func() {
		initDB()
		router := DefaultRouter()
		handler := router.GetHandler()

		_, err := handler.SearchObjects(getContext(), "", &se.SearchQuery{})
		So(err, ShouldNotBeNil)

		_, err = handler.SearchObjects(getContext(), "some-collection-id", nil)
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_DeleteObject1(t *testing.T) {
	Convey("OBJECTS - DELETE: cannot delete if one the followings parameters is not provided: collection-id, object-id", t, func() {
		initDB()
		router := DefaultRouter()
		handler := router.GetHandler()

		err := handler.DeleteObject(getContext(), "", "some-object")
		So(err, ShouldNotBeNil)

		err = handler.DeleteObject(getContext(), "juventus", "")
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_DeleteObject(t *testing.T) {
	Convey("OBJECTS - HEADER: can delete an object if user is admin or or the context satisfies one of DELETE rules if the object", t, func() {
		initDB()
		router := DefaultRouter()
		handler := router.GetHandler()

		psgCtx := userContextInRegisteredClient(getContext(), "pochettino")

		err := handler.DeleteObject(psgCtx, "paris-sg", "n10")
		So(err, ShouldBeNil)
	})
}

func TestHandler_DeleteCollection1(t *testing.T) {
	Convey("COLLECTION - DELETE: cannot delete a collection if id is not provided", t, func() {
		initDB()

		router := DefaultRouter()
		handler := router.GetHandler()

		// Retrieve new created collection from admin context
		adminContext := userContext(getContext(), "admin")
		err := handler.DeleteCollection(adminContext, "")
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_DeleteCollection2(t *testing.T) {
	Convey("COLLECTION - DELETE: cannot delete a collection if user is not admin", t, func() {
		initDB()

		router := DefaultRouter()
		handler := router.GetHandler()

		err := handler.DeleteCollection(getContext(), "juventus")
		So(err, ShouldNotBeNil)

		user1Context := userContext(getContext(), "pirlo")
		err = handler.DeleteCollection(user1Context, "juventus")
		So(err, ShouldNotBeNil)

		user1Context = userContextInRegisteredClient(getContext(), "pirlo")
		err = handler.DeleteCollection(user1Context, "juventus")
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_DeleteCollection(t *testing.T) {
	Convey("COLLECTION - DELETE: can collection if the user is admin", t, func() {
		initDB()

		router := DefaultRouter()
		handler := router.GetHandler()

		adminContext := userContext(getContext(), "admin")

		err := handler.DeleteCollection(adminContext, "juventus")
		So(err, ShouldBeNil)
		err = handler.DeleteCollection(adminContext, "barcelona")
		So(err, ShouldBeNil)
		err = handler.DeleteCollection(adminContext, "paris-sg")
		So(err, ShouldBeNil)

		// Retrieve all the created collection from admin context
		cols, err := handler.ListCollections(adminContext)
		So(err, ShouldBeNil)
		So(cols, ShouldHaveLength, 0)
	})
}
