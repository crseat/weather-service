package server_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"weather-service/internal/server"
)

func TestRequestIDGeneratedAndPropagated(t *testing.T) {
	// Handler echoes the request ID from context in a response header
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := server.GetRequestID(r.Context())
		if id == "" {
			t.Fatalf("request id missing in context")
		}
		w.Header().Set("X-Echo-Request-Id", id)
		w.WriteHeader(http.StatusOK)
	})

	h := server.WithMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d want %d", rec.Code, http.StatusOK)
	}

	gen := rec.Header().Get("X-Request-ID")
	echo := rec.Header().Get("X-Echo-Request-Id")
	if gen == "" {
		t.Fatalf("X-Request-ID header not set")
	}
	if gen != echo {
		t.Fatalf("context/header mismatch: %q != %q", echo, gen)
	}
}

func TestRequestIDUsesProvidedHeader(t *testing.T) {
	want := "req-12345"
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := server.GetRequestID(r.Context())
		if got != want {
			t.Fatalf("GetRequestID=%q want %q", got, want)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	h := server.WithMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", want)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status=%d want %d", rec.Code, http.StatusNoContent)
	}
	if got := rec.Header().Get("X-Request-ID"); got != want {
		t.Fatalf("X-Request-ID header=%q want %q", got, want)
	}
}

func TestLoggerWritesStatusAndPath(t *testing.T) {
	var buf bytes.Buffer
	// Capture logs from middleware.logger via default logger
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	h := server.WithMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/log-test", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d want %d", rec.Code, http.StatusCreated)
	}

	logOut := buf.String()
	if !strings.Contains(logOut, "status=201") {
		t.Fatalf("log does not contain status: %s", logOut)
	}
	if !strings.Contains(logOut, "path=/log-test") {
		t.Fatalf("log does not contain path: %s", logOut)
	}
}

func TestRecovererHandlesPanic(t *testing.T) {
	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("boom")
	})

	h := server.WithMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d want %d", rec.Code, http.StatusInternalServerError)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Fatalf("content-type=%q want application/json; charset=utf-8", ct)
	}

	var m map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if m["error"] != http.StatusText(http.StatusInternalServerError) {
		t.Fatalf("error field = %v", m["error"])
	}
}
