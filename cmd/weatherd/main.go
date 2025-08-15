package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"weather-service/internal/cache"
	"weather-service/internal/config"
	"weather-service/internal/forecast"
	logpkg "weather-service/internal/log"
	"weather-service/internal/nws"
	"weather-service/internal/server"
)

func main() {
	cfg := config.FromEnv()

	logger := logpkg.New(cfg.LogLevel)
	slog.SetDefault(logger)

	if cfg.NWSUserAgent == "" {
		logger.Error("NWS_USER_AGENT is required (include contact info)")
		os.Exit(1)
	}

	httpClient := &http.Client{Timeout: cfg.HTTPTimeout}
	nwsClient := nws.NewClient(cfg.NWSBaseURL, cfg.NWSUserAgent, httpClient, logger)

	memCache := cache.NewMemory(cfg.CacheTTL)
	svc := forecast.NewService(nwsClient, memCache, forecast.Bands{
		ColdMax: cfg.ColdMax,
		HotMin:  cfg.HotMin,
	})

	h := server.NewHandler(logger, svc)
	mux := h.Routes()

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           server.WithMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	logger.Info("starting weather-service", "port", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server error", "err", err)
		os.Exit(1)
	}
}
