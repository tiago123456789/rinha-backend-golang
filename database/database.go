package database

import (
	"database/sql"
	"log"
	"os"
)

func Start() *sql.DB {
	db, err := sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal(err)
	}

	db.Ping()
	_, err = db.Exec(
		"CREATE TABLE IF NOT EXISTS peoples (id uuid NOT NULL," +
			" apelido varchar(32) NOT NULL, nome varchar(100) NOT NULL, " +
			" nascimento DATE NOT NULL, stack JSONB )",
	)
	if err != nil {
		log.Fatal(err)
	}

	return db
}
