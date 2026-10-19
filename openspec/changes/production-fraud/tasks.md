# Tasks

- [x] Rename module path.
- [x] Publish initial OpenAPI contract and distinguish implemented vs planned routes.
- [x] Replace float amount with exact amount/currency and validated rules.
- [x] Add internal velocity, geo/device mismatch, new-payee and supplied-structuring rules with invariant tests.
- [x] Persist replay-safe decisions through memory/Postgres ports and expose explanation/config/feedback use cases; DATABASE_URL server selection is implemented and tested.
- [x] Infer near-band structuring patterns from account/window event history; preserve contextual flag only as an additional signal.
- [x] Add real memory/Postgres parity/restart, HTTP/runtime hardening, measured application latency and operational docs.
- [x] Supply CI workflow plus locally tested portable Go weighted coverage gate; hosted execution remains UNVERIFIED.
- [x] Implement R1 playback through production HTTP router: 10,000 logical calls, 2% same-event duplicates, >=5% transient 503/429, 1% per-attempt timeouts, bounded concurrency, six-attempt full-jitter exponential retries; assert all calls terminal and one decision per unique event identity.
- [x] Expose runtime heap/goroutine/GC-pause/TCP-connection sampling and loopback-only pprof URLs; test response fields and external-peer denial.

## Open acceptance (proposal remains Partial)

- [x] Validated executable startup policy via bounded strict FRAUD_RULES_JSON, wired to memory/SQL scoring; existing durable learned policy/replay state takes precedence. Manager acceptance requires no online unauthenticated admin write API.
- [ ] Scalable relational SQL and bounded account-history retention/time indexing (current durable snapshot is correctness-first).
- [x] Complete explanation/rules success schemas and targeted response shape/reference checks for every existing v1 status.
- [ ] Full semantic OpenAPI/request-schema validation beyond the explicitly supported response subset.
- [ ] Cold-clone under-two-minute timing and hosted workflow execution evidence.
- [ ] Manager-owned real-process S1/10k fault/reconciliation evidence, not inferred from this service's microtiming/playback.

Do not archive this broader change as fully accepted while these rows are open. Focused implemented milestone changes are archived independently.
