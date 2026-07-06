package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port                   string
	DBPath                 string
	AdminPassword          string
	CORSOrigin             string
	DeliverySyncInterval   time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Port:          getenv("PORT", "8080"),
		DBPath:        getenv("DB_PATH", "./data/gallup-q14.db"),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
		CORSOrigin:    getenv("CORS_ORIGIN", "http://localhost:5173"),
	}

	intervalHours, err := parseSyncIntervalHours(getenv("DELIVERY_SYNC_INTERVAL_HOURS", "24"))
	if err != nil {
		return Config{}, err
	}
	cfg.DeliverySyncInterval = time.Duration(intervalHours) * time.Hour

	if cfg.AdminPassword == "" {
		return Config{}, fmt.Errorf("ADMIN_PASSWORD is required")
	}

	return cfg, nil
}

func parseSyncIntervalHours(raw string) (int, error) {
	hours, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("DELIVERY_SYNC_INTERVAL_HOURS must be an integer number of hours")
	}
	if hours < 0 {
		return 0, fmt.Errorf("DELIVERY_SYNC_INTERVAL_HOURS must be >= 0")
	}
	return hours, nil
}

func getenv(name, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	return value
}
