package acl

import (
	pb "github.com/omecodes/store/gen/go/proto"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestDefaultManager_SaveNamespaceConfig(t *testing.T) {
	Convey("Save Namespace", t, func() {
		initDBs()
		man := &defaultManager{}
		err = man.SaveNamespaceConfig(fullManagerContext(), docNamespace)
		So(err, ShouldBeNil)
	})
}

func TestDefaultManager_SaveACL(t *testing.T) {
	Convey("Save ACL", t, func() {
		initDBs()

		man := &defaultManager{}
		for _, a := range dataACL {
			err = man.SaveACL(fullManagerContext(), &pb.ACL{
				Object:     a.Object,
				Relation:   a.Relation,
				Subject:    a.Subject,
			})
			So(err, ShouldBeNil)
		}
	})
}

func TestDefaultManager_CheckACL(t *testing.T) {
	Convey("Validate that /documents/description.pdf is readable by admin", t, func() {
		initDBs()

		man := &defaultManager{}
		checked, err := man.CheckACL(fullManagerContext(), "admin", &pb.SubjectSet{
			Object:   "doc:d12",
			Relation: "viewer",
		}, 0)
		So(err, ShouldBeNil)
		So(checked, ShouldBeTrue)
	})
}

func TestCloseDBs(t *testing.T) {
	Convey("Close DBs", t, func() {
		defer closeDBs()
		initDBs()
	})
}
