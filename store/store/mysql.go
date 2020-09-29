package store

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/google/uuid"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/omestore/pb"
	"strings"
	"time"
)

const createTableQuery = `
create table if not exists $table$ (
	id varchar(255) not null primary key,
	created_by varchar(255) not null,
	created_at integer not null,
	size integer not null,
	value json not null
) ENGINE=InnoDB;
`

const createTableGraftQuery = `
create table if not exists $graft_table$ (
	id varchar(255) not null,
  	data_id varchar(255) not null,
	created_by varchar(255) not null,
	created_at integer not null,
	size integer not null,
	value json not null,
	unique (id, data_id),
	foreign key (data_id) references $table$(id) on delete cascade
) ENGINE=InnoDB;
`

func MySQL(db *sql.DB) (*mysqlStore, error) {
	od := &mysqlStore{
		DB: db,
	}
	_, err := db.Exec(`create table if not exists collections (
			name varchar(255) primary key not null,
			metadata json not null,
			created_at int
		) ENGINE=InnoDB;`)
	return od, err
}

type mysqlStore struct {
	*sql.DB
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

func (ms *mysqlStore) collectionExists(collection string) (bool, error) {
	rows, err := ms.DB.Query("select 1 from collections where name=?;", collection)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), err
}

func (ms *mysqlStore) tableName(collection string) string {
	return "col_" + hex.EncodeToString([]byte(collection))
}

func (ms *mysqlStore) graftTableName(collection string) string {
	return "col_grafts_" + hex.EncodeToString([]byte(collection))
}

