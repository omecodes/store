package files

import (
	"context"
	"database/sql"
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
		sourceManager, err = NewSourceSQLManager(db, bome.SQLite3, "sources")
		So(err, ShouldBeNil)
	}
}

func initDir() {
	if workingDir == "" {
		workingDir = "./.test-work"
		var err error
		workingDir, err = filepath.Abs(workingDir)
		So(err, ShouldBeNil)
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
			URI:         "files://" + workingDir,
			ExpireTime:  -1,
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
