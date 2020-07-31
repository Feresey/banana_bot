package db

import (
	"context"
	"fmt"
	"time"

	"github.com/Feresey/banana_bot/internal/db/migrations"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/zapadapter"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jackc/pgx/v4/stdlib"
	"go.uber.org/zap"
)

const schemaName = "bot."

var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

type Config struct {
	SQL      string
	LogLevel string
	Migrate  int

	ConnectTimeout time.Duration
}

// Database
// Главный (и единственный) инстанс датабейза.
type Database struct {
	c    *Config
	log  *zap.Logger
	pool *pgxpool.Pool
}

func New(log *zap.Logger, c Config) *Database {
	return &Database{
		c:   &c,
		log: log.Named("db"),
	}
}

func (db *Database) Init(ctx context.Context) error {
	poolConf, err := pgxpool.ParseConfig(db.c.SQL)
	if err != nil {
		return fmt.Errorf("parse sql string: %w", err)
	}

	poolConf.ConnConfig.LogLevel, err = pgx.LogLevelFromString(db.c.LogLevel)
	if err != nil {
		return fmt.Errorf("parse log level: %w", err)
	}
	poolConf.ConnConfig.Logger = zapadapter.NewLogger(db.log)

	pool, err := pgxpool.ConnectConfig(ctx, poolConf)
	if err != nil {
		return fmt.Errorf("connect to db: %w", err)
	}
	db.pool = pool

	sql := stdlib.OpenDB(*poolConf.Copy().ConnConfig)
	defer sql.Close()

	return migrations.Migrate(sql, db.c.Migrate)
}

func (db *Database) Shutdown(_ context.Context) error {
	db.pool.Close()
	return nil
}

func (db *Database) tx(ctx context.Context, f func(pgx.Tx) error) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction %w", err)
	}
	if err := f(tx); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("execute transaction %w, rejected", err)
	}
	return tx.Commit(ctx)
}

type executor interface {
	Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

type sqlizer interface {
	ToSql() (string, []interface{}, error)
}

func zero(ctx context.Context, tx executor, qb sqlizer) error {
	query, params, err := qb.ToSql()
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, query, params...)
	return err
}

func one(ctx context.Context, tx executor, qb sqlizer, vptr ...interface{}) error {
	query, params, err := qb.ToSql()
	if err != nil {
		return err
	}
	row := tx.QueryRow(ctx, query, params...)
	return row.Scan(vptr...)
}
