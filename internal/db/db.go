package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func Connect(url string) *sql.DB {
	db, err := sql.Open("postgres", url)
	if err != nil {
		log.Fatal("failed to connect to DB:", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("failed to ping DB:", err)
	}
	return db
}
