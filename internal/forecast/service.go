package forecast

import (
	"context"
	"errors"
	"fmt"
	"time"

	"weather-service/internal/cache"
	"weather-service/internal/nws"
)

const (
	source = "api.weather.gov"
)

// Service provides forecast data operations.
type Service interface {
	// GetTodaysForcast returns a summarized forecast result for the given coordinates.
	GetTodaysForcast(ctx context.Context, lat, lon float64) (Result, error)
}

type service struct {
	client *nws.Client
	cache  *cache.Memory
	bands  Bands
}

// NewService constructs a forecast Service using the given NWS client, cache, and bands.
func NewService(client *nws.Client, cache *cache.Memory, bands Bands) Service {
	return &service{client: client, cache: cache, bands: bands}
}

// Result is the API response payload returned by the forecast service for Today.
type Result struct {
	Coords struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"coords"`
	Date  string `json:"date"`
	Today struct {
		Name          string `json:"name"`
		ShortForecast string `json:"shortForecast"`
		Temperature   struct {
			Value int    `json:"value"`
			Unit  string `json:"unit"`
			Type  string `json:"type"` // hot|moderate|cold
		} `json:"temperature"`
	} `json:"today"`
	Source string      `json:"source"`
	Meta   interface{} `json:"meta,omitempty"`
}

// GetTodaysForcast resolves the NWS grid point for the given lat/lon, fetches (with caching)
// the associated forecast, selects today's period relative to the current time, and
// returns a summarized Result. It classifies the temperature using the configured
// Bands (hot/moderate/cold) and includes the upstream document's update time in
// Result.Meta["updated"].
//
// Caching:
//   - points: maps lat/lon -> forecast URL
//   - forecast: caches the full forecast document
//
// Errors are returned when the point has no forecast URL, when no usable forecast
// periods are available for today, or when upstream calls fail.
func (s *service) GetTodaysForcast(ctx context.Context, lat, lon float64) (Result, error) {
	pointsKey := fmt.Sprintf("points:%.4f,%.4f", lat, lon)
	var forecastURL string
	if v, ok := s.cache.Get(pointsKey); ok {
		forecastURL, _ = v.(string)
	} else {
		pts, err := s.client.Points(ctx, lat, lon)
		if err != nil {
			return Result{}, err
		}
		forecastURL = pts.Properties.Forecast
		if forecastURL == "" {
			return Result{}, errors.New("no forecast URL for point")
		}
		s.cache.Set(pointsKey, forecastURL)
	}

	fcKey := "forecast:" + forecastURL
	var fc nws.Forecast
	if v, ok := s.cache.Get(fcKey); ok {
		if cached, ok2 := v.(nws.Forecast); ok2 {
			fc = cached
		}
	}
	if len(fc.Properties.Periods) == 0 {
		var err error
		fc, err = s.client.Forecast(ctx, forecastURL)
		if err != nil {
			return Result{}, err
		}
		s.cache.Set(fcKey, fc)
	}

	now := time.Now()
	period, ok := nws.SelectToday(fc.Properties.Periods, now)
	if !ok {
		return Result{}, errors.New("no forecast periods available")
	}

	var res Result
	res.Source = source
	res.Coords.Lat, res.Coords.Lon = lat, lon
	res.Date = period.StartTime.Format("2006-01-02")
	res.Today.Name = period.Name
	res.Today.ShortForecast = period.ShortForecast
	res.Today.Temperature.Value = period.Temperature
	res.Today.Temperature.Unit = period.TemperatureUnit
	res.Today.Temperature.Type = Classify(period.Temperature, s.bands)

	// Include some useful meta
	res.Meta = map[string]any{
		"updated": fc.Properties.Updated.Format(time.RFC3339),
	}

	return res, nil
}
