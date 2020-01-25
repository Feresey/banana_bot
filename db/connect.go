package db

import (
	"database/sql"
	"os"

	_ "github.com/lib/pq" // постгря

	"github.com/Feresey/banana_bot/logging"
)

var (
	db      *sql.DB
	log     *logging.Logger
	maxWarn int

	warn   string = "warn"
	report string = "sub"
)

// Connect :
func Connect(logger *logging.Logger, warnMax int) error {
	var err error
	log = logger
	maxWarn = warnMax
	if maxWarn <= 0 {
		log.Panic("max warn in negative")
	}
	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	return err
}

// Shutdown :
func Shutdown() {
	db.Close()
}
