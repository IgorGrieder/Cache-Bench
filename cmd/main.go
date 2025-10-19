package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/IgorGrieder/Cache-Bench/internal/config"
	"github.com/IgorGrieder/Cache-Bench/internal/database"
)

func main() {
	fmt.Println("Starting the program")

	// ENVs
	cfg := config.NewConfig()

	// Database connections
	redis := database.SetupRedis()
	pg := database.SetupPG(cfg)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /test", func(http.ResponseWriter, *http.Request) {
		redis.Set()
		pg.BeginTx()
	})

	svr := &http.Server{Addr: fmt.Sprintf(":%d", cfg.PORT), Handler: mux}

	if err := svr.ListenAndServe(); err != nil {
		fmt.Println("Server crashed for some reason")
		os.Exit(1)
	}
}
