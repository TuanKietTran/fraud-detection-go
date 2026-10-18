# Fraud detection

Exact-money weighted scoring with transactional process-memory replay, explanations, feedback and account-window velocity. PostgreSQL persistence and operational hardening remain unfinished.

## Stack

Go module target 1.22; verified using Go 1.27.1 darwin/arm64. Chi, Google UUID. Money is parsed directly into integer USD cents without float conversion.

## Public API

- `POST /v1/score`
- `GET /v1/events/{event_id}/explanation`
- `GET /v1/rules`
- `POST /v1/feedback`

Authoritative schemas and compatibility aliases: `api/openapi.yaml`.

## Checkpoints

| Date | Change | Evidence |
| --- | --- | --- |
| 2026-10-17 | Baseline spec recorded from the shipped service | `1a9c732`. No handler tests and no Postgres repository yet. |
| 2026-10-17 | 1000 requests in one minute, retry, playback, coverage | `internal/playback`. 1000 scores plus one 503 retry each finished in 77ms. Replay returned the same decisions, including allow, review, and block. Statement coverage 55.9%. |
| 2026-10-17 | Definition of done and playback storyboard | `openspec/dod.md` and `openspec/gifs/playback-flow.gif`. GIF review: retry stays at the pre-write state, playback shows the unchanged result. |

## Pending mandatory work

- Durable PostgreSQL adapter/parity, server-selected repository and configuration updates.
- HTTP JSON errors, request IDs, readiness/metrics and runtime hardening.
- Inferred structuring patterns (current contextual flag is supplied), measured p99 and operational docs.

## Native macOS continuation checkpoint

Starting commit `8d4ceab`; card account-window-velocity adds repository transactions and internal velocity. Evidence: `GOTOOLCHAIN=local go test -p 2 ./... -race -count=1 -coverprofile=<profile>` in `ctl/logs/fraud/card4c-gates.log`; domain 91.9%, weighted non-cmd 227/278 = 81.65%. Build/vet/tidy pass. Card commit hash is recorded append-only in STATUS-fraud.md to avoid self-referential commit hashes.

## Definition of done

See `openspec/dod.md`. The playback storyboard is `openspec/gifs/playback-flow.gif`.
