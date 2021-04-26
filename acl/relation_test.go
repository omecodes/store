package acl

import (
	"context"
	"github.com/omecodes/store/common/utime"

	pb "github.com/omecodes/store/gen/go/proto"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestRelationSQLStore_Save(t *testing.T) {
	Convey("Save acl relation tuple", t, func() {
		initRelationDB()

		ctx := context.Background()
		commitTime := utime.Now()

		for _, a := range dataACL {
			err = tupleStore.Save(ctx, &pb.DBEntry{
				Sid:        0,
				Object:     a.Object,
				Relation:   a.Relation,
				Subject:    a.Subject,
				CommitTime: commitTime,
			})
			So(err, ShouldBeNil)
		}
	})
}


func TestRelationSQLStore_Check(t *testing.T) {
	Convey("Check if relations exists", t, func() {
		initRelationDB()

		ctx := context.Background()
		var exists bool
		exists, err = tupleStore.Check(ctx, &pb.DBEntry{
			Object:     "doc:d1",
			Relation:   "editor",
			Subject:    "user1",
		})
		So(err, ShouldBeNil)
		So(exists, ShouldBeTrue)
	})
}

func TestRelationSQLStore_GetSubjectSet(t *testing.T) {
	Convey("Get subject set for a given relation and object", t, func() {
		initRelationDB()
		ctx := context.Background()

		var set []string
		set, err = tupleStore.GetSubjectSet(ctx, &pb.DBSubjectSetInfo{
			Object:   "doc:d11",
			Relation: "parent",
			MinAge:   0,
		})
		So(err, ShouldBeNil)
		So(set, ShouldHaveLength, 2)
	})
}

func TestRelationSQLStore_Delete(t *testing.T) {
	Convey("Check", t, func() {
		initRelationDB()

		ctx := context.Background()
		for _, a := range dataACL {
			err = tupleStore.Delete(ctx, &pb.DBEntry{
				Object:     a.Object,
				Relation:   a.Relation,
				Subject:    a.Subject,
				CommitTime: 0,
			})
			So(err, ShouldBeNil)
		}
	})
}


func TestCloseRelationDB(t *testing.T) {
	Convey("Closing database", t, func() {
		closeTupleDBConn()
		initNamespaceDB()
	})
}