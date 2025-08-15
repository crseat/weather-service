package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"weather-service/internal/forecast"
	"weather-service/internal/server"
	"weather-service/internal/version"
)

type fakeSvc struct {
	res    forecast.Result
	err    error
	gotLat float64
	gotLon float64
}

func (f *fakeSvc) GetTodaysForcast(_ context.Context, lat, lon float64) (forecast.Result, error) {
	f.gotLat, f.gotLon = lat, lon
	return f.res, f.err
}

func newHandlerWithFake(t *testing.T, f *fakeSvc) *server.Handler {
	t.Helper()
	return server.NewHandler(nil, f)
}

func decodeBody[T any](t *testing.T, b []byte) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatalf("json decode failed: %v; body=%s", err, string(b))
	}
	return v
}

func TestHealthz(t *testing.T) {
	h := newHandlerWithFake(t, &fakeSvc{})
	mux := h.Routes()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Fatalf("content-type=%q want application/json; charset=utf-8", ct)
	}

	body := rec.Body.Bytes()
	got := decodeBody[map[string]any](t, body)
	if got["status"] != "ok" {
		t.Fatalf("status field = %v want ok", got["status"])
	}
	if got["version"] != version.Version {
		t.Fatalf("version field = %v want %v", got["version"], version.Version)
	}
}

func TestGetForecast_Success(t *testing.T) {
	fake := &fakeSvc{res: forecast.Result{Source: "testsrc"}}
	fake.res.Coords.Lat = 10
	fake.res.Coords.Lon = 20
	fake.res.Date = "2025-08-15"
	fake.res.Today.Name = "Today"
	fake.res.Today.ShortForecast = "Sunny"
	fake.res.Today.Temperature.Value = 75
	fake.res.Today.Temperature.Unit = "F"
	fake.res.Today.Temperature.Type = "moderate"

	h := newHandlerWithFake(t, fake)
	mux := h.Routes()

	req := httptest.NewRequest(http.MethodGet, "/v1/forecast?lat=10&lon=20", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d want %d", rec.Code, http.StatusOK)
	}
	got := decodeBody[forecast.Result](t, rec.Body.Bytes())
	if got.Source != fake.res.Source || got.Date != fake.res.Date {
		t.Fatalf("unexpected body: got=%+v want=%+v", got, fake.res)
	}
	if fake.gotLat != 10 || fake.gotLon != 20 {
		t.Fatalf("service received lat/lon = (%v,%v), want (10,20)", fake.gotLat, fake.gotLon)
	}
}

func TestGetForecast_BadParams(t *testing.T) {
	h := newHandlerWithFake(t, &fakeSvc{})
	mux := h.Routes()

	cases := []string{
		"/v1/forecast?lat=abc&lon=20",          // bad lat
		"/v1/forecast?lat=91&lon=20",           // out of range lat
		"/v1/forecast?lat=10&lon=181",          // out of range lon
		"/v1/forecast?lat=10&lon=not-a-number", // bad lon
	}
	for _, u := range cases {
		req := httptest.NewRequest(http.MethodGet, u, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("%s: status=%d want 400", u, rec.Code)
		}
		// Body should be JSON with error fields
		m := decodeBody[map[string]any](t, rec.Body.Bytes())
		if m["error"] != http.StatusText(http.StatusBadRequest) {
			t.Fatalf("%s: error field = %v", u, m["error"])
		}
		if _, ok := m["msg"].(string); !ok {
			t.Fatalf("%s: msg field missing or not string", u)
		}
	}
}

func TestGetForecast_UpstreamError(t *testing.T) {
	fake := &fakeSvc{err: context.DeadlineExceeded}
	h := newHandlerWithFake(t, fake)
	mux := h.Routes()

	req := httptest.NewRequest(http.MethodGet, "/v1/forecast?lat=1&lon=2", bytes.NewReader(nil))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status=%d want %d", rec.Code, http.StatusBadGateway)
	}
	m := decodeBody[map[string]any](t, rec.Body.Bytes())
	if m["error"] != http.StatusText(http.StatusBadGateway) {
		t.Fatalf("error field = %v", m["error"])
	}
}
