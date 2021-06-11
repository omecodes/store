package acl

import (
	"github.com/omecodes/errors"
	pb "github.com/omecodes/store/gen/go/proto"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestNewNamespaceSQLStore(t *testing.T) {
	Convey("SQLite based Namespace tupleStore should be initialized with no errors", t, func() {
		setupNamespaceDB()
	})
}

func TestNamespaceSQLStore_SaveNamespace(t *testing.T) {
	Convey("Save doc namespace", t, func() {
		setupNamespaceDB()
		err := namespaceStore.SaveNamespace(docNamespace)
		So(err, ShouldBeNil)
	})
}

func TestNamespaceSQLStore_GetNamespace(t *testing.T) {
	Convey("Get namespace", t, func() {
		setupNamespaceDB()

		ns, err := namespaceStore.GetNamespace(docNamespace.Namespace)
		So(err, ShouldBeNil)
		So(ns.Namespace, ShouldEqual, docNamespace.Namespace)
	})
}

func TestNamespaceSQLStore_GetRelationDefinition(t *testing.T) {
	Convey("Get namespace relation definition", t, func() {
		setupNamespaceDB()

		var def *pb.RelationDefinition
		def, err := namespaceStore.GetRelationDefinition(docNamespace.Namespace, "viewer")
		So(err, ShouldBeNil)
		So(def, ShouldNotBeNil)
		So(def.Name, ShouldEqual, "viewer")
	})
}

func TestNamespaceSQLStore_DeleteNamespace(t *testing.T) {
	Convey("Delete namespace", t, func() {
		setupNamespaceDB()

		err := namespaceStore.DeleteNamespace(docNamespace.Namespace)
		So(err, ShouldBeNil)

		ns, err := namespaceStore.GetNamespace(docNamespace.Namespace)
		So(errors.IsNotFound(err), ShouldBeTrue)
		So(ns, ShouldBeNil)
	})
}

func TestCloseNamespaceDB(t *testing.T) {
	Convey("Closing database", t, func() {
		defer closeNamespaceDBConn()
		setupNamespaceDB()
	})
}
