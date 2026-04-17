package config

import (
	"os"
	"strconv"
)

// Config holds all runtime configuration sourced from environment variables.
type Config struct {
	Port            string
	ProjectID       string
	RateLimitRPS    int // token-bucket refill rate (requests/sec) per IP
	RateLimitBurst  int // maximum burst size per IP
	CacheTTLSeconds int // default TTL for individual-car cache entries
}

func Load() *Config {
	return &Config{
		Port:            getEnv("PORT", "8080"),
		ProjectID:       getEnv("GCP_PROJECT_ID", "cars-api-local"),
		RateLimitRPS:    getEnvInt("RATE_LIMIT_RPS", 10),
		RateLimitBurst:  getEnvInt("RATE_LIMIT_BURST", 20),
		CacheTTLSeconds: getEnvInt("CACHE_TTL_SECONDS", 300),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
