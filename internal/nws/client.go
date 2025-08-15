package nws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Client wraps access to the api.weather.gov HTTP API.
type Client struct {
	base   string
	ua     string
	http   *http.Client
	logger *slog.Logger
}

// NewClient constructs a new NWS API client.
func NewClient(baseURL, userAgent string, httpClient *http.Client, logger *slog.Logger) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}
	return &Client{
		base:   strings.TrimRight(baseURL, "/"),
		ua:     userAgent,
		http:   httpClient,
		logger: logger,
	}
}

// Points returns the NWS points metadata for the given latitude and longitude.
func (c *Client) Points(ctx context.Context, lat, lon float64) (PointsResponse, error) {
	var pr PointsResponse
	url := fmt.Sprintf("%s/points/%f,%f", c.base, lat, lon)
	if err := c.doJSON(ctx, http.MethodGet, url, &pr); err != nil {
		return PointsResponse{}, err
	}
	return pr, nil
}

// Forecast gets the forecast document at the provided forecast URL.
func (c *Client) Forecast(ctx context.Context, forecastURL string) (Forecast, error) {
	var f Forecast
	if err := c.doJSON(ctx, http.MethodGet, forecastURL, &f); err != nil {
		return Forecast{}, err
	}
	return f, nil
}

func (c *Client) doJSON(ctx context.Context, method, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return err
	}
	if c.ua == "" {
		return errors.New("nws user agent is required; set NWS_USER_AGENT")
	}
	req.Header.Set("User-Agent", c.ua)
	req.Header.Set("Accept", "application/geo+json")

	var lastErr error
	backoff := 250 * time.Millisecond

	for attempt := 0; attempt < 3; attempt++ {
		resp, doErr := c.http.Do(req)
		if doErr != nil {
			lastErr = doErr
			time.Sleep(backoff)
			backoff *= 2
			continue
		}
		func() {
			defer func(Body io.ReadCloser) {
				closeErr := Body.Close()
				if closeErr != nil {
					c.logger.Error("Error closing body of readCloser in doJson")
				}
			}(resp.Body)
			if resp.StatusCode == http.StatusOK {
				body, readAllErr := io.ReadAll(resp.Body)
				if readAllErr != nil {
					lastErr = readAllErr
					return
				}
				if err3 := json.Unmarshal(body, out); err3 != nil {
					lastErr = err3
					return
				}
				lastErr = nil
				return
			}

			// Retry on 429/503 using Retry-After when present.
			if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
				delay := backoff
				if ra := parseRetryAfter(resp.Header.Get("Retry-After")); ra > 0 {
					delay = ra
				}
				c.logger.Warn("nws throttled, retrying", "status", resp.StatusCode, "delay", delay, "url", url)
				time.Sleep(delay)
				backoff *= 2
				lastErr = fmt.Errorf("nws throttled: %s", resp.Status)
				return
			}

			// Non-retryable
			b, _ := io.ReadAll(resp.Body)
			lastErr = fmt.Errorf("nws http %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
		}()
		if lastErr == nil {
			return nil
		}
	}
	return lastErr
}

func parseRetryAfter(v string) time.Duration {
	if v == "" {
		return 0
	}
	// seconds
	if n, err := strconv.Atoi(v); err == nil && n >= 0 {
		return time.Duration(n) * time.Second
	}
	// HTTP date
	if t, err := http.ParseTime(v); err == nil {
		d := time.Until(t)
		if d > 0 {
			return d
		}
	}
	return 0
}
