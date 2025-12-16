package middleware

import (
	"fmt"
	"md/model/common"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

// 数据库读写连接
var Db *sqlx.DB

// 数据库写连接
var DbW *sqlx.DB

// 建表语句
var createTableSql = `
CREATE TABLE IF NOT EXISTS t_user
(
	id varchar(50) PRIMARY KEY NOT NULL,
	name text NOT NULL,
	password text NOT NULL,
	create_time bigint NOT NULL
);

CREATE TABLE IF NOT EXISTS t_document
(
	id varchar(50) PRIMARY KEY NOT NULL,
	name text NOT NULL,
	content text NOT NULL,
	type text NOT NULL,
	published boolean NOT NULL,
	create_time bigint NOT NULL,
	update_time bigint NOT NULL,
	book_id varchar(50) NOT NULL,
	user_id varchar(50) NOT NULL
);

CREATE TABLE IF NOT EXISTS t_book
(
	id varchar(50) PRIMARY KEY NOT NULL,
	name text NOT NULL,
	create_time bigint NOT NULL,
	user_id varchar(50) NOT NULL
);

CREATE TABLE IF NOT EXISTS t_picture
(
	id varchar(50) PRIMARY KEY NOT NULL,
	name text NOT NULL,
	path text NOT NULL,
	hash text NOT NULL,
	size bigint NOT NULL,
	create_time bigint NOT NULL,
	user_id varchar(50) NOT NULL
);

CREATE INDEX IF NOT EXISTS "book_user_id"
ON "t_book" (
  "user_id" ASC
);

CREATE INDEX IF NOT EXISTS "document_user_id_book_id"
ON "t_document" (
  "user_id" ASC,
  "book_id" ASC
);

CREATE INDEX IF NOT EXISTS "picture_size_hash"
ON "t_picture" (
  "size" ASC,
  "hash" ASC
);

CREATE INDEX IF NOT EXISTS "picture_user_id"
ON "t_picture" (
  "user_id" ASC
);

CREATE UNIQUE INDEX IF NOT EXISTS "user_name"
ON "t_user" (
  "name" ASC
);
`

// 初始化数据库连接
func InitDB() error {
	var err error
	if common.PostgresHost != "" && common.PostgresPort != "" && common.PostgresUser != "" && common.PostgresPassword != "" && common.PostgresDB != "" {
		err = initPostgres()
	} else {
		err = initSqlite()
	}
	if err != nil {
		return err
	}

	// Create base tables
	Db.MustExec(createTableSql)

	// Run migrations for schema updates
	if err = RunMigrations(); err != nil {
		Log.Error("Database migration failed: ", err)
		return err
	}

	return nil
}

// 初始化sqlite
func initSqlite() error {
	// 开启数据库文件
	var err error
	Db, err = sqlx.Connect("sqlite", common.DataPath+"md.db")
	if err != nil {
		Log.Error("开启sqlite数据库文件失败：", err)
		return err
	}

	DbW, err = sqlx.Connect("sqlite", common.DataPath+"md.db")
	if err != nil {
		Log.Error("开启sqlite数据库文件失败：", err)
		return err
	}
	DbW.SetMaxOpenConns(1)

	Log.Info("已连接sqlite")
	return nil
}

// 初始化postgres
func initPostgres() error {
	var err error
	Db, err = sqlx.Connect("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", common.PostgresHost, common.PostgresPort, common.PostgresUser, common.PostgresPassword, common.PostgresDB))
	if err != nil {
		Log.Error("postgres连接失败：", err)
		return err
	}

	DbW = Db

	Log.Info("已连接postgres")
	return nil
}
