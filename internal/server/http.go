package server

import (
	"encoding/json"
	"errors"
	"log/slog"
	"math"
	"net/http"
	"strconv"

	"weather-service/internal/forecast"
	"weather-service/internal/version"
)

// Handler wires HTTP routes to the forecast service.
type Handler struct {
	log *slog.Logger
	svc forecast.Service
}

// NewHandler creates a new HTTP handler for the weather service.
func NewHandler(log *slog.Logger, svc forecast.Service) *Handler {
	return &Handler{log: log, svc: svc}
}

// Routes returns the HTTP mux with all registered endpoints.
func (h *Handler) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/forecast", h.GetForecast)
	mux.HandleFunc("GET /healthz", h.Health)
	return mux
}

// GetForecast handles GET /v1/forecast returning today's forecast.
func (h *Handler) GetForecast(w http.ResponseWriter, r *http.Request) {
	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")
	lat, lon, err := parseLatLon(latStr, lonStr)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}

	res, err := h.svc.GetTodaysForcast(r.Context(), lat, lon)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}

	writeJSON(w, http.StatusOK, res)
}

// Health handles GET /healthz returning a simple health status.
func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"version": version.Version,
	})
}

func parseLatLon(latStr, lonStr string) (float64, float64, error) {
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil || math.IsNaN(lat) || lat < -90 || lat > 90 {
		return 0, 0, errors.New("invalid lat")
	}
	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil || math.IsNaN(lon) || lon < -180 || lon > 180 {
		return 0, 0, errors.New("invalid lon")
	}
	return lat, lon, nil
}

func writeErr(w http.ResponseWriter, code int, err error) {
	writeJSON(w, code, map[string]any{
		"error": http.StatusText(code),
		"msg":   err.Error(),
	})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
