package db

import "github.com/jackc/pgx/v4"

import "context"

func QueryRow(q string, args ...interface{}) pgx.Row {
	log.Info("query:", q)
	return db.QueryRow(context.Background(), q, args...)
}

func Query(q string, args ...interface{}) (pgx.Rows, error) {
	log.Info("query:", q)
	return db.Query(context.Background(), q, args...)
}