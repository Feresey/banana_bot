package db

import (
	"context"
	"os"

	"github.com/Feresey/bot-tg/logging"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	db  *pgxpool.Pool
	log *logging.Logger
)

func Connect(logger *logging.Logger) {
	var err error
	log = logger
	db, err = pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Warn("Unable connect to database:", err)
	}
}
