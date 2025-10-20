package handlers

import (
	"database/sql"
	"net/http"

	"github.com/redis/go-redis/v9"
)

type handler struct {
	redis *redis.Client
	pg    *sql.DB
}

func NewHandler(redis *redis.Client, pg *sql.DB) *handler {
	return &handler{redis: redis, pg: pg}
}

func (h *handler) HandlerTest(w http.ResponseWriter, r *http.Request) {

}
