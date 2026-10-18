# Runtime operations and durable-store selection

Replace bare ListenAndServe with validated environment configuration, bounded HTTP/server/database startup deadlines, graceful signal shutdown and structured logs. Select the already parity-tested PostgreSQL adapter via DATABASE_URL; unset remains explicitly non-durable memory. Expose /healthz, /readyz and JSON /metrics counters; readiness probes the selected event store.

Use an injected listener for real HTTP shutdown tests without synchronization sleeps. Bound every route, including readiness, with TimeoutHandler and a consistent JSON timeout error. Distinguish liveness from dependency readiness. Never log a database URL or connection error that could expose credentials.

No downstream write occurs in fraud, so downstream outbox/retry/circuit-breaker machinery is not applicable. PostgreSQL snapshot scalability, inferred structuring and measured latency remain subsequent work; this change does not make a full production-readiness claim.
