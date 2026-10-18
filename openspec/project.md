# Fraud detection

Exact-money weighted scoring with transactional process-memory replay, explanations, feedback and account-window velocity. Optional PostgreSQL persistence and bounded operational runtime are implemented; scalability limitations and final operational documentation remain. Inferred structuring and scoped-memory application latency are now measured/tested.

## Stack

Go module target 1.25 (dependency graph); verified using Go 1.27.1 darwin/arm64. Chi, Google UUID. Money is parsed directly into integer USD cents without float conversion.

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
| 2026-10-18 | Transactional account-window velocity | `ff6b560`; `go test -p 2 ./... -race -count=1 -coverprofile=<profile>`; domain 91.9%, weighted non-cmd 81.65%. |
| 2026-10-18 | PostgreSQL adapter parity and database restart | `3320761`; `go test -p 2 ./internal/repository -run TestPostgresEventParity -race -count=1 -v` passed with PostgreSQL 16.9. Card hash recorded in append-only STATUS. |

## HTTP boundary checkpoint

Starting commit `3320761`; `GOTOOLCHAIN=local FRAUD_POSTGRES_CACHE=<cache> go test -p 2 ./... -race -count=1 -coverprofile=<profile>`: API 95.0%, domain 91.9%, query 95.6%, repository 88.9%, playback 61.5%; weighted non-cmd 325/367 = 88.56%. Build/vet/tidy/format pass. Exact commit hash and full log path in STATUS-fraud.md.

## Runtime operations checkpoint

Starting commit `4293c22`; native `GOTOOLCHAIN=local FRAUD_POSTGRES_CACHE=<cache> go test -p 2 ./... -race -count=1 -coverprofile=<profile>` passes with real PostgreSQL DATABASE_URL selection. Domain 91.9%, API 95.0%, query 95.6%, repository 88.9%, service 90.9%, playback 61.5%; weighted non-cmd 445/499 = 89.18%. Build/vet/tidy/format pass; `ctl/logs/fraud/card5b-gates.log`; card commit hash in append-only STATUS.

## Pending mandatory work

- Rule configuration updates beyond validated injected policies and deterministic feedback learning.
- Hiring-manager README, CI/DoD completion and final OpenSpec reconciliation.
- Scalable PostgreSQL/per-account retention remain explicit limits, not completed production capabilities.

## Inferred structuring / scoped memory checkpoint

Starting commit `26757f9`; native build/vet/race/tidy/format gates passed (`ctl/logs/fraud/card5c-gates.log`), including real PostgreSQL inferred-pattern parity (91.593s). Domain92.1%, weighted non-cmd504/558=90.32%. `go test ./internal/domain -run '^$' -fuzz '^FuzzStructuringBand$' -fuzztime=2s -parallel=2`:145053 executions. `go run ./cmd/latency -samples 10000 -raw`: unique p99=5459ns, replay p99=4209ns, memory application handler only. Environment/scope/raw report in `docs/performance.md`. Card hash recorded in append-only STATUS.

## Native macOS continuation checkpoint

Starting commit `8d4ceab`; card account-window-velocity adds repository transactions and internal velocity. Evidence: `GOTOOLCHAIN=local go test -p 2 ./... -race -count=1 -coverprofile=<profile>` in `ctl/logs/fraud/card4c-gates.log`; domain 91.9%, weighted non-cmd 227/278 = 81.65%. Build/vet/tidy pass. Card commit hash is recorded append-only in STATUS-fraud.md to avoid self-referential commit hashes.

## Definition of done

See `openspec/dod.md`. The playback storyboard is `openspec/gifs/playback-flow.gif`.
