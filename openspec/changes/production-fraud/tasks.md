# Tasks

- [x] Rename module path.
- [x] Publish initial OpenAPI contract and distinguish implemented vs planned routes.
- [x] Replace float amount with exact amount/currency and validated rules.
- [x] Add internal velocity, geo/device mismatch, new-payee and supplied-structuring rules with invariant tests.
- [x] Persist replay-safe decisions through memory/Postgres ports and expose explanation/config/feedback use cases; DATABASE_URL server selection is implemented and tested.
- [x] Infer near-band structuring patterns from account/window event history; preserve contextual flag only as an additional signal.
- [x] Add real memory/Postgres parity/restart, HTTP/runtime hardening, measured application latency and operational docs.
- [x] Supply CI workflow plus locally tested weighted coverage gate; hosted execution remains UNVERIFIED.

## Open acceptance (proposal remains Partial)

- [ ] Runtime policy editing beyond injected RuleConfig and deterministic feedback (no API/env yet; route contract expansion through manager).
- [ ] Scalable relational SQL and bounded account-history retention/time indexing (current durable snapshot is correctness-first).
- [ ] Complete success schemas and semantic contract validation for explanation/rules.
- [ ] Cold-clone under-two-minute timing and hosted workflow execution evidence.
- [ ] Manager-owned real-process S1/10k fault/reconciliation evidence, not inferred from this service's microtiming/playback.

Do not archive this broader change as fully accepted while these rows are open. Focused implemented milestone changes are archived independently.
