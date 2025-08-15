// Package config loads application configuration from environment variables.
package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config represents runtime configuration settings for the service.
// Values are typically sourced from environment variables; see FromEnv.
type Config struct {
	Port         string        // HTTP port to listen on
	LogLevel     slog.Level    // Log level (DEBUG|INFO|WARN|ERROR)
	HTTPTimeout  time.Duration // Timeout for outbound HTTP requests
	NWSBaseURL   string        // Base URL for api.weather.gov
	NWSUserAgent string        // Required User-Agent for NWS requests
	CacheTTL     time.Duration // In-memory cache TTL
	ColdMax      int           // Max Temperature in Fahrenheit to be considered "cold"
	HotMin       int           // Min Temperature in Fahrenheit to be considered "hot"
}

// FromEnv builds a Config from environment variables, applying sensible defaults.
func FromEnv() Config {
	return Config{
		Port:         getenv("PORT", "8080"),
		LogLevel:     parseLevel(getenv("LOG_LEVEL", "INFO")),
		HTTPTimeout:  parseDur(getenv("HTTP_TIMEOUT", "5s"), 5*time.Second),
		NWSBaseURL:   getenv("NWS_BASE_URL", "https://api.weather.gov"),
		NWSUserAgent: getenv("NWS_USER_AGENT", ""),
		CacheTTL:     parseDur(getenv("CACHE_TTL", "10m"), 10*time.Minute),
		ColdMax:      parseInt(getenv("TEMP_BAND_COLD_MAX", "45"), 45),
		HotMin:       parseInt(getenv("TEMP_BAND_HOT_MIN", "85"), 85),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func parseDur(s string, def time.Duration) time.Duration {
	if d, err := time.ParseDuration(s); err == nil {
		return d
	}
	return def
}

func parseInt(s string, def int) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return def
}

func parseLevel(s string) slog.Level {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
