# Tasks

- [x] Add pgx-backed transactional event-state adapter.
- [x] Add memory/Postgres parity and actual database restart test.
- [x] Execute parity test on macOS: `go test -p 2 ./internal/repository -run TestPostgresEventParity -race -count=1 -v` passed, PostgreSQL 16.9 (93.36s including download/start/restart).
- [x] Run full native milestone gates, archive this change and publish evidence: (local gate log, not committed); repository 88.9%, domain 91.9%, weighted non-cmd 272/318 = 85.53%; build/vet/race/tidy/format passed using Go 1.27.1 darwin/arm64.
