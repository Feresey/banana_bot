package migrations

import (
	"database/sql"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
)

//go:generate go run github.com/go-bindata/go-bindata/v3/go-bindata -pkg migrations -ignore '.*\.go' -prefix . .
//go:generate go fmt ./...

// Migrate schema fot n steps. If steps == 0 then migrate all up.
func Migrate(sql *sql.DB, steps int) error {
	s, err := bindata.WithInstance(bindata.Resource(AssetNames(), Asset))
	if err != nil {
		return err
	}

	p, err := postgres.WithInstance(sql, &postgres.Config{
		MigrationsTable:  "bot_migrate",
		StatementTimeout: time.Minute,
	})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance(
		"go-bindata", s,
		"pgx", p)
	if err != nil {
		return err
	}

	if steps != 0 {
		return m.Steps(steps)
	}
	return m.Up()
}
