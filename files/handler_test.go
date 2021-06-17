package files

import (
	"bytes"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/omecodes/bome"
	"github.com/omecodes/store/acl"
	"github.com/omecodes/store/auth"
	pb "github.com/omecodes/store/gen/go/proto"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	clientApp     *pb.ClientApp
	adminApp      *pb.ClientApp
	workingDir    string
	accessManager AccessManager
	tupleStore    acl.TupleStore
	nsConfigStore acl.NamespaceConfigStore
)

func setupWorkingDir() {
	if workingDir == "" {
		workingDir = "./.test-work"
		var err error
		workingDir, err = filepath.Abs(workingDir)
		So(err, ShouldBeNil)

		err = os.MkdirAll(workingDir, os.ModePerm)
		So(err == nil || os.IsExist(err), ShouldBeTrue)
	}
}

func setupDatabases() {
	if accessManager == nil {
		db, err := sql.Open(bome.SQLite3, ":memory:")
		So(err, ShouldBeNil)
		accessManager, err = NewAccessSQLManager(db, bome.SQLite3, "store")
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
	err := nsConfigStore.SaveNamespace(&groupNamespaceConfig)
	So(err, ShouldBeNil)

	err = nsConfigStore.SaveNamespace(&accessNamespaceConfig)
	So(err, ShouldBeNil)

	err = nsConfigStore.SaveNamespace(&fileNamespaceConfig)
	So(err, ShouldBeNil)

	err = tupleStore.Save(context.Background(), &pb.DBEntry{
		Sid:      1,
		Object:   adminsGroup,
		Relation: relationMember,
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
	setupWorkingDir()
}

func tearDown() {
	_ = os.Remove("namespaces.db")
	_ = os.Remove("tuples.db")
	if workingDir != "" {
		workingDir = "./.test-work"
		var err error
		workingDir, err = filepath.Abs(workingDir)
		So(err, ShouldBeNil)
		_ = os.RemoveAll(workingDir)
	}
}

func baseContext() context.Context {
	ctx := context.Background()
	ctx = ContextWithAccessManager(ctx, accessManager)
	ctx = acl.ContextWithTupleStore(ctx, tupleStore)
	ctx = acl.ContextWithNamespaceConfigStore(ctx, nsConfigStore)
	ctx = acl.ContextWithManager(ctx, &acl.DefaultManager{})
	return ctx
}

func getContextWithoutAccessManager() context.Context {
	return context.Background()
}

func getContextWithUser(user string) context.Context {
	ctx := baseContext()
	ctx = auth.ContextWithUser(ctx, &pb.User{Name: user})
	return ctx
}

func adminContext() context.Context {
	return getContextWithUser("admin")
}

func getContextWithUserFromClientAndNoAccessManager(user string) context.Context {
	return auth.ContextWithUser(getContextWithoutAccessManager(), &pb.User{Name: user})
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

func Test_initializeDatabase(t *testing.T) {
	Convey("DATABASE: initialization", t, func() {
		tearDown()
		setup()
	})
}

func TestHandler_CreateAccess1(t *testing.T) {
	Convey("ACCESS - CREATE: cannot create access if one the following parameters is not provided: access", t, func() {
		setup()

		err := CreateAccess(baseContext(), nil, CreateAccessOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_CreateAccess2(t *testing.T) {
	Convey("ACCESS - CREATE: cannot create access if one the following parameters is not provided: type, uri", t, func() {
		setup()

		access := &pb.FSAccess{
			Id:          "access",
			Label:       "Access de tests",
			Description: "Access de tests",
			Type:        0,
			Uri:         "",
			ExpireTime:  -1,
		}
		err := CreateAccess(baseContext(), access, CreateAccessOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_CreateAccess3(t *testing.T) {
	Convey("ACCESS - CREATE: cannot create access if context has no authenticated user", t, func() {
		setup()

		access := &pb.FSAccess{
			Id:          "main",
			Label:       "Root access",
			Description: "Root access",
			CreatedBy:   "admin",
			Type:        pb.AccessType_Default,
			Uri:         "files://" + workingDir,
			ExpireTime:  -1,
		}
		err := CreateAccess(baseContext(), access, CreateAccessOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_CreateAccess4(t *testing.T) {
	Convey("ACCESS - CREATE: cannot create root access if the context user is not admin", t, func() {
		setup()

		access := &pb.FSAccess{
			Id:          "main",
			Label:       "Root access",
			Description: "Root access",
			Type:        pb.AccessType_Default,
			Uri:         "files://" + workingDir,
			ExpireTime:  -1,
		}

		err := CreateAccess(getContextWithUser("user"), access, CreateAccessOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_CreateAccess5(t *testing.T) {
	Convey("ACCESS - CREATE: can create root access if the context user is admin", t, func() {
		setup()
		access := &pb.FSAccess{
			Id:          "main",
			Label:       "Root access",
			Description: "Root access",
			CreatedBy:   "",
			Type:        pb.AccessType_Default,
			Uri:         SchemeFS + "://" + workingDir,
			Encryption:  nil,
			IsFolder:    true,
			ExpireTime:  -1,
			EncodedInfo: "",
		}
		err := CreateAccess(fullAdminContext(), access, CreateAccessOptions{})
		So(err, ShouldBeNil)
	})
}

func TestHandler_CreateAccess6(t *testing.T) {
	Convey("ACCESS - CREATE: cannot create FS access context has no access manager", t, func() {
		setup()

		access := &pb.FSAccess{
			Id:          "main",
			Label:       "Root access",
			Description: "Root access",
			Type:        pb.AccessType_Default,
			Uri:         "files://" + workingDir,
			ExpireTime:  -1,
		}

		userContext := getContextWithUserFromClientAndNoAccessManager("admin")
		err := CreateAccess(userContext, access, CreateAccessOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_GetAccess1(t *testing.T) {
	Convey("ACCESS - GET: cannot get FS access if id is not provided", t, func() {
		setup()

		access, err := GetAccess(baseContext(), "", GetAccessOptions{})
		So(err, ShouldNotBeNil)
		So(access, ShouldBeNil)
	})
}

func TestHandler_GetAccess2(t *testing.T) {
	Convey("ACCESS - GET: cannot get FS access if context has no authenticated user", t, func() {
		setup()

		access, err := GetAccess(baseContext(), "main", GetAccessOptions{})
		So(err, ShouldNotBeNil)
		So(access, ShouldBeNil)
	})
}

func TestHandler_GetAccess3(t *testing.T) {
	Convey("ACCESS - GET: cannot get the main access if the context user is not admin", t, func() {
		setup()

		access, err := GetAccess(getContextWithUser("ome"), "main", GetAccessOptions{})
		So(err, ShouldNotBeNil)
		So(access, ShouldBeNil)
	})
}

func TestHandler_GetAccess4(t *testing.T) {
	Convey("ACCESS - GET: can get the main access if the context user is admin", t, func() {
		setup()
		access, err := GetAccess(fullAdminContext(), "main", GetAccessOptions{})
		So(err, ShouldBeNil)
		So(access.Id, ShouldEqual, "main")
	})
}

func TestHandler_CreateDir1(t *testing.T) {
	Convey("FILES - MKDIR: cannot create a directory if one the following parameters is not set: accessID, filename", t, func() {
		setup()

		err := CreateDir(baseContext(), "", "user1", CreateDirOptions{})
		So(err, ShouldNotBeNil)

		err = CreateDir(baseContext(), "main", "", CreateDirOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_CreateDir2(t *testing.T) {
	Convey("FILES - MKDIR: cannot create a directory in a restricted access if context has no user", t, func() {
		setup()

		err := CreateDir(baseContext(), "main", "user1", CreateDirOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_CreateDir3(t *testing.T) {
	Convey("FILES - MKDIR: cannot create a directory if context user has no access to target access", t, func() {
		setup()

		userContext := getContextWithUser("ome")
		err := CreateDir(userContext, "main", "user1", CreateDirOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_CreateDir4(t *testing.T) {
	Convey("FILES - MKDIR: can create a directory if context user is admin or has has rights permissions in parent", t, func() {
		setup()
		err := CreateDir(adminContext(), "main", "user1", CreateDirOptions{})
		So(err, ShouldBeNil)
	})
}

func TestHandler_CreateDir5(t *testing.T) {
	Convey("FILES - MKDIR: cannot create a directory if context has no access manager", t, func() {
		setup()

		ctx := getContextWithUserFromClientAndNoAccessManager("admin")
		err := CreateDir(ctx, "main", "user1", CreateDirOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_CreateAccess7(t *testing.T) {
	Convey("ACCESS - CREATE: can create access (share) for another user", t, func() {
		setup()

		user1Access := &pb.FSAccess{
			Id:          "user1-access",
			Label:       "User1 Files",
			Description: "User1 FS access",
			CreatedBy:   "admin",
			Type:        pb.AccessType_Reference,
			Uri:         SchemeRef + "://main/user1",
			ExpireTime:  -1,
			IsFolder:    true,
		}

		ctx := fullAdminContext()
		err := CreateAccess(ctx, user1Access, CreateAccessOptions{})
		So(err, ShouldBeNil)

		err = acl.SaveACL(ctx, &pb.ACL{
			Object:   accessNamespace + ":user1-access",
			Relation: relationViewer,
			Subject:  "user1",
		}, acl.SaveACLOptions{})
		So(err, ShouldBeNil)

		err = acl.SaveACL(ctx, &pb.ACL{
			Object:   fileNamespace + ":user1-access",
			Relation: relationEditor,
			Subject:  "user1",
		}, acl.SaveACLOptions{})
		So(err, ShouldBeNil)

		err = acl.SaveACL(ctx, &pb.ACL{
			Object:   fileNamespace + ":user1-access",
			Relation: relationSharer,
			Subject:  "user1",
		}, acl.SaveACLOptions{})
		So(err, ShouldBeNil)
	})
}

func TestHandler_ListAccess1(t *testing.T) {
	Convey("ACCESS LIST: cannot list access if context has no access manager", t, func() {
		setup()

		accessList, err := GetAccessList(getContextWithUserFromClientAndNoAccessManager("admin"), GetAccessListOptions{})
		So(err, ShouldNotBeNil)
		So(accessList, ShouldBeNil)
	})
}

func TestHandler_ListAccess2(t *testing.T) {
	Convey("ACCESS LIST: can list access which one of the READ rule is satisfied by the context user", t, func() {
		/*setup()

		accessList, err := GetAccessList(fullAdminContext(), GetAccessListOptions{})
		So(err, ShouldBeNil)
		So(accessList, ShouldHaveLength, 1)

		accessList, err = GetAccessList(getContextWithUser("user1"), GetAccessListOptions{})
		So(err, ShouldBeNil)
		So(accessList, ShouldHaveLength, 1) */
	})
}

func TestHandler_CreateDir6(t *testing.T) {
	Convey("FILES - CREATE DIR: can create dir in access if context user is admin or satisfies WRITE permission ", t, func() {
		setup()

		user1Context := getContextWithUser("user1")
		err := CreateDir(user1Context, "user1-access", "Documents", CreateDirOptions{})
		So(err, ShouldBeNil)
	})
}

func TestHandler_CreateDir7(t *testing.T) {
	Convey("FILES - CREATE DIR: can create dir in access if context user is admin or satisfies WRITE permission ", t, func() {
		setup()

		user1Context := getContextWithUser("user1")
		err := CreateDir(user1Context, "user1-access", "Documents/photo", CreateDirOptions{})
		So(err, ShouldBeNil)
	})
}

func TestHandler_CreateAccess8(t *testing.T) {
	Convey("ACCESS - CREATE: can create access (share) for another user", t, func() {
		setup()

		user2Access := &pb.FSAccess{
			Id:          "user2-access",
			Label:       "User2 Files",
			Description: "",
			CreatedBy:   "user1",
			Type:        pb.AccessType_Reference,
			Uri:         SchemeRef + "://user1-access/Documents/photo",
			ExpireTime:  -1,
			IsFolder:    true,
		}

		ctx := userContext(clientAppContext(baseContext()), "user1")

		err := CreateAccess(ctx, user2Access, CreateAccessOptions{})
		So(err, ShouldBeNil)

		// access relation
		err = acl.SaveACL(ctx, &pb.ACL{
			Object:   accessNamespace + ":user2-access",
			Relation: relationViewer,
			Subject:  "user2",
		}, acl.SaveACLOptions{})
		So(err, ShouldBeNil)

		// file relations
		err = acl.SaveACL(ctx, &pb.ACL{
			Object:   fileNamespace + ":user2-access",
			Relation: relationEditor,
			Subject:  "user2",
		}, acl.SaveACLOptions{})
		So(err, ShouldBeNil)

		err = acl.SaveACL(ctx, &pb.ACL{
			Object:   fileNamespace + ":user1-access",
			Relation: relationSharer,
			Subject:  "user2",
		}, acl.SaveACLOptions{})
		So(err, ShouldBeNil)
	})
}

func TestHandler_WriteFileContent1(t *testing.T) {
	Convey("FILES - WRITE: cannot write file if one of the following parameters is not set: accessID, filename, content, size", t, func() {
		setup()

		err := WriteFileContent(baseContext(), "", "filename", bytes.NewBufferString("a"), 1, WriteOptions{})
		So(err, ShouldNotBeNil)

		err = WriteFileContent(baseContext(), "main", "", bytes.NewBufferString("a"), 1, WriteOptions{})
		So(err, ShouldNotBeNil)

		err = WriteFileContent(baseContext(), "main", "filename", nil, 1, WriteOptions{})
		So(err, ShouldNotBeNil)

		err = WriteFileContent(baseContext(), "main", "filename", bytes.NewBufferString("a"), 0, WriteOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_WriteFileContent2(t *testing.T) {
	Convey("FILES - WRITE: cannot write file content if context has no user", t, func() {
		setup()

		err := WriteFileContent(baseContext(), "main", "filename", bytes.NewBufferString("a"), 1, WriteOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_WriteFileContent3(t *testing.T) {
	Convey("FILES - WRITE: cannot write file content if context has user that has no write permission on target folder", t, func() {
		setup()

		userContext := getContextWithUser("user1")
		err := WriteFileContent(userContext, "main", "filename", bytes.NewBufferString("a"), 1, WriteOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_WriteFileContent4(t *testing.T) {
	Convey("FILES - WRITE: cannot write file content if context has no access manager", t, func() {
		setup()

		user1Access := getContextWithUserFromClientAndNoAccessManager("admin")
		err := WriteFileContent(user1Access, "user1-access", "file.txt", bytes.NewBufferString("a"), 1, WriteOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_WriteFileContent5(t *testing.T) {
	Convey("FILES - WRITE: can write file content if context has access manager and context has user with WRITE permission on target folder", t, func() {
		setup()

		err := WriteFileContent(adminContext(), "main", "file.txt", bytes.NewBufferString("a"), 1, WriteOptions{})
		So(err, ShouldBeNil)

		err = WriteFileContent(getContextWithUser("user1"), "user1-access", "file.txt", bytes.NewBufferString("a"), 1, WriteOptions{})
		So(err, ShouldBeNil)
	})
}

func TestHandler_ListAccess3(t *testing.T) {
	Convey("ACCESS LIST: can list accessDB which one of the READ rule is satisfied by the context user", t, func() {
		//todo: add ACL search on user, relation, objects for a given namespace
		/*setup()

		access, err := GetAccessList(clientAppContext(adminContext()), GetAccessListOptions{})
		So(err, ShouldBeNil)
		So(access, ShouldHaveLength, 1)

		access, err = GetAccessList(clientAppContext(getContextWithUser("user1")), GetAccessListOptions{})
		So(err, ShouldBeNil)
		So(access, ShouldHaveLength, 1) */
	})
}

func TestHandler_ListDir1(t *testing.T) {
	Convey("FILES - LS: cannot list directory if one of the following parameters is not set", t, func() {
		setup()

		_, err := ListDir(baseContext(), "", "/", ListDirOptions{})
		So(err, ShouldNotBeNil)

		_, err = ListDir(baseContext(), "user-access1", "", ListDirOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_ListDir2(t *testing.T) {
	Convey("FILES - LS: cannot list directory if context has no access manager", t, func() {
		setup()

		_, err := ListDir(getContextWithUserFromClientAndNoAccessManager("user1"), "user1-access", "/", ListDirOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_ListDir3(t *testing.T) {
	Convey("FILES - LS: cannot list directory if context user has no READ permission on it", t, func() {
		setup()

		_, err := ListDir(getContextWithUser("user1"), "main", "/", ListDirOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_ListDir4(t *testing.T) {
	Convey("FILES - LS: can list directory if context user has READ permission on it", t, func() {
		setup()

		_, err := ListDir(getContextWithUser("user1"), "user1-access", "/", ListDirOptions{})
		So(err, ShouldBeNil)
	})
}

func TestHandler_DeleteAccess1(t *testing.T) {
	Convey("ACCESS - DELETE: cannot delete access if one the following parameters is not provided: accessID", t, func() {
		setup()

		err := DeleteAccess(baseContext(), "", DeleteAccessOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_DeleteAccess2(t *testing.T) {
	Convey("ACCESS - DELETE: cannot delete access if the context has no access manager", t, func() {
		setup()

		adminContext := getContextWithUserFromClientAndNoAccessManager("admin")
		err := DeleteAccess(adminContext, "user1-access", DeleteAccessOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_DeleteAccess3(t *testing.T) {
	Convey("ACCESS - DELETE: cannot delete access if the context has no user", t, func() {
		setup()

		err := DeleteAccess(baseContext(), "main", DeleteAccessOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_DeleteAccess4(t *testing.T) {
	Convey("ACCESS - DELETE: cannot delete access if context user is not admin or the access is created by another user", t, func() {
		setup()

		err := DeleteAccess(getContextWithUser("user-1"), "main", DeleteAccessOptions{})
		So(err, ShouldNotBeNil)
	})
}

func TestHandler_DeleteAccess5(t *testing.T) {
	Convey("ACCESS - DELETE: can delete a access if it has been created by the context user", t, func() {
		setup()

		err := DeleteAccess(clientAppContext(adminContext()), "user1-access", DeleteAccessOptions{})
		So(err, ShouldBeNil)
	})
}

func TestHandler_DeleteAccess6(t *testing.T) {
	Convey("ACCESS - DELETE: cannot delete non existing access", t, func() {
		setup()

		err := DeleteAccess(adminContext(), "access", DeleteAccessOptions{})
		So(err, ShouldNotBeNil)
	})
}

/*func TestHandler_Clean(t *testing.T) {
	Convey("", t, func() {
		clearDir()
	})
} */
