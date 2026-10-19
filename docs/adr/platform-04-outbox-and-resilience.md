# Platform ADR 04: outbox and downstream resilience

Status: not applicable to this service.

Fraud does not call a downstream HTTP service or trigger a downstream write, so it has no outbox, downstream retry policy or circuit breaker. The playback client retries its own HTTP requests with bounded exponential backoff and full jitter; this does not imply a downstream dependency exists. See [ADR 003](003-runtime-deadlines-and-readiness.md).
