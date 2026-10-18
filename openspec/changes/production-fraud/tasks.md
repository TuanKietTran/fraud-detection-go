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

- [x] Validated executable startup policy via bounded strict FRAUD_RULES_JSON, wired to memory/SQL scoring; existing durable learned policy/replay state takes precedence. Manager acceptance requires no online unauthenticated admin write API.
- [ ] Scalable relational SQL and bounded account-history retention/time indexing (current durable snapshot is correctness-first).
- [x] Complete explanation/rules success schemas and targeted response shape/reference checks for every existing v1 status.
- [ ] Full semantic OpenAPI/request-schema validation beyond the explicitly supported response subset.
- [ ] Cold-clone under-two-minute timing and hosted workflow execution evidence.
- [ ] Manager-owned real-process S1/10k fault/reconciliation evidence, not inferred from this service's microtiming/playback.

Do not archive this broader change as fully accepted while these rows are open. Focused implemented milestone changes are archived independently.
