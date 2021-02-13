package objects

import (
	"context"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/omecodes/bome"
	"github.com/omecodes/store/auth"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
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

func userContext(ctx context.Context, name string) context.Context {
	return auth.ContextWithAuh(ctx, &auth.User{Name: name})
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

func Test_CreateCollection(t *testing.T) {
	Convey("Only admin is allowed to create collection", t, func() {
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
	})
}
