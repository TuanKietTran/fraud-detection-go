# Fraud detection project

Exact-money synthetic risk scoring with pure weighted rules, replay-safe account-window history, inferred structuring, immutable explanations and deterministic feedback learning. The process selects indexed scoped memory or a durable PostgreSQL snapshot adapter. Runtime has validated environment, bounded serving/shutdown, JSON logs and operational endpoints.

## Stack and public surface

Module `github.com/TuanKietTran/fraud-detection-go`, minimum Go1.25; native evidence uses Go1.27.1 darwin/arm64. Chi/UUID, pgx for optional SQL, embedded PostgreSQL for real parity tests, yaml.v3 for parsed contracts/workflow tests. Money is integer USD cents parsed directly from decimal text, never through float64.

- POST `/v1/score`
- GET `/v1/events/{event_id}/explanation`
- GET `/v1/rules`
- POST `/v1/feedback`
- GET `/healthz`, `/readyz`, `/metrics`

Authoritative routes/statuses: `api/openapi.yaml`. Current semantics: `openspec/specs/score-event/{spec,design}.md`. Run/compatibility/config: `README.md`.

## Checkpoints

Commands verify the named capability at the current tree; hashes identify its milestone. Historical Windows notes do not establish current native/runtime performance. Full current gates are below.

| Date | Milestone / local commit | Verification |
| --- | --- | --- |
| 2026-10-17 | Baseline OpenSpec, `6945a39`; playback, `e6e15f3`; historical DoD/storyboard, `f456258` | `go test ./internal/playback -run TestThousandPerMinuteAndPlayback -count=1 -v`; preserved1000-call httptest retry/replay regression, not real-process S1 |
| 2026-10-18 | Mechanical module rename, `e209c02` | `go list -m` and `go build ./...` |
| 2026-10-18 | Initial contract, `1c250dd`; exact cents/parser/HTTP wiring, `ba64df0`, `08874cc` | `go test ./internal/domain ./internal/api -count=1` |
| 2026-10-18 | Weighted context/replay, `7768a82`; explanations/rules/feedback, `0ff61a3`; HTTP replay regression, `8d4ceab` | `go test ./internal/query ./internal/api -race -count=1` |
| 2026-10-18 | Internal account velocity/repository transactions, `ff6b560` | `go test ./internal/query -run 'TestVelocityWindow|TestSharedRepository' -race -count=1`; recorded weighted non-cmd81.65%, domain91.9% |
| 2026-10-18 | Durable SQL parity/restart, `3320761` | `go test ./internal/repository -run TestPostgresEventParity -race -count=1 -v`; actual PostgreSQL16.9, no skips |
| 2026-10-18 | HTTP failure boundaries, `4293c22` | `go test ./internal/api -race -count=1`; JSON errors/IDs/body limits/recovery; recorded weighted88.56% |
| 2026-10-18 | Runtime and DATABASE_URL selection, `26757f9` | `go test ./internal/service ./internal/repository -race -count=1`; real listener/shutdown and persisted runtime selection; recorded weighted89.18% |
| 2026-10-18 | Inferred structuring/scoped memory/retry ambiguity, `97c4c98` | `go test -p 2 ./... -race -count=1 -coverprofile=coverage.out`; domain92.1%, weighted504/558=90.32%; full log `ctl/logs/fraud/card5c-gates.log` |
| 2026-10-18 | Memory application latency, `97c4c98` | `go run ./cmd/latency -samples 10000 -raw`; unique p99=5459ns, replay=4209ns; method/environment in `docs/performance.md`, not HTTP/SQL/S1 |
| 2026-10-18 | Documentation/CI gate card, parent `97c4c98` | `go test ./scripts -count=1`; workflow parsed and weighted thresholds fixture-tested. Hosted workflow UNVERIFIED: not run. Exact card hash/evidence appended to STATUS to avoid self-referential commit hashes |

## Reproducible native gates

```sh
export GOTOOLCHAIN=local
export FRAUD_POSTGRES_CACHE="${TMPDIR:-/tmp}/fraud-postgres-cache"
test -z "$(gofmt -l .)"
go build ./...
go vet ./...
go test -p 2 ./... -race -count=1 -coverprofile=coverage.out
go tool cover -func=coverage.out
bash scripts/check-coverage.sh coverage.out
go mod tidy
git diff --exit-code -- go.mod go.sum
```

Structuring fuzz at `97c4c98`: `go test ./internal/domain -run '^$' -fuzz '^FuzzStructuringBand$' -fuzztime=2s -parallel=2`,145053 executions. Other parser fuzz target: FuzzParseMinorUnits. Local control logs/profiles are evidence locations, not dependencies of a clean clone.

## Open acceptance / limits

The active `openspec/changes/production-fraud` proposal remains **Partial**, not archived as fully production-complete:

- Validated RuleConfig injection and feedback learning exist, but no runtime policy-edit API/env. GET rules exposes PascalCase fields.
- SQL stores one locked full-history JSONB snapshot. Memory scoped operations copy selected account history under one mutex; retention/time indexing/scalable relational SQL are missing.
- Route/status/error families are tested, but success explanation/rules schemas and semantic schema validation are incomplete.
- Cold-clone under-two-minute timing, hosted CI, Linux/Windows, production migrations/backup/restore and real-process SQL HTTP crash drills are not verified.
- Separate-process10k S1 with injected faults belongs to manager integration/load; neither microtiming nor1000-call playback proves it.

No downstream HTTP/write exists in fraud: downstream outbox, jitter retry and circuit-breaker requirements are N/A here. No push/tag/PR/deploy/publication has occurred. Full Met/Partial/Unmet table: `openspec/dod.md`.

`openspec/gifs/playback-flow.gif` is a preserved historical storyboard, not a capture of the current service. The old77ms/55.9% values are historical only.
