# Tasks

- [x] Validate HTTP_ADDR, DATABASE_URL and bounded request/startup/shutdown durations.
- [x] Select memory or pgx event store and keep configuration/error logs free of connection information.
- [x] Wire signal-aware shutdown with forced close if graceful deadline expires.
- [x] Set header/read/write/idle timeouts and bounded header size.
- [x] Expose health/readiness/metrics and structured request logs; bound all routes including readiness.
- [x] Test real listener serving and cancellation, bad config/listener errors, timeout cancellation, metrics and readiness.
- [x] Extend real embedded PostgreSQL test to prove DATABASE_URL selects persisted learned state.
- [x] Complete native milestone gates: (local gate log, not committed); Go 1.27.1 darwin/arm64; build/vet/race/tidy/format/diff pass. Domain 91.9%, API 95.0%, query 95.6%, repository 88.9%, service 90.9%, playback 61.5%; weighted non-cmd 445/499 = 89.18%. Real PostgreSQL suite 91.554s, no skips. Prior fallback JSON regression failed correctly, then passed after propagating request IDs into timeout buffers.
