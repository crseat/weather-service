package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"weather-service/internal/cache"
	"weather-service/internal/config"
	"weather-service/internal/forecast"
	logpkg "weather-service/internal/log"
	"weather-service/internal/nws"
	"weather-service/internal/server"
)

const (
	ReadHeaderTimeout = 5 * time.Second
	IdleTimeout       = 60 * time.Second
	contextTimeout    = 5 * time.Second
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

	memCache := cache.NewCache(cfg.CacheTTL)
	svc := forecast.NewService(nwsClient, memCache, forecast.Bands{
		ColdMax: cfg.ColdMax,
		HotMin:  cfg.HotMin,
	})

	h := server.NewHandler(logger, svc)
	mux := h.Routes()

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           server.WithMiddleware(mux),
		ReadHeaderTimeout: ReadHeaderTimeout,
		IdleTimeout:       IdleTimeout,
	}

	go func() {
		logger.Info("starting weather-service", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server startup error", "err", err)
			os.Exit(1)
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("initiating graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server shutdown error", "err", err)
	} else {
		logger.Info("server shutdown complete")
	}
}
