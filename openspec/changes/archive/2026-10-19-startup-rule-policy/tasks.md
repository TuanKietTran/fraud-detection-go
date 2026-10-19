# Tasks

- [x] Add bounded strict FRAUD_RULES_JSON parser with exact integer/default-overlay semantics and sanitized failures.
- [x] Wire InitialRules into memory/SQL initialization and validate active stored state without reset.
- [x] Prove configured HTTP scoring, feedback bounds/idempotency and replay preservation in memory; new memory instance explicitly gets its new seed.
- [x] Extend existing real PostgreSQL fixture using isolated schema: fresh seed affects HTTP scoring, changed/empty seeds preserve learned state/replay across reopen and actual DB restart; invalid incoming/stored policy fails startup.
- [x] ONE native gate pass succeeded: (local gate log, not committed) and card9.cover. Domain76/76=100%, aggregate558/609=91.63%, service91.3%; actual SQL/repository85.354s, no skips. Docs/acceptance reconciled, broader production-fraud remains Partial. Archive focused change, green commit and STATUS boundary only; no further card under urgent stop.
