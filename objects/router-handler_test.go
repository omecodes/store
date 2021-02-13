package objects

import (
	"context"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/omecodes/bome"
	"github.com/omecodes/store/auth"
	. "github.com/smartystreets/goconvey/convey"
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
							Rule:        "true",
						},
					},
					Write: []*auth.Permission{
						{
							Name:        "restricted-write",
							Label:       "Restricted Write",
							Description: "Only creator can write",
							Rule:        "user.name==o.header.creator",
						},
					},
					Delete: []*auth.Permission{
						{
							Name:        "restricted-delete",
							Label:       "Restricted Delete",
							Description: "Only creator can delete",
							Rule:        "user.name==o.header.creator",
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
							Rule:        "true",
						},
					},
					Write: []*auth.Permission{
						{
							Name:        "restricted-write",
							Label:       "Restricted Write",
							Description: "Only creator can write",
							Rule:        "user.name==o.header.creator",
						},
					},
					Delete: []*auth.Permission{
						{
							Name:        "restricted-delete",
							Label:       "Restricted Delete",
							Description: "Only creator can delete",
							Rule:        "user.name==o.header.creator",
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
							Rule:        "true",
						},
					},
					Write: []*auth.Permission{
						{
							Name:        "restricted-write",
							Label:       "Restricted Write",
							Description: "Only creator can write",
							Rule:        "user.name==o.header.creator",
						},
					},
					Delete: []*auth.Permission{
						{
							Name:        "restricted-delete",
							Label:       "Restricted Delete",
							Description: "Only creator can delete",
							Rule:        "user.name==o.header.creator",
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
	return auth.ContextWithAuh(ctx, &auth.User{Name: name})
}

func userContextInRegisteredClient(ctx context.Context, name string) context.Context {
	// this create a context as if it was created by the authentication interceptor when
	// receiving a request from a user by the means of a registered app client
	return auth.ContextWithAuh(ctx, &auth.User{Name: name, Access: "client"})
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

func TestHandler_CreateCollection(t *testing.T) {
	Convey("COLLECTION - CREATE: a collection MUST have at least an id AND default access security rules", t, func() {
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
								Rule:        "user.name==o.header.creator",
							},
						},
						Delete: []*auth.Permission{
							{
								Name:        "restricted-delete",
								Label:       "Restricted Delete",
								Description: "Only creator can delete",
								Rule:        "user.name==o.header.creator",
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
		So(err.Error(), ShouldContainSubstring, "bad input")

		col.Id = "objects"
		col.DefaultAccessSecurityRules = nil
		err = handler.CreateCollection(adminContext, col)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "bad input")

		// Try to create collection as unauthenticated user
		err = handler.CreateCollection(getContext(), col)
		So(err, ShouldNotBeNil)

		// Try to create collection as user1
		user1Context := userContext(getContext(), "user1")
		err = handler.CreateCollection(user1Context, col)
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_CreateCollection1(t *testing.T) {
	Convey("COLLECTION - CREATE: Only admin is allowed to create collection", t, func() {
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

func TestHandler_GetCollection(t *testing.T) {
	Convey("COLLECTION - GET: Only admin and user from within a verified client can get collection info", t, func() {
		initDB()

		router := DefaultRouter()
		handler := router.GetHandler()

		// Try to retrieve collection info from non authenticated user
		col, err := handler.GetCollection(getContext(), "objects")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "forbidden")
		So(col, ShouldBeNil)

		// Retrieve new created collection from user1 context
		user1Context := userContext(getContext(), "user1")
		col, err = handler.GetCollection(user1Context, "objects")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "forbidden")
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

func TestHandler_GetCollection1(t *testing.T) {
	Convey("COLLECTION - GET: id parameter is required to load collection", t, func() {
		initDB()
		router := DefaultRouter()
		handler := router.GetHandler()

		// Retrieve new created collection from admin context
		adminContext := userContext(getContext(), "admin")
		col, err := handler.GetCollection(adminContext, "")
		So(err, ShouldNotBeNil)
		So(col, ShouldBeNil)
		So(err.Error(), ShouldContainSubstring, "bad input")
	})
}

func TestHandler_PutObject1(t *testing.T) {
	Convey("OBJECTS CREATE: one cannot create objects without settings manager in context", t, func() {
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
	Convey("OBJECTS CREATE: one cannot create is collection one of the following is not specified: collection, header or object data", t, func() {
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
	Convey("OBJECTS CREATE: one could not create object if settings 'SettingsDataMaxSizePath' is not set or has value other than number", t, func() {
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
	Convey("OBJECTS CREATE: one could not create object with data size is greater than 'SettingsDataMaxSizePath' value", t, func() {
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

func TestHandler_PutObject(t *testing.T) {
	Convey("OBJECTS CREATE: authenticated user from verified clients can put objects", t, func() {
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
	Convey("OBJECTS PATCH: one cannot patch objects without settings manager in context", t, func() {
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
	Convey("OBJECTS CREATE: one cannot patch is collection one of the following is not specified: collection, patch or object data", t, func() {
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
	Convey("OBJECTS PATCH: one cannot patch object if settings 'SettingsDataMaxSizePath' is not set or has value other than number", t, func() {
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
	Convey("OBJECTS PATCH: one could not patch object with data size is greater than 'SettingsDataMaxSizePath' value", t, func() {
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
	Convey("OBJECTS CREATE: patch object", t, func() {
		initDB()

		router := DefaultRouter()
		handler := router.GetHandler()
		patch := &Patch{
			ObjectId: "m10",
			At:       "$.age",
			Data:     "30",
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

func TestHandler_MoveObject(t *testing.T) {
	Convey("", t, func() {

	})
}

func TestHandler_GetObject1(t *testing.T) {
	Convey("OBJECTS GET: one cannot get object if collection or target object id is not set", t, func() {
		initDB()
		router := DefaultRouter()
		handler := router.GetHandler()

		_, err := handler.GetObject(getContext(), "", "some-object", GetOptions{})
		So(err, ShouldNotBeNil)

		_, err = handler.GetObject(getContext(), "juventus", "", GetOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_GetObject(t *testing.T) {
	Convey("OBJECTS GET: one cannot get object if collection or target object id is not set", t, func() {
		initDB()
		router := DefaultRouter()
		handler := router.GetHandler()

		psgCtx := userContextInRegisteredClient(getContext(), "pochettino")

		object, err := handler.GetObject(psgCtx, "juventus", "cr7", GetOptions{})
		So(err, ShouldBeNil)
		So(object.Header.Id, ShouldEqual, "cr7")
	})
}

func TestHandler_DeleteCollection(t *testing.T) {
	Convey("COLLECTION - DELETE: Only admin is allowed to delete a collection", t, func() {
		initDB()

		router := DefaultRouter()
		handler := router.GetHandler()

		// Try to retrieve collection info from non authenticated user
		err := handler.DeleteCollection(getContext(), "objects")
		So(err, ShouldNotBeNil)

		// Retrieve new created collection from user1 context
		user1Context := userContext(getContext(), "user1")
		err = handler.DeleteCollection(user1Context, "objects")
		So(err, ShouldNotBeNil)

		// Retrieve new created collection from user1 context
		user1Context = userContextInRegisteredClient(getContext(), "user1")
		err = handler.DeleteCollection(user1Context, "objects")
		So(err, ShouldNotBeNil)

		// Retrieve new created collection from admin context
		adminContext := userContext(getContext(), "admin")

		err = handler.DeleteCollection(adminContext, "juventus")
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

func TestHandler_DeleteCollection1(t *testing.T) {
	Convey("COLLECTION - DELETE: requires the collection id", t, func() {
		initDB()

		router := DefaultRouter()
		handler := router.GetHandler()

		// Retrieve new created collection from admin context
		adminContext := userContext(getContext(), "admin")
		err := handler.DeleteCollection(adminContext, "")
		So(err, ShouldNotBeNil)
	})
}
