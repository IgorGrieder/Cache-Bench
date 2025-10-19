package config

import (
	"os"
	"strconv"
)

type Config struct {
	PORT       int
	REDIS_ADDR string
	REDIS_PORT int
	HOST       string
	PORT_PG    int
	USER       string
	DB_NAME    string
	PG_PASS    string
	HASH       string
}

func NewConfig() *Config {
	port := parseInt(getEnv("PORT", "8080"))
	reddisAddr := getEnv("REDIS_ADDR", "redis")
	reddisPort := parseInt(getEnv("REDIS_PORT", "6379"))
	host := getEnv("PG_HOST", "localhost")
	portPG := parseInt(getEnv("PG_PORT", "5432"))
	user := getEnv("PG_USER", "postgres")
	dbname := getEnv("PG_DB", "bench")
	pgPass := getEnv("PG_PASS", "none")
	hash := getEnv("HASH", "none")

	return &Config{
		PORT:       port,
		REDIS_ADDR: reddisAddr,
		REDIS_PORT: reddisPort,
		HOST:       host,
		PORT_PG:    portPG,
		USER:       user,
		DB_NAME:    dbname,
		PG_PASS:    pgPass,
		HASH:       hash,
	}
}

func getEnv(key string, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
