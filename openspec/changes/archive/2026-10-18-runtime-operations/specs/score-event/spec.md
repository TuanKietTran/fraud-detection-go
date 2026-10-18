## ADDED Requirements

### Requirement: Validated bounded runtime
The service SHALL validate HTTP_ADDR (default :8103), optional PostgreSQL DATABASE_URL and REQUEST_TIMEOUT/STARTUP_TIMEOUT/SHUTDOWN_TIMEOUT durations (defaults 5s/10s/10s, positive and at most 1m). Every HTTP route SHALL have a bounded response deadline. Server headers/read/write/idle operations and database startup SHALL be bounded. Signals SHALL trigger graceful shutdown, followed by forced close if its deadline expires. Logs SHALL be structured and omit database connection information.

#### Scenario: Native HTTP lifecycle
- **GIVEN** a server with an injected TCP listener
- **WHEN** readiness is requested then the service context is canceled
- **THEN** readiness is 200 and serving terminates without a synchronization sleep.

### Requirement: Operational endpoints
GET /healthz SHALL report process liveness, GET /readyz SHALL probe the selected event store and return 503 if unavailable, and GET /metrics SHALL return process-local requests_total and failures_total (5xx) counters.

#### Scenario: Store unavailable
- **GIVEN** a repository returning an error
- **WHEN** health and readiness are queried
- **THEN** health is 200 and readiness is 503.
