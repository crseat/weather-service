# Architecture

**Flow:**

1. `GET /v1/forecast?lat=..&lon=..`
2. `internal/forecast.Service.Today`:
   - Resolve NWS forecast URL via `GET /points/{lat},{lon}` (cached).
   - Fetch forecast at that URL (cached).
   - Select *Today's* period (`name == "Today"` or first daytime period on today's local date).
   - Classify temperature using configured bands.
3. Respond JSON.

**Caching:**

- In-memory TTL cache (default 10m) keyed by:
  - `points:<lat>,<lon>` → forecast URL (string)
  - `forecast:<url>` → parsed forecast struct

**Configuration:**

See `.env.example`. `NWS_USER_AGENT` is required and must include contact info.

**Operational:**

- Health endpoint at `/healthz`.
- Sane server timeouts.
- Structured logs via `slog` with request id.
