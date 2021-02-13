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

	collection = &Collection{
		Id:          "objects",
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
	return auth.ContextWithAuh(ctx, &auth.User{Name: name, Access: "client"})
}

func initDB() {
	if db == nil {
		conn, err := sql.Open("sqlite3", "test.db")
		So(err, ShouldBeNil)

		db, err = NewSqlDB(conn, bome.SQLite3, "store")
		So(err, ShouldBeNil)
		/*
			conn, err := sql.Open(bome.MySQL, "bome:bome@tcp(localhost:3306)/bome?charset=utf8")
			So(err, ShouldBeNil)
			db, err = NewSqlDB(conn, bome.MySQL, "store")
			So(err, ShouldBeNil)
		*/
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
	Convey("COLLECTION - CREATE: Only admin is allowed to create collection", t, func() {
		initDB()

		router := DefaultRouter()
		handler := router.GetHandler()

		// Try to create collection as unauthenticated user
		err := handler.CreateCollection(getContext(), collection)
		So(err, ShouldNotBeNil)

		// Try to create collection as user1
		user1Context := userContext(getContext(), "user1")
		err = handler.CreateCollection(user1Context, collection)
		So(err, ShouldNotBeNil)

		// Try to create collection as admin
		adminContext := userContext(getContext(), "admin")
		err = handler.CreateCollection(adminContext, collection)
		So(err, ShouldBeNil)

		// Retrieve new created collection from admin context
		col, err := handler.GetCollection(adminContext, "objects")
		So(err, ShouldBeNil)
		So(col.Id, ShouldEqual, "objects")
	})
}

func TestHandler_CreateCollection1(t *testing.T) {
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
		col, err = handler.GetCollection(user1Context, "objects")
		So(err, ShouldBeNil)
		So(col.Id, ShouldEqual, "objects")

		// Retrieve new created collection from admin context
		adminContext := userContext(getContext(), "admin")
		col, err = handler.GetCollection(adminContext, "objects")
		So(err, ShouldBeNil)
		So(col.Id, ShouldEqual, "objects")

		// Retrieve all the created collection from admin context
		cols, err := handler.ListCollections(adminContext)
		So(err, ShouldBeNil)
		So(cols, ShouldHaveLength, 1)
		So(cols[0].Id, ShouldEqual, "objects")
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

func TestHandler_PutObject(t *testing.T) {
	Convey("OBJECTS CREATE: Only admin and authenticated user from verified clients can put objects", t, func() {
		initDB()

		router := DefaultRouter()
		handler := router.GetHandler()
		data := `{"name": "user1", "age": 30, "city": "Paris"}`
		object := &Object{
			Header: &Header{
				Id:        "",
				CreatedBy: "user1",
				CreatedAt: time.Now().UnixNano(),
				Size:      int64(len(data)),
			},
			Data: data,
		}
		var err error
		// Retrieve new created collection from user1 context
		user1Context := userContextInRegisteredClient(getContext(), "user1")
		object.Header.Id, err = handler.PutObject(user1Context, "objects", object, nil, nil, PutOptions{})
		So(err, ShouldBeNil)
	})
}

func TestHandler_PutObject2(t *testing.T) {
	Convey("OBJECTS CREATE: Only authenticated users are allowed to create objects", t, func() {

	})
}

func TestHandler_PutObject3(t *testing.T) {
	Convey("OBJECTS CREATE: Only authenticated users are allowed to create objects", t, func() {

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
		So(err.Error(), ShouldContainSubstring, "forbidden")

		// Retrieve new created collection from user1 context
		user1Context := userContext(getContext(), "user1")
		err = handler.DeleteCollection(user1Context, "objects")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "forbidden")

		// Retrieve new created collection from user1 context
		user1Context = userContextInRegisteredClient(getContext(), "user1")
		err = handler.DeleteCollection(user1Context, "objects")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "forbidden")

		// Retrieve new created collection from admin context
		adminContext := userContext(getContext(), "admin")
		err = handler.DeleteCollection(adminContext, "objects")
		So(err, ShouldBeNil)

		// Retrieve all the created collection from admin context
		cols, err := handler.ListCollections(adminContext)
		So(err, ShouldBeNil)
		So(cols, ShouldHaveLength, 0)
	})
}
