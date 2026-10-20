package config

import (
	"os"
	"time"
)

func Env(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func Retry(attempt int) time.Duration {
	if attempt > 6 {
		attempt = 6
	}
	return time.Duration(attempt+1) * time.Second
}
