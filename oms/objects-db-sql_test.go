package oms

import (
	"bytes"
	"context"
	"database/sql"
	"github.com/omecodes/bome"
	"io/ioutil"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	db      *sql.DB
	objects Objects
)

func initDB() {
	if objects == nil {
		var err error
		db, err = sql.Open(testDialect, testDBUri)
		So(err, ShouldBeNil)
		So(db, ShouldNotBeNil)

		objects, err = NewSQLObjects(db, "unknown-dialect")
		So(objects, ShouldBeNil)
		So(err, ShouldNotBeNil)

		objects, err = NewSQLObjects(db, testDialect)
		So(err, ShouldBeNil)
		So(objects, ShouldNotBeNil)

		err = objects.Clear()
		So(err, ShouldBeNil)
	}
}

func TestNewStore(t *testing.T) {
	Convey("Init objects store", t, func() {
		initDB()
	})
}

func TestMysqlStore_Save(t *testing.T) {
	Convey("Save entries", t, func() {
		initDB()

		var content = `{
	"project": "ome",
	"private": true,
	"git": "https://github.com/omecodes/ome.git",
	"description": "Service Authority. Generates and signs certificates for services."
}`
		o := new(Object)
		o.SetHeader(&Header{
			Id:        "ome-ca",
			CreatedBy: "ome",
			Size:      int64(len(content)),
		})
		o.SetContent(bytes.NewBufferString(content), int64(len(content)))
		err := objects.Save(context.Background(), o)
		So(err, ShouldBeNil)

		content = `{
	"project": "accounts",
	"private": true,
	"git": "https://github.com/omecodes/accounts.git",
	"description": "Account manager application. Supports OAUTH2"
}`
		o = new(Object)
		o.SetHeader(&Header{
			Id:        "ome-accounts",
			CreatedBy: "ome",
			Size:      int64(len(content)),
		})
		o.SetContent(bytes.NewBufferString(content), int64(len(content)))
		err = objects.Save(context.Background(), o)
		So(err, ShouldBeNil)

		content = `{
	"project": "tdb",
	"private": true,
	"git": "https://github.com/omecodes/tdb.git",
	"description": "Token store app"
}`
		o = new(Object)
		o.SetHeader(&Header{
			Id:        "ome-tdb",
			CreatedBy: "ome",
			Size:      int64(len(content)),
		})
		o.SetContent(bytes.NewBufferString(content), int64(len(content)))
		err = objects.Save(context.Background(), o)
		So(err, ShouldBeNil)

		content = `{
	"project": "libome",
	"private": true,
	"git": "https://github.com/omecodes/libome.git",
	"description": "Base library for all service definition"
}`
		o = new(Object)
		o.SetHeader(&Header{
			Id:        "ome-libome",
			CreatedBy: "ome",
			Size:      int64(len(content)),
		})
		o.SetContent(bytes.NewBufferString(content), int64(len(content)))
		err = objects.Save(context.Background(), o)
		So(err, ShouldBeNil)
	})
}

func TestMysqlStore_Patch(t *testing.T) {
	if jsonTestEnabled {
		Convey("Patch object", t, func() {
			initDB()

			ctx := context.Background()
			o, err := objects.GetAt(ctx, "ome-libome", "$.private")
			So(err, ShouldBeNil)

			value, err := ioutil.ReadAll(o.Content())
			So(err, ShouldBeNil)
			So(string(value), ShouldEqual, "true")

			err = objects.Patch(ctx, "ome-libome", "$.private", "false")
			So(err, ShouldBeNil)

			o, err = objects.GetAt(ctx, "ome-libome", "$.private")
			So(err, ShouldBeNil)

			value, err = ioutil.ReadAll(o.Content())
			So(err, ShouldBeNil)
			So(string(value), ShouldEqual, "false")
		})
	}
}

func TestMysqlStore_Get(t *testing.T) {
	Convey("Get item", t, func() {
		initDB()
		o, err := objects.Get(context.Background(), "ome-ca")
		So(err, ShouldBeNil)
		So(o.header.Id, ShouldEqual, "ome-ca")

		o, err = objects.Get(context.Background(), "non-existing-object-id")
		So(err, ShouldNotBeNil)
		So(o, ShouldBeNil)
	})
}

func TestMysqlStore_GetAt(t *testing.T) {
	Convey("Get content at", t, func() {
		initDB()

		o, err := objects.GetAt(context.Background(), "ome-ca", "$.fire")
		So(err, ShouldNotBeNil)
		So(o, ShouldBeNil)
	})
}

func TestMysqlStore_Info(t *testing.T) {
	Convey("Get object header", t, func() {
		initDB()

		header, err := objects.Info(context.Background(), "ome-accounts")
		So(err, ShouldBeNil)
		So(header.Id, ShouldEqual, "ome-accounts")
	})
}

func TestMysqlStore_List(t *testing.T) {
	Convey("List objects", t, func() {
		initDB()

		now := time.Now().Unix()

		list, err := objects.List(context.Background(), now, 3, FilterObjectFunc(func(o *Object) (bool, error) {
			return true, nil
		}))
		So(err, ShouldBeNil)
		So(list.Objects, ShouldHaveLength, 3)

		list, err = objects.List(context.Background(), now, 3, FilterObjectFunc(func(o *Object) (bool, error) {
			return false, nil
		}))
		So(err, ShouldBeNil)
		So(list.Objects, ShouldHaveLength, 0)

	})
}

func TestMysqlStore_ListAt(t *testing.T) {
	Convey("List objects item at", t, func() {
		now := time.Now().Unix()
		list, err := objects.ListAt(context.Background(), "$.private", now, 3, FilterObjectFunc(func(o *Object) (bool, error) {
			return true, nil
		}))
		So(err, ShouldBeNil)
		So(list.Objects, ShouldHaveLength, 3)

		for _, o := range list.Objects {
			content, err := ioutil.ReadAll(o.Content())
			So(err, ShouldBeNil)
			So(string(content), ShouldBeIn, "true", "false")
		}
	})
}

func TestMysqlStore_Delete(t *testing.T) {
	Convey("DeleteObject objects", t, func() {
		initDB()

		ctx := context.Background()

		err := objects.Delete(ctx, "ome-accounts")
		So(err, ShouldBeNil)

		o, err := objects.Get(ctx, "ome-accounts")
		So(bome.IsNotFound(err), ShouldBeTrue)
		So(o, ShouldBeNil)
	})
}
