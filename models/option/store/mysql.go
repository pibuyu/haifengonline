package store

import "github.com/jmoiron/sqlx"

const (
	defaultTable = "dirties"
)

type MysqlConfig struct {
	Dsn       string
	Database  string
	TableName string
}

type Subject struct {
	Id   int64  `db:"id"`
	Word string `db:"word"`
}

type MysqlModel struct {
	store     *sqlx.DB
	TableName string
	addChan   chan string
	delChan   chan string
}
