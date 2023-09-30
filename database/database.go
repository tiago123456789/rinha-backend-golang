package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

func Start() *sql.DB {
	url := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		5432,
		os.Getenv("DB_NAME"),
	)

	db, err := sql.Open("postgres", url)

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
