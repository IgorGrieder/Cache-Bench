package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/IgorGrieder/Cache-Bench/cmd/handlers"
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

	// Handler
	h := handlers.NewHandler(redis, pg)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /test/cache-aside", h.GetProductCacheAside)
	mux.HandleFunc("GET /test/write-behind", h.UpdateProductWriteBehind)
	mux.HandleFunc("GET /test/write-through", h.UpdateProductWriteThrough)

	svr := &http.Server{Addr: fmt.Sprintf(":%d", cfg.PORT), Handler: mux}

	if err := svr.ListenAndServe(); err != nil {
		fmt.Println("Server crashed for some reason")
		os.Exit(1)
	}
}
