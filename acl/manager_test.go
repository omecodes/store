package acl

import (
	"fmt"
	pb "github.com/omecodes/store/gen/go/proto"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestDefaultManager_SaveNamespaceConfig(t *testing.T) {
	Convey("Save Namespace", t, func() {
		setup()

		man := &DefaultManager{}
		err := man.SaveNamespaceConfig(fullManagerContext(), docNamespace)
		So(err, ShouldBeNil)

		err = man.SaveNamespaceConfig(fullManagerContext(), groupNamespace)
		So(err, ShouldBeNil)
	})
}

func TestDefaultManager_SaveACL(t *testing.T) {
	Convey("Save ACL", t, func() {
		setup()

		man := &DefaultManager{}
		for _, a := range dataACL {
			err := man.SaveACL(fullManagerContext(), &pb.ACL{
				Object:   a.Object,
				Relation: a.Relation,
				Subject:  a.Subject,
			})
			So(err, ShouldBeNil)
		}
	})
}

func TestDefaultManager_CheckACL(t *testing.T) {
	Convey("Validate that /documents/description.pdf is readable by admin", t, func() {
		setup()

		man := &DefaultManager{}
		checked, err := man.CheckACL(fullManagerContext(), "admin", &pb.SubjectSet{
			Object:   "doc:d12",
			Relation: "viewer",
		})
		So(err, ShouldBeNil)
		So(checked, ShouldBeTrue)
	})
}

func TestDefaultManager_CheckACL1(t *testing.T) {
	Convey("Validate that /documents/description.pdf is readable by admin", t, func() {
		setup()

		man := &DefaultManager{}
		checked, err := man.CheckACL(fullManagerContext(), "yaba", &pb.SubjectSet{
			Object:   "group:external",
			Relation: "member",
		})
		So(err, ShouldBeNil)
		So(checked, ShouldBeTrue)
	})
}

func TestDefaultManager_CheckACL2(t *testing.T) {
	Convey("Validate that /documents/description.pdf is readable by admin", t, func() {
		setup()

		fmt.Println()

		man := &DefaultManager{}
		checked, err := man.CheckACL(fullManagerContext(), "yaba", &pb.SubjectSet{
			Object:   "doc:d12",
			Relation: "viewer",
		})
		So(err, ShouldBeNil)
		So(checked, ShouldBeTrue)
	})
}

func TestDefaultManager_ResolveUserSet(t *testing.T) {
	Convey("", t, func() {
		setup()
	})
}

func TestCloseDBs(t *testing.T) {
	Convey("Close DBs", t, func() {
		defer closeDBs()
		setup()
	})
}
