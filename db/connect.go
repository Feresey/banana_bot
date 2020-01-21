package db

import (
	"database/sql"
	"os"

	_ "github.com/lib/pq"

	"github.com/Feresey/banana_bot/logging"
)

var (
	DB  *sql.DB
	log *logging.Logger
)

func Connect(logger *logging.Logger) {
	var err error
	log = logger
	DB, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Warn("Unable connect to database:", err)
	}
}
