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

// GetProductCacheAside demonstrates the Cache-Aside pattern for reads.
func (h *handler) GetProductCacheAside(w http.ResponseWriter, r *http.Request) {
	// Let's assume we get the product ID from the query, e.g., /product?id=123
	productID := r.URL.Query().Get("id")
	if productID == "" {
		http.Error(w, "Product ID is required", http.StatusBadRequest)
		return
	}

	// 1. Look for data in the cache (Cache Hit/Miss)
	cacheKey := "product:" + productID
	cachedProduct, err := h.redis.Get(context.Background(), cacheKey).Bytes()

	// 2. Cache Hit: Data is found in Redis
	if err == nil {
		log.Println("Cache HIT")
		w.Header().Set("Content-Type", "application/json")
		w.Write(cachedProduct)
		return
	}

	// 3. Cache Miss: Data is not in Redis
	log.Println("Cache MISS")
	var product Product
	// Query the primary database (PostgreSQL)
	err = h.pg.QueryRow("SELECT id, name, price FROM products WHERE id = $1", productID).Scan(&product.ID, &product.Name, &product.Price)
	if err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	// 4. Populate the cache for next time
	productJSON, _ := json.Marshal(product)
	// Set with a TTL (Time-To-Live) of 5 minutes to avoid stale data
	h.redis.Set(context.Background(), cacheKey, productJSON, 5*time.Minute)

	// Return the response
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

	// 1. Write to the database first (or start a transaction)
	_, err := h.pg.Exec("UPDATE products SET name = $1, price = $2 WHERE id = $3", product.Name, product.Price, product.ID)
	if err != nil {
		http.Error(w, "Failed to update database", http.StatusInternalServerError)
		return
	}

	// 2. Then, write the same data to the cache
	productJSON, _ := json.Marshal(product)
	err = h.redis.Set(context.Background(), cacheKey, productJSON, 5*time.Minute).Err()
	if err != nil {
		// In a real scenario, you'd need a rollback/compensation strategy
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

	// 1. Write data to the cache immediately
	productJSON, _ := json.Marshal(product)
	err := h.redis.Set(context.Background(), cacheKey, productJSON, 5*time.Minute).Err()
	if err != nil {
		http.Error(w, "Failed to write to cache", http.StatusInternalServerError)
		return
	}

	// 2. Return success to the user immediately
	log.Println("Write-Behind: Cache updated. Acknowledged to client.")
	w.WriteHeader(http.StatusAccepted) // 202 Accepted is a good status code here

	// 3. Asynchronously write to the database after a delay
	go func() {
		// In a real application, this might be a more robust queueing system
		// instead of just a simple timed goroutine.
		time.Sleep(10 * time.Second) // Simulate a delay

		_, err := h.pg.Exec("UPDATE products SET name = $1, price = $2 WHERE id = $3", product.Name, product.Price, product.ID)
		if err != nil {
			log.Printf("ERROR: Failed write-behind to database: %v", err)
			// Need an error handling/retry mechanism here
		} else {
			log.Println("Write-Behind: Database updated successfully.")
		}
	}()
}
