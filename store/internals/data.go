package internals

import (
	"database/sql"
	"github.com/omecodes/common/dao"
	"strings"
)

type Store interface {
	Set(name, value string) error
	Update(name, value, part string) error
	Get(name, item string) (string, error)
	Delete(name, item string) error
}

func at(jp string) string {
	jp = strings.Replace(jp, "/", ".", -1)
	if strings.HasPrefix(jp, "$.") {
		return jp
	}
	if strings.HasPrefix(jp, ".") {
		return "$" + jp
	}
	return "$." + jp
}

func escaped(value string) string {
	replace := map[string]string{"\\": "\\\\", "'": `\'`, "\\0": "\\\\0", "\n": "\\n", "\r": "\\r", `"`: `\"`, "\x1a": "\\Z"}
	for b, a := range replace {
		value = strings.Replace(value, b, a, -1)
	}
	return value
}

type db struct {
	dao.SQL
}

func NewStore(d *sql.DB) (Store, error) {
	sdb := new(db)
	sdb.
		AddTableDefinition("name", `create table if not exists internals(
			name varchar(255) not null primary key,
			value json
		) ENGINE=InnoDB;`).
		AddStatement("insert", `insert into internals values (?, ?) on duplicate key update value=?;`).
		AddStatement("update_part", `update internals set value=json_insert(value, ?, ?) where name=?;`).
		AddStatement("select", `select value from internals where name=?;`).
		AddStatement("select_path", `select json_extract(value, ?) from internals where name=?;`).
		AddStatement("delete_part", `update internals set value=json_remove(value, ?) where name=?;`).
		AddStatement("delete_by_name", `delete from internals where name=?;`).
		RegisterScanner("string", dao.NewScannerFunc(func(row dao.Row) (interface{}, error) {
			var s string
			return s, row.Scan(&s)
		}))

	return sdb, sdb.InitWithMySQLDB(d)
}

func (sdb *db) Set(name string, value string) error {
	return sdb.Exec("insert", name, value, value).Error
}

func (sdb *db) Update(name string, value string, part string) error {
	return sdb.Exec("updated_part", escaped(part), escaped(value), name).Error
}

func (sdb *db) Get(name string, item string) (string, error) {
	var o interface{}
	var err error

	if item == "" {
		o, err = sdb.QueryOne("select", "string", name)
	} else {
		item = at(escaped(item))
		o, err = sdb.QueryOne("select_path", "string", item, name)
	}
	if err != nil {
		return "", err
	}
	return o.(string), err
}

func (sdb *db) Delete(name, item string) error {
	if item == "" {
		return sdb.Exec("delete_by_name", name).Error
	}
	return sdb.Exec("delete_part", escaped(item), name).Error
}
