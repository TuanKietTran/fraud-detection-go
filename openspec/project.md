# Fraud detection

CQRS scoring query. The HTTP API scores one event from its amount and returns allow, review, or block. It does not persist the decision.

## Stack

Go 1.22, Chi, Google UUID. Money on this API is a JSON number, not `decimal`.

## Public API

- `POST /v1/score`

## Checkpoints

| Date | Change | Evidence |
| --- | --- | --- |
| 2026-10-17 | Baseline spec recorded from the shipped service | `1a9c732`. No handler tests and no Postgres repository yet. |
| 2026-10-17 | 1000 requests in one minute, retry, playback, coverage | `internal/playback`. 1000 scores plus one 503 retry each finished in 77ms. Replay returned the same decisions, including allow, review, and block. Statement coverage 55.9%. |

## Out of scope until implemented

- `RecordDecision`, `UpsertRule`, and `GetDecision` are named in the README. The process does not call `InMemoryFraud`.
- `migrations/001_init.sql` is not applied by the server.
- Velocity and other non-amount rules.
