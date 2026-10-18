# ADR 003: Separate liveness, readiness and bounded serving

Status: accepted.

Liveness only proves the process is serving. Readiness probes the actual selected event repository, so a lost PostgreSQL connection cannot remain falsely ready. Metrics are process-local JSON counters, not an external monitoring integration.

Every route is wrapped in a response deadline that also cancels its context. API operations may return 408 on cancellation; if the response budget expires first, the outer timeout returns 503 with the same JSON error shape and a server-generated request ID. A timed-out write can have committed; consumers must retry the same event identity, never create a new identity to retry. No claim is made that timeout rolls back a committed transaction.

Read-header, total read, write and idle limits bound HTTP resource use. Signal cancellation initiates graceful shutdown; failure to drain within its deadline forces connection closure. Database initialization is separately bounded. DATABASE_URL is optional and is never logged; absence explicitly selects non-durable memory.

There are no downstream HTTP calls or downstream writes in this service, so outbox, circuit-breaker and downstream retry policies do not apply here. PostgreSQL pool calls use request contexts. The singleton snapshot adapter still serializes full-history reads and writes; readiness itself takes that lock, which is a documented scalability limit rather than a production throughput claim.

Tests use an injected TCP listener and channels, without synchronization sleeps. CI workflow execution, production deployment, backup/restore and production restart remain unverified.
