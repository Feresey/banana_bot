package db

import (
	"database/sql"
)

func QueryRow(q string, args ...interface{}) *sql.Row {
	log.Info("query:", q)
	return db.QueryRow(q, args...)
}

func Query(q string, args ...interface{}) (*sql.Rows, error) {
	log.Info("query:", q)
	return db.Query(q, args...)
}
