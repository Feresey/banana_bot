package db

import (
	"context"
	"os"

	"github.com/Feresey/bot-tg/logging"
	pgx "github.com/jackc/pgx/v4"
)

var (
	DB  *pgx.Conn
	log *logging.Logger
)

func Connect(logger *logging.Logger) {
	var err error
	log = logger
	DB, err = pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Warn("Unable connect to database:", err)
	}
}