func (ms *mysqlStore) createCollectionAndSaveData(ctx context.Context, data *pb.Data) error {
	tx, err := ms.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	_, err = tx.Exec("insert into collections values(?, ?, ?);", data.Collection, "{}", time.Now().Unix())
	if err != nil {
		return err
	}

	table := ms.tableName(data.Collection)
	query := strings.Replace(createTableQuery, "$table$", table, 1)
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	graftTable := ms.graftTableName(data.Collection)
	query = strings.Replace(createTableGraftQuery, "$graft_table$", graftTable, 1)
	query = strings.Replace(query, "$table$", table, 1)
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	_, err = tx.Exec("insert into "+table+" values (?, ?, ?, ?, ?);", data.ID, data.CreatedBy, data.CreatedAt, data.Size, data.Content)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (ms *mysqlStore) Save(ctx context.Context, data *pb.Data) error {
	exists, err := ms.collectionExists(data.Collection)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if !exists {
		return ms.createCollectionAndSaveData(ctx, data)
	}

	table := ms.tableName(data.Collection)
	_, err = ms.DB.Exec(
		fmt.Sprintf("insert into `%s` values (?, ?, ?, ?, ?) on duplicate key update value=?;", table),
		data.ID,
		data.CreatedBy,
		data.CreatedAt,
		data.Size,
		data.Content,
		data.Content,
	)
	return err
}

func (ms *mysqlStore) Update(ctx context.Context, data *pb.Data) error {
	table := ms.tableName(data.Collection)
	var err error
	_, err = ms.DB.Exec(fmt.Sprintf("update `%s` set value=? where id=?;", table), data.Content, data.ID)
	return err
}

func (ms *mysqlStore) Delete(ctx context.Context, data *pb.Data) error {
	table := ms.tableName(data.Collection)
	var err error
	_, err = ms.DB.Exec(fmt.Sprintf("delete from `%s` where id=?;", table), data.ID)
	return err
}

func (ms *mysqlStore) List(ctx context.Context, collection string, opts pb.ListOptions) (*pb.DataList, error) {
	exists, err := ms.collectionExists(collection)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	if !exists {
		return nil, errors.NotFound
	}

	table := ms.tableName(collection)
	if opts.Path != "" {
		s := "select id, created_by, created_at, size, json_unquote(json_extract(value, \"%s\")) from `%s` where created_at < ?;"
		s = fmt.Sprintf(fmt.Sprintf(s, at(escaped(opts.Path)), table))
		rows, err := ms.DB.Query(s)
		if err != nil {
			if sql.ErrNoRows == err {
				return nil, errors.NotFound
			}
			return nil, err
		}
		return &pb.DataList{
			Collection: collection,
			Cursor:     newDataCursor(rows, opts.IDFilter, opts.Count),
		}, nil

	} else {
		rows, err := ms.DB.Query(fmt.Sprintf("select * from `%s`;", table))
		if err != nil {
			if sql.ErrNoRows == err {
				return nil, errors.NotFound
			}
			return nil, err
		}

		return &pb.DataList{
			Collection: collection,
			Cursor:     newDataCursor(rows, opts.IDFilter, opts.Count),
		}, nil
	}
}

func (ms *mysqlStore) Collections(ctx context.Context) ([]string, error) {
	c, err := ms.DB.Query("select name from collections;")
	if err != nil {
		return nil, err
	}

	defer c.Close()
	var names []string
	for c.Next() {
		var name string
		err = c.Scan(&name)
		if err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, nil
}

func (ms *mysqlStore) Get(ctx context.Context, collection string, id string, opts pb.DataOptions) (*pb.Data, error) {
	exists, err := ms.collectionExists(collection)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	if !exists {
		return nil, errors.NotFound
	}

	table := ms.tableName(collection)
	if opts.Path != "" {
		table := ms.tableName(collection)
		filter := at(escaped(strings.Replace(opts.Path, "/", ".", -1)))
		row := ms.DB.QueryRow(fmt.Sprintf("select id, created_by, created_at, size, json_unquote(json_extract(value, '%s')) from `%s` where id=?;", filter, table), id)

		data := new(pb.Data)
		err := row.Scan(&data.ID, &data.CreatedBy, &data.CreatedAt, &data.Size, &data.Content)
		if err != nil {
			if sql.ErrNoRows == err {
				return nil, errors.NotFound
			}
			return nil, err
		}
		return data, nil
	}

	query := fmt.Sprintf("select * from `%s` where id=?;", table)
	row := ms.DB.QueryRow(query, id)

	data := new(pb.Data)
	err = row.Scan(&data.ID, &data.CreatedBy, &data.CreatedAt, &data.Size, &data.Content)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (ms *mysqlStore) Info(ctx context.Context, collection string, id string) (*pb.Info, error) {
	exists, err := ms.collectionExists(collection)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	if !exists {
		return nil, errors.NotFound
	}

	table := ms.tableName(collection)
	query := fmt.Sprintf("select id, created_by, created_at, size from `%s` where id=?;", table)
	row := ms.DB.QueryRow(query, id)

	info := new(pb.Info)
	err = row.Scan(&info.ID, &info.CreatedBy, &info.CreatedAt, &info.Size)
	if err != nil {
		if sql.ErrNoRows == err {
			return nil, errors.NotFound
		}
		return nil, err
	}
	info.Collection = collection
	return info, nil
}

func (ms *mysqlStore) Search(ctx context.Context, collection string, condition *any.Any, opts pb.ListOptions) (*pb.DataList, error) {
	exists, err := ms.collectionExists(collection)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	if !exists {
		return nil, errors.NotFound
	}

	/*table := ms.tableName(collection)
	table = escaped(table)
	whereClause, err := clauseFromCondition(condition)
	if err != nil {
		return nil, err
	}

	builder := strings.Builder{}
	if opts.Path != "" {
		builder.WriteString("select id, created_by, created_at, json_extract(value, \"")
		builder.WriteString(at(escaped(opts.Path)))
	} else {
		builder.WriteString("select * from")
	}
	builder.WriteString(fmt.Sprintf(" %s", table))
	builder.WriteString(" where")
	builder.WriteString(" ")
	builder.WriteString(whereClause)

	query := builder.String()

	rows, err := ms.DB.Query(query, "string")

	dl := &pb.DataList{
		Collection: collection,
		Cursor: pb.NewDataCursor(rows, opts.IDFilter, opts.Count),
	}

	return dl, err */
	return nil, nil
}

func (ms *mysqlStore) SaveGraft(ctx context.Context, graft *pb.Graft) (string, error) {
	table := ms.graftTableName(graft.Collection)
	query := fmt.Sprintf("insert into %s values(?, ?, ?, ?, ?, ?);", table)

	id := uuid.New().String()
	_, err := ms.DB.Exec(query, id, graft.DataID, graft.CreatedBy, graft.CreatedAt, graft.Size, &graft.Content)
	return id, err
}

func (ms *mysqlStore) GetGraft(ctx context.Context, collection string, dataID string, id string) (*pb.Graft, error) {
	table := ms.graftTableName(collection)
	query := fmt.Sprintf("select * from `%s` where data_id=? and id=?;", table)
	row := ms.DB.QueryRow(query, dataID, id)

	graft := new(pb.Graft)
	err := row.Scan(&graft.ID, &graft.DataID, &graft.CreatedBy, &graft.CreatedAt, &graft.Size, &graft.Content)
	if err != nil {
		if sql.ErrNoRows == err {
			return nil, errors.NotFound
		}
		return nil, err
	}
	return graft, nil
}

func (ms *mysqlStore) GraftInfo(ctx context.Context, collection string, dataID string, id string) (*pb.GraftInfo, error) {
	table := ms.graftTableName(collection)
	query := fmt.Sprintf("select id, data_id, created_by, created_at, size from `%s` where data_id=? and id=?;", table)
	row := ms.DB.QueryRow(query, dataID, id)

	info := new(pb.GraftInfo)
	err := row.Scan(&info.ID, &info.DataID, &info.CreatedBy, &info.CreatedAt, &info.Size)
	if err != nil {
		if sql.ErrNoRows == err {
			return nil, errors.NotFound
		}
		return nil, err
	}
	return info, nil
}

func (ms *mysqlStore) DeleteGraft(ctx context.Context, collection string, dataID string, id string) error {
	table := ms.graftTableName(collection)
	query := fmt.Sprintf("delete from `%s` where data_id=? and id=?;", table)
	_, err := ms.DB.Exec(query, dataID, id)
	return err
}

func (ms *mysqlStore) GetAllGraft(ctx context.Context, collection string, dataID string, opts pb.ListOptions) (*pb.GraftList, error) {
	table := ms.graftTableName(collection)
	if opts.Before == 0 {
		opts.Before = time.Now().Unix()
	}
	query := fmt.Sprintf("select * from `%s` where data_id=? and created_at < ?;", table)
	rows, err := ms.DB.Query(query, dataID, opts.Before)
	if err != nil {
		return nil, err
	}

	gl := &pb.GraftList{
		Collection: collection,
		DataID:     dataID,
		Cursor:     newGraftCursor(rows, opts.IDFilter, opts.Count),
	}
	return gl, nil
}

func clauseFromCondition(condition *any.Any) (string, error) {
	switch strings.ToLower(condition.TypeUrl) {
	case "eval":
		return evalWhereClause(condition)
	case "or", "and":
		return operatorCondition(condition)
	case "not":
		return operatorCondition(condition)
	default:
		log.Info("operator not supported", log.Field("name", condition.TypeUrl))
		return "", errors.NotSupported
	}
}

func operatorCondition(any *any.Any) (string, error) {
	b := strings.Builder{}
	var conditions pb.Conditions
	err := ptypes.UnmarshalAny(any, &conditions)
	if err != nil {
		return "", err
	}

	condition := conditions.Items[0]
	strCond, err := clauseFromCondition(condition)
	if err != nil {
		return "", err
	}

	b.WriteString("(")
	b.WriteString(strCond)

	for _, condition := range conditions.Items[1:] {
		strCond, err := clauseFromCondition(condition)
		if err != nil {
			return "", err
		}
		b.WriteString(" ")
		b.WriteString(any.TypeUrl)
		b.WriteString(" ")
		b.WriteString(strCond)
	}
	return b.String(), nil
}

func evalWhereClause(any *any.Any) (string, error) {
	var p pb.OperationParams
	err := ptypes.UnmarshalAny(any, &p)
	if err != nil {
		return "", err
	}

	switch p.Func {
	case "has":
		return fmt.Sprintf("json_contains(value, '\"%s\"', '\"%s\"')", escaped(p.Path), escaped(p.Value)), nil
	case "ex":
		return fmt.Sprintf("json_contains_path(value, '\"%s\"')", escaped(p.Path)), nil
	case "eq":
		return fmt.Sprintf("json_extracts(value, '\"%s\"')=='\"%s\"'", escaped(p.Path), escaped(p.Value)), nil
	case "eqn":
		return fmt.Sprintf("json_extracts(value, '\"%s\"')==%s", escaped(p.Path), escaped(p.Value)), nil
	default:
		log.Info("function not supported", log.Field("name", p.Func))
		return "", errors.NotSupported
	}
}
