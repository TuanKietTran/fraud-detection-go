## MODIFIED Requirements

### Requirement: Operational endpoints
GET /healthz SHALL report process liveness, GET /readyz SHALL probe the selected event store and return 503 when unavailable, and GET /metrics SHALL return process-local requests_total and failures_total (HTTP 5xx), goroutine count, heap allocated/system bytes, cumulative GC count/pause nanoseconds and live TCP connection count. Runtime pprof profiles SHALL be served only to loopback peers at `/debug/pprof/` and documented profile subpaths.

#### Scenario: Runtime resource sampling
- **WHEN** `/metrics` is queried
- **THEN** it includes heap, goroutine, connection and cumulative GC pause counters.
- **AND** pprof is accessible from loopback and returns not found to non-loopback peers.
