# Score event design

- `cmd/server` only loads validated environment configuration, configures JSON slog and signal handling, and invokes `internal/service`.
- `service` selects `MemoryEvents` or `PostgresEvents` with DATABASE_URL, exposes operational routes and bounded HTTP lifecycle, and mounts `api` under `/v1`.
- API parses exact USD cents without floats and maps typed application errors to documented JSON envelopes. Request IDs are authoritative server-generated values. Timeout buffering explicitly propagates the ID so JSON bodies and headers agree.
- `query.ScoreEventHandler` serializes replay lookup, prior-account velocity, pure weighted domain scoring and result storage in one repository transaction. Injected clocks define (now-60s, now] velocity. Same event identity/payload preserves its original result; conflicting payloads retain the legacy HTTP400 status.
- Feedback commits its identity and bounded learned review threshold together. Explanations return defensive copies. Domain DENY is emitted as BLOCK on the legacy score response; explanation currently returns canonical DENY. Consumers must accept the documented enum aliases.
- Memory state is process-lifetime only. PostgreSQL uses a singleton locked JSONB row (`fraud_state`) retaining inputs, decisions, timestamps, rules and feedback. Constructor DDL initializes this adapter, not the unused legacy `migrations/001_init.sql` risk_decisions table.
- Full-state cloning/loading and global snapshot serialization are known scalability limits. No latency or unbounded-history claim is made. Inferred structuring and measured latency remain pending; structuring is currently a supplied flag.
- The legacy `InMemoryFraud` SaveDecision/GetDecision adapter remains for compatibility regression tests and is not the runtime store. `RiskDecision.ID` remains zero on the scoring path; event_id is its identity.
