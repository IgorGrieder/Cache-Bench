package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type Product struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type handler struct {
	redis *redis.Client
	pg    *sql.DB
}

func NewHandler(redis *redis.Client, pg *sql.DB) *handler {
	return &handler{redis: redis, pg: pg}
}

func (h *handler) GetProductCacheAside(w http.ResponseWriter, r *http.Request) {
	// Let's assume we get the product ID from the query params
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Product ID is required", http.StatusBadRequest)
		return
	}

	// Look for data in the cache
	cacheKey := "product:" + id
	cachedProduct, err := h.redis.Get(context.Background(), cacheKey).Bytes()

	// Cache hit
	if err == nil {
		log.Println("Cache HIT")
		w.Header().Set("Content-Type", "application/json")
		w.Write(cachedProduct)
		return
	}

	// Cache Miss
	log.Println("Cache MISS")
	var product Product

	// Query the database
	err = h.pg.QueryRow("SELECT id, name, price FROM products WHERE id = $1", id).Scan(&product.ID, &product.Name, &product.Price)
	if err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	// Populate the cache for next time
	productJSON, _ := json.Marshal(product)

	// Set with a TTL of 5 minutes
	h.redis.Set(context.Background(), cacheKey, productJSON, 5*time.Minute)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

// UpdateProductWriteThrough demonstrates the Write-Through pattern for writes.
func (h *handler) UpdateProductWriteThrough(w http.ResponseWriter, r *http.Request) {
	var product Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	cacheKey := "product:" + product.ID

	// Write to the database first
	_, err := h.pg.Exec("UPDATE products SET name = $1, price = $2 WHERE id = $3", product.Name, product.Price, product.ID)
	if err != nil {
		http.Error(w, "Failed to update database", http.StatusInternalServerError)
		return
	}

	// Write the same data to the cache
	productJSON, _ := json.Marshal(product)
	err = h.redis.Set(context.Background(), cacheKey, productJSON, 5*time.Minute).Err()
	if err != nil {
		// In a real prod environment a more persistent way of this double write should be enforced
		// if the cache write fails after the DB write succeeds.
		log.Printf("CRITICAL: DB updated but cache failed: %v", err)
		http.Error(w, "Failed to update cache", http.StatusInternalServerError)
		return
	}

	log.Println("Write-Through successful")
	w.WriteHeader(http.StatusOK)
}

// UpdateProductWriteBehind demonstrates the Write-Behind (or "Lazy Writing") pattern.
func (h *handler) UpdateProductWriteBehind(w http.ResponseWriter, r *http.Request) {
	var product Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	cacheKey := "product:" + product.ID

	// Write data to the cache immediately
	productJSON, _ := json.Marshal(product)
	err := h.redis.Set(context.Background(), cacheKey, productJSON, 5*time.Minute).Err()
	if err != nil {
		http.Error(w, "Failed to write to cache", http.StatusInternalServerError)
		return
	}

	// Return success to the user immediately
	log.Println("Write-Behind: Cache updated. Acknowledged to client.")
	w.WriteHeader(http.StatusAccepted) // 202 Accepted is a good status code here

	// Asynchronously write to the database after a delay
	go func() {
		// In a real prod application this need to be a more robust queueing system
		// instead of just a simple timed goroutine.
		time.Sleep(10 * time.Second)

		_, err := h.pg.Exec("UPDATE products SET name = $1, price = $2 WHERE id = $3", product.Name, product.Price, product.ID)
		if err != nil {
			log.Printf("ERROR: Failed write-behind to database: %v", err)
			// Need an error handling/retry mechanism here to ensure consistency
		} else {
			log.Println("Write-Behind: Database updated successfully.")
		}
	}()
}

