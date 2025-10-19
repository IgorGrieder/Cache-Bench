package database

import (
	"fmt"
	"os"

	"database/sql"

	_ "github.com/lib/pq"
)

func SetupPG() *sql.DB {

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		cfg.HOST, 5432, cfg.USER, cfg.PG_PASS, cfg.DB_NAME)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		fmt.Printf("Ending the execution %v", err)
		os.Exit(1)
	}

	return db
}
