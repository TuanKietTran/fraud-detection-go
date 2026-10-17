# Score event design

- `query.ScoreEventHandler` has no repository. `main` does not construct `InMemoryFraud`.
- Thresholds are strict `>`: 5000 adds `HIGH_AMOUNT` (40), 10000 adds `VERY_HIGH_AMOUNT` (30). `REVIEW` starts at 40. `BLOCK` starts at 70.
- `RiskDecision.ID` is left as the zero UUID on this path.
- `internal/repository/inmemory.go` can save and get a decision by `event_id`, and returns nil, nil when the id is missing. Nothing in the server calls it.
- `migrations/001_init.sql` defines `risk_decisions`. The server does not open Postgres.
