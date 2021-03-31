package files

import (
	"bytes"
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

func clearDir() {
	if workingDir != "" {
		workingDir = "./.test-work"
		var err error
		workingDir, err = filepath.Abs(workingDir)
		So(err, ShouldBeNil)

		_ = os.RemoveAll(workingDir)
	}
}

func getContext() context.Context {
	ctx := context.Background()
	return ContextWithSourceManager(ctx, sourceManager)
}

func getContextWithoutSourceManager() context.Context {
	return context.Background()
}

func getContextWithUser(user string) context.Context {
	ctx := getContext()
	ctx = auth.ContextWithUser(ctx, &auth.User{Name: user})
	return ctx
}

func contextWithApp(parent context.Context, name string, clientType auth.ClientType) context.Context {
	return auth.ContextWithApp(parent, &auth.ClientApp{
		Key:    name,
		Secret: "",
		Type:   clientType,
	})
}

func adminContext() context.Context {
	return getContextWithUser("admin")
}

func getContextWithUserFromClientAndNoSourceManager(user string) context.Context {
	return auth.ContextWithUser(getContextWithoutSourceManager(), &auth.User{Name: user})
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
		userContext := getContextWithUser("user")
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
		userContext := getContextWithUser("admin")
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

		source, err := handler.GetSource(getContextWithUser("ome"), "main")
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

		userContext := getContextWithUser("ome")
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
	Convey("FILES - MKDIR: cannot create a directory if context has no source manager", t, func() {
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

		sources, err = handler.ListSources(getContextWithUser("user1"))
		So(err, ShouldBeNil)
		So(sources, ShouldHaveLength, 1)
	})
}

func TestHandler_CreateDir6(t *testing.T) {
	Convey("FILES - CREATE DIR: can create dir in source if context user is admin or satisfies WRITE permission ", t, func() {
		initDir()
		initDir()

		router := DefaultFilesRouter()
		handler := router.GetHandler()

		user1Context := getContextWithUser("user1")
		err := handler.CreateDir(user1Context, "user1-source", "Documents")
		So(err, ShouldBeNil)
	})
}

func TestHandler_CreateDir7(t *testing.T) {
	Convey("FILES - CREATE DIR: can create dir in source if context user is admin or satisfies WRITE permission ", t, func() {
		initDir()
		initDir()

		router := DefaultFilesRouter()
		handler := router.GetHandler()

		user1Context := getContextWithUser("user1")
		err := handler.CreateDir(user1Context, "user1-source", "Documents/photo")
		So(err, ShouldBeNil)
	})
}

func TestHandler_CreateSource8(t *testing.T) {
	Convey("SOURCE - CREATE: can create source (share) for another user", t, func() {
		initDB()
		initDir()

		router := DefaultFilesRouter()
		handler := router.GetHandler()

		user1Source := &Source{
			ID:          "user2-source",
			Label:       "User2 Files",
			Description: "",
			CreatedBy:   "user1",
			Type:        TypeReference,
			URI:         SchemeSource + "://user1-source/Documents/photo",
			PermissionOverrides: &Permissions{
				Filename: "/Documents/photo",
				Read: []*auth.Permission{
					{
						Name:         "user2-can-read",
						Label:        "User2 can read",
						Description:  "User2 has permission to read all file in this source",
						Rule:         "user.name=='user2'",
						RelatedUsers: []string{"user2"},
					},
				},
				Write: []*auth.Permission{
					{
						Name:         "user2-write-perm",
						Label:        "User2 write permission",
						Description:  "User2 has permission to read all file in this source",
						Rule:         "user.name=='user2'",
						RelatedUsers: []string{"user2"},
					},
				},
				Chmod: []*auth.Permission{
					{
						Name:         "user2-chmod-perm",
						Label:        "User2 chmod permission",
						Description:  "User2 cannot chmod files in this source",
						Rule:         "false",
						RelatedUsers: []string{"public"},
					},
				},
			},
			ExpireTime: -1,
		}

		err := handler.CreateSource(adminContext(), user1Source)
		So(err, ShouldBeNil)
	})
}

func TestHandler_WriteFileContent1(t *testing.T) {
	Convey("FILES - WRITE: cannot write file if one of the following parameters is not set: sourceID, filename, content, size", t, func() {
		initDB()
		initDir()

		handler := DefaultFilesRouter().GetHandler()
		err := handler.WriteFileContent(getContext(), "", "filename", bytes.NewBufferString("a"), 1, WriteOptions{})
		So(err, ShouldNotBeNil)

		err = handler.WriteFileContent(getContext(), "main", "", bytes.NewBufferString("a"), 1, WriteOptions{})
		So(err, ShouldNotBeNil)

		err = handler.WriteFileContent(getContext(), "main", "filename", nil, 1, WriteOptions{})
		So(err, ShouldNotBeNil)

		err = handler.WriteFileContent(getContext(), "main", "filename", bytes.NewBufferString("a"), 0, WriteOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_WriteFileContent2(t *testing.T) {
	Convey("FILES - WRITE: cannot write file content if context has no user", t, func() {
		initDB()
		initDir()

		handler := DefaultFilesRouter().GetHandler()

		err := handler.WriteFileContent(getContext(), "main", "filename", bytes.NewBufferString("a"), 1, WriteOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_WriteFileContent3(t *testing.T) {
	Convey("FILES - WRITE: cannot write file content if context has user that has no write permission on target folder", t, func() {
		initDB()
		initDir()

		handler := DefaultFilesRouter().GetHandler()

		userContext := getContextWithUser("user1")
		err := handler.WriteFileContent(userContext, "main", "filename", bytes.NewBufferString("a"), 1, WriteOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_WriteFileContent4(t *testing.T) {
	Convey("FILES - WRITE: cannot write file content if context has no source manager", t, func() {
		initDB()
		initDir()

		handler := DefaultFilesRouter().GetHandler()

		user1Source := getContextWithUserFromClientAndNoSourceManager("admin")
		err := handler.WriteFileContent(user1Source, "user1-source", "file.txt", bytes.NewBufferString("a"), 1, WriteOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_WriteFileContent5(t *testing.T) {
	Convey("FILES - WRITE: can write file content if context has source manager and context has user with WRITE permission on target folder", t, func() {
		initDB()
		initDir()

		handler := DefaultFilesRouter().GetHandler()

		err := handler.WriteFileContent(adminContext(), "main", "file.txt", bytes.NewBufferString("a"), 1, WriteOptions{})
		So(err, ShouldBeNil)

		err = handler.WriteFileContent(getContextWithUser("user1"), "user1-source", "file.txt", bytes.NewBufferString("a"), 1, WriteOptions{})
		So(err, ShouldBeNil)
	})
}

func TestHandler_ListSource3(t *testing.T) {
	Convey("SOURCES LIST: can list sources which one of the READ rule is satisfied by the context user", t, func() {
		initDB()
		initDir()

		router := DefaultFilesRouter()
		handler := router.GetHandler()

		sources, err := handler.ListSources(adminContext())
		So(err, ShouldBeNil)
		So(sources, ShouldHaveLength, 1)

		sources, err = handler.ListSources(getContextWithUser("user1"))
		So(err, ShouldBeNil)
		So(sources, ShouldHaveLength, 1)
	})
}

func TestHandler_ListDir1(t *testing.T) {
	Convey("FILES - LS: cannot list directory if one of the following parameters is not set", t, func() {
		initDB()
		initDir()

		handler := DefaultFilesRouter().GetHandler()
		_, err := handler.ListDir(getContext(), "", "/", ListDirOptions{})
		So(err, ShouldNotBeNil)

		_, err = handler.ListDir(getContext(), "user-source1", "", ListDirOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_ListDir2(t *testing.T) {
	Convey("FILES - LS: cannot list directory if context has no source manager", t, func() {
		initDB()
		initDir()

		handler := DefaultFilesRouter().GetHandler()
		_, err := handler.ListDir(getContextWithUserFromClientAndNoSourceManager("user1"), "user1-source", "/", ListDirOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_ListDir3(t *testing.T) {
	Convey("FILES - LS: cannot list directory if context user has no READ permission on it", t, func() {
		initDB()
		initDir()

		handler := DefaultFilesRouter().GetHandler()
		_, err := handler.ListDir(getContextWithUser("user1"), "main", "/", ListDirOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_ListDir4(t *testing.T) {
	Convey("FILES - LS: can list directory if context user has READ permission on it", t, func() {
		initDB()
		initDir()

		handler := DefaultFilesRouter().GetHandler()
		_, err := handler.ListDir(getContextWithUser("user1"), "user1-source", "/", ListDirOptions{})
		So(err, ShouldBeNil)
	})
}

func TestHandler_DeleteSource1(t *testing.T) {
	Convey("SOURCE - DELETE: cannot delete source if one the following parameters is not provided: sourceID", t, func() {
		initDB()
		initDir()

		handler := DefaultFilesRouter().GetHandler()
		err := handler.DeleteSource(getContext(), "")
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_DeleteSource2(t *testing.T) {
	Convey("SOURCE - DELETE: cannot delete source if the context has no source manager", t, func() {
		initDB()
		initDir()

		handler := DefaultFilesRouter().GetHandler()
		adminContext := getContextWithUserFromClientAndNoSourceManager("admin")
		err := handler.DeleteSource(adminContext, "user1-source")
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_DeleteSource3(t *testing.T) {
	Convey("SOURCE - DELETE: cannot delete source if the context has no user", t, func() {
		initDB()
		initDir()

		handler := DefaultFilesRouter().GetHandler()
		err := handler.DeleteSource(getContext(), "main")
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_DeleteSource4(t *testing.T) {
	Convey("SOURCE - DELETE: cannot delete source if context user is not admin or the source is created by another user", t, func() {
		initDB()
		initDir()

		handler := DefaultFilesRouter().GetHandler()
		err := handler.DeleteSource(getContextWithUser("user-1"), "main")
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_DeleteSource5(t *testing.T) {
	Convey("SOURCE - DELETE: can delete a source if it has been created by the context user", t, func() {
		initDB()
		initDir()

		handler := DefaultFilesRouter().GetHandler()
		err := handler.DeleteSource(adminContext(), "user1-source")
		So(err, ShouldBeNil)
	})
}

func TestHandler_DeleteSource6(t *testing.T) {
	Convey("SOURCE - DELETE: cannot delete non existing source", t, func() {
		initDB()
		initDir()

		handler := DefaultFilesRouter().GetHandler()
		err := handler.DeleteSource(adminContext(), "source")
		So(err, ShouldNotBeNil)
	})
}

/*func TestHandler_Clean(t *testing.T) {
	Convey("", t, func() {
		clearDir()
	})
} */