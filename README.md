# Weather Service

A small Go HTTP service that returns a short forecast for "Today" and a temperature characterization (`hot | moderate | cold`) for given latitude and longitude using the National Weather Service (NWS) API.

## Quick start

```bash
# Clone and enter
git clone <your-repo> weather-service
cd weather-service

# Set your User-Agent (required by NWS)
export NWS_USER_AGENT="WeatherService/1.0 (you@example.com)"

# Run
make run

# Request
curl "http://localhost:8080/v1/forecast?lat=37.7749&lon=-122.4194"
```

## Run with Docker

```bash
# Set your User-Agent (required by NWS)
export NWS_USER_AGENT="WeatherService/1.0 (you@example.com)"

# Run
docker-compose up --build

# Request
curl "http://localhost:8080/v1/forecast?lat=37.7749&lon=-122.4194"
```

Example response:
```json
{
  "coords": {"lat": 37.7749, "lon": -122.4194},
  "date": "2025-08-13",
  "today": {
    "name": "Today",
    "shortForecast": "Partly Cloudy",
    "temperature": {"value": 72, "unit": "F", "type": "moderate"}
  },
  "source": "api.weather.gov",
  "meta": {"gridId":"MTR","gridX":88,"gridY":126}
}
```

## Configuration

Environment variables (see `.env.example`):

- `PORT` (default `8080`)
- `LOG_LEVEL` (default `INFO`)
- `HTTP_TIMEOUT` (default `5s`)
- `NWS_BASE_URL` (default `https://api.weather.gov`)
- `NWS_USER_AGENT` (**required** by NWS; include contact info)
- `CACHE_TTL` (default `10m`)
- `TEMP_BAND_COLD_MAX` (default `45`)
- `TEMP_BAND_HOT_MIN` (default `85`)

## Build & Run

```bash
make run         # go run ./cmd/weatherd
make test        # go test ./... -race -cover
make build       # build local binary
make docker-build
```

## API

- `GET /v1/forecast?lat=<float>&lon=<float>` — returns today's short forecast and classification.
- `GET /healthz` — liveness probe.

OpenAPI spec: `api/openapi.yaml`.

## Notes

- Uses the NWS discovery pattern: `/points/{lat},{lon}` => `properties.forecast` URL; then GET that URL to obtain periods.
- Caches `/points` lookups and forecast responses in-memory with a simple TTL to avoid hammering the API.
- Requires Go **1.22+** (uses the new stdlib ServeMux patterns like `GET /path`).
- Implements a graceful shutdown with a 5-second timeout.

## Possible Improvements

- More Unit Tests especially for the server
- If we were in prod, we would probably use a database like redis for the cache so that the cache can be shared across 
  multiple instances of the service. 
- CICD with GitHub Actions
