package files

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/omecodes/bome"
	"github.com/omecodes/store/auth"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	workingDir    string
	sourceManager SourceManager
	mainSource    *Source
)

func initDB() {
	if sourceManager == nil {
		db, err := sql.Open(bome.SQLite3, ":memory:")
		So(err, ShouldBeNil)
		sourceManager, err = NewSourceSQLManager(db, bome.SQLite3, "store")
		So(err, ShouldBeNil)
	}
}

func initDir() {
	if workingDir == "" {
		workingDir = "./.test-work"
		var err error
		workingDir, err = filepath.Abs(workingDir)
		So(err, ShouldBeNil)

		err = os.MkdirAll(workingDir, os.ModePerm)
		So(err == nil || os.IsExist(err), ShouldBeTrue)
	}
}

func getContext() context.Context {
	ctx := context.Background()
	return ContextWithSourceManager(ctx, sourceManager)
}

func getContextWithoutSourceManager() context.Context {
	return context.Background()
}

func getContextWithUserFromClient(user string) context.Context {
	ctx := getContext()
	ctx = auth.ContextWithAuh(ctx, &auth.User{Name: user, Access: "client"})
	return ctx
}

func adminContext() context.Context {
	return getContextWithUserFromClient("admin")
}

func getContextWithUserFromClientAndNoSourceManager(user string) context.Context {
	return auth.ContextWithAuh(getContextWithoutSourceManager(), &auth.User{Name: user, Access: "client"})
}

func Test_initializeDatabase(t *testing.T) {
	Convey("DATABASE: initialization", t, func() {
		initDB()
	})
}

