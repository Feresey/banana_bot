package migrations

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
)

//go:embed *.sql
var migrations embed.FS

func Migrate(db *sql.DB) error {
	sourceInstance, err := httpfs.New(http.FS(migrations), ".")
	if err != nil {
		return err
	}
	defer sourceInstance.Close()

	targetInstance, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable:  "bot_migrate",
		StatementTimeout: time.Minute,
	})
	if err != nil {
		return err
	}
	defer targetInstance.Close()

	m, err := migrate.NewWithInstance("go-embed", sourceInstance, "postgres", targetInstance)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("up: %w", err)
		}
	}
	return nil
}
