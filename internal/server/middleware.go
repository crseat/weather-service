package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type ctxKey int

const reqIDKey ctxKey = 0

// WithMiddleware wraps the provided handler with standard middleware (logging, recovery, request ID).
func WithMiddleware(next http.Handler) http.Handler {
	return requestID(recoverer(logger(next)))
}

// logger logs the request and response.
func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(ww, r)
		reqID := GetRequestID(r.Context())
		slogger := slog.Default()
		slogger.Info("http_request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", reqID,
		)
	})
}

// recoverer recovers from panics and returns a 500 error.
func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Default().Error("panic recovered", "err", rec)
				writeErr(w, http.StatusInternalServerError, http.ErrAbortHandler)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// requestID adds a request ID to the context if not already present.
func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = genID()
		}
		ctx := context.WithValue(r.Context(), reqIDKey, id)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID extracts the request ID from context if present.
func GetRequestID(ctx context.Context) string {
	if v, ok := ctx.Value(reqIDKey).(string); ok {
		return v
	}
	return ""
}

type responseWriter struct {
	http.ResponseWriter

	status int
}

func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// genID generates a random ID using UUID v4.
func genID() string {
	return uuid.NewString()
}