func TestHandler_CreateSource1(t *testing.T) {
	Convey("SOURCE - CREATE: cannot create source if one the following parameters is not provided: source", t, func() {
		initDB()
		router := DefaultFilesRouter()
		handler := router.GetHandler()

		err := handler.CreateSource(getContext(), nil)
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_CreateSource2(t *testing.T) {
	Convey("SOURCE - CREATE: cannot create source if one the following parameters is not provided: type, uri", t, func() {
		initDB()
		router := DefaultFilesRouter()
		handler := router.GetHandler()

		source := &Source{
			ID:          "source",
			Label:       "Source de tests",
			Description: "Source de tests",
			Type:        0,
			URI:         "",
			ExpireTime:  -1,
		}
		err := handler.CreateSource(getContext(), source)
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_CreateSource3(t *testing.T) {
	Convey("SOURCE - CREATE: cannot create source if context has no authenticated user", t, func() {
		initDB()
		initDir()

		router := DefaultFilesRouter()
		handler := router.GetHandler()

		mainSource = &Source{
			ID:          "main",
			Label:       "Root source",
			Description: "Root source",
			CreatedBy:   "admin",
			Type:        TypeDisk,
			URI:         "files://" + workingDir,
			ExpireTime:  -1,
		}
		err := handler.CreateSource(getContext(), mainSource)
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_CreateSource4(t *testing.T) {
	Convey("SOURCE - CREATE: cannot create root source if the context user is not admin", t, func() {
		initDB()
		initDir()

		router := DefaultFilesRouter()
		handler := router.GetHandler()

		mainSource = &Source{
			ID:          "main",
			Label:       "Root source",
			Description: "Root source",
			Type:        TypeDisk,
			URI:         "files://" + workingDir,
			ExpireTime:  -1,
		}
		userContext := getContextWithUserFromClient("user")
		err := handler.CreateSource(userContext, mainSource)
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_CreateSource5(t *testing.T) {
	Convey("SOURCE - CREATE: can create root source if the context user is admin", t, func() {
		initDB()
		initDir()

		router := DefaultFilesRouter()
		handler := router.GetHandler()

		mainSource = &Source{
			ID:          "main",
			Label:       "Root source",
			Description: "Root source",
			Type:        TypeDisk,
			URI:         SchemeFS + "://" + workingDir,
			ExpireTime:  -1,
			PermissionOverrides: &Permissions{
				Filename: "/user1",
				Read: []*auth.Permission{
					{
						Name:         "admin-can-read",
						Label:        "Admin can read",
						Description:  "Admin has permission to read all file in this source",
						Rule:         "user.name=='admin'",
						RelatedUsers: []string{"admin"},
					},
				},
				Write: []*auth.Permission{
					{
						Name:         "admin-write-perm",
						Label:        "Admin write permission",
						Description:  "Admin has permission to read all file in this source",
						Rule:         "user.name=='admin'",
						RelatedUsers: []string{"admin"},
					},
				},
				Chmod: []*auth.Permission{
					{
						Name:         "admin-chmod-perm",
						Label:        "admin chmod permission",
						Description:  "admin has permission to chmod all file in this source",
						Rule:         "user.name=='admin'",
						RelatedUsers: []string{"admin"},
					},
				},
			},
		}
		userContext := getContextWithUserFromClient("admin")
		err := handler.CreateSource(userContext, mainSource)
		So(err, ShouldBeNil)
	})
}

func TestHandler_CreateSource6(t *testing.T) {
	Convey("SOURCE - CREATE: cannot create source context has no source manager", t, func() {
		initDB()
		initDir()

		router := DefaultFilesRouter()
		handler := router.GetHandler()

		mainSource = &Source{
			ID:          "main",
			Label:       "Root source",
			Description: "Root source",
			Type:        TypeDisk,
			URI:         "files://" + workingDir,
			ExpireTime:  -1,
		}

		userContext := getContextWithUserFromClientAndNoSourceManager("admin")
		err := handler.CreateSource(userContext, mainSource)
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_GetSource1(t *testing.T) {
	Convey("SOURCE - GET: cannot get source if id is not provided", t, func() {
		initDir()
		initDir()

		router := DefaultFilesRouter()
		handler := router.GetHandler()

		source, err := handler.GetSource(getContext(), "")
		So(err, ShouldNotBeNil)
		So(source, ShouldBeNil)
	})
}

func TestHandler_GetSource2(t *testing.T) {
	Convey("SOURCE - GET: cannot get source if context has no authenticated user", t, func() {
		initDir()
		initDir()

		router := DefaultFilesRouter()
		handler := router.GetHandler()

		source, err := handler.GetSource(getContext(), "main")
		So(err, ShouldNotBeNil)
		So(source, ShouldBeNil)
	})
}

func TestHandler_GetSource3(t *testing.T) {
	Convey("SOURCE - GET: cannot get the main source if the context user is not admin", t, func() {
		initDir()
		initDir()

		router := DefaultFilesRouter()
		handler := router.GetHandler()

		source, err := handler.GetSource(getContextWithUserFromClient("ome"), "main")
		So(err, ShouldNotBeNil)
		So(source, ShouldBeNil)
	})
}

func TestHandler_GetSource4(t *testing.T) {
	Convey("SOURCE - GET: can get the main source if the context user is admin", t, func() {
		initDir()
		initDir()

		router := DefaultFilesRouter()
		handler := router.GetHandler()

		source, err := handler.GetSource(adminContext(), "main")
		So(err, ShouldBeNil)
		So(source.ID, ShouldEqual, "main")
	})
}

func TestHandler_CreateDir1(t *testing.T) {
	Convey("FILES - MKDIR: cannot create a directory if one the following parameters is not set: sourceID, filename", t, func() {
		initDir()
		initDir()

		router := DefaultFilesRouter()
		handler := router.GetHandler()

		err := handler.CreateDir(getContext(), "", "user1")
		So(err, ShouldNotBeNil)

		err = handler.CreateDir(getContext(), "main", "")
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_CreateDir2(t *testing.T) {
	Convey("FILES - MKDIR: cannot create a directory in a restricted source if context has no user", t, func() {
		initDir()
		initDir()

		router := DefaultFilesRouter()
		handler := router.GetHandler()

		err := handler.CreateDir(getContext(), "main", "user1")
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_CreateDir3(t *testing.T) {
	Convey("FILES - MKDIR: cannot create a directory if context user has no access to target source", t, func() {
		initDir()
		initDir()

		router := DefaultFilesRouter()
		handler := router.GetHandler()

		userContext := getContextWithUserFromClient("ome")
		err := handler.CreateDir(userContext, "main", "user1")
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_CreateDir4(t *testing.T) {
	Convey("FILES - MKDIR: can create a directory if context user is admin or has has rights permissions in parent", t, func() {
		initDir()
		initDir()

		router := DefaultFilesRouter()
		handler := router.GetHandler()

		err := handler.CreateDir(adminContext(), "main", "user1")
		So(err, ShouldBeNil)
	})
}

func TestHandler_CreateDir5(t *testing.T) {
	Convey("FILES - MKDIR: can create a directory if context has no source manager", t, func() {
		initDir()
		initDir()

		router := DefaultFilesRouter()
		handler := router.GetHandler()

		adminContext := getContextWithUserFromClientAndNoSourceManager("admin")
		err := handler.CreateDir(adminContext, "main", "user1")
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_CreateSource7(t *testing.T) {
	Convey("SOURCE - CREATE: can create source (share) for another user", t, func() {
		initDB()
		initDir()

		router := DefaultFilesRouter()
		handler := router.GetHandler()

		user1Source := &Source{
			ID:          "user1-source",
			Label:       "User1 Files",
			Description: "",
			CreatedBy:   "admin",
			Type:        TypeReference,
			URI:         SchemeSource + "://main/user1",
			PermissionOverrides: &Permissions{
				Filename: "/user1",
				Read: []*auth.Permission{
					{
						Name:         "user1-can-read",
						Label:        "User1 can read",
						Description:  "User1 has permission to read all file in this source",
						Rule:         "user.name=='user1'",
						RelatedUsers: []string{"user1"},
					},
				},
				Write: []*auth.Permission{
					{
						Name:         "user1-write-perm",
						Label:        "User1 write permission",
						Description:  "User1 has permission to read all file in this source",
						Rule:         "user.name=='user1'",
						RelatedUsers: []string{"user1"},
					},
				},
				Chmod: []*auth.Permission{
					{
						Name:         "user1-chmod-perm",
						Label:        "User1 chmod permission",
						Description:  "User1 has permission to chmod all file in this source",
						Rule:         "user.name=='user1'",
						RelatedUsers: []string{"user1"},
					},
				},
			},
			ExpireTime: -1,
		}

		err := handler.CreateSource(adminContext(), user1Source)
		So(err, ShouldBeNil)
	})
}

func TestHandler_ListSource1(t *testing.T) {
	Convey("SOURCES LIST: cannot list source if context has no source manager", t, func() {
		initDB()
		initDir()

		router := DefaultFilesRouter()
		handler := router.GetHandler()

		sources, err := handler.ListSources(getContextWithUserFromClientAndNoSourceManager("admin"))
		So(err, ShouldNotBeNil)
		So(sources, ShouldBeNil)
	})
}

func TestHandler_ListSource2(t *testing.T) {
	Convey("SOURCES LIST: can list sources which one of the READ rule is satisfied by the context user", t, func() {
		initDB()
		initDir()

		router := DefaultFilesRouter()
		handler := router.GetHandler()

		sources, err := handler.ListSources(adminContext())
		So(err, ShouldBeNil)
		So(sources, ShouldHaveLength, 1)

		sources, err = handler.ListSources(getContextWithUserFromClient("user1"))
		So(err, ShouldBeNil)
		So(sources, ShouldHaveLength, 1)
	})
}
