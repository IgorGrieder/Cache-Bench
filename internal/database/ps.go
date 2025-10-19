package database

import (
	"fmt"
	"os"

	"database/sql"

	"github.com/IgorGrieder/Cache-Bench/internal/config"
	_ "github.com/lib/pq"
)

func SetupPG(cfg *config.Config) *sql.DB {

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		cfg.HOST, cfg.PORT_PG, cfg.USER, cfg.PG_PASS, cfg.DB_NAME)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		fmt.Printf("Ending the execution %v", err)
		os.Exit(1)
	}

	return db
}
