# Fraud detection project

Exact-money synthetic risk scoring with pure weighted rules, replay-safe account-window history, inferred structuring, immutable explanations and deterministic feedback learning. The process selects indexed scoped memory or a durable PostgreSQL snapshot adapter. Runtime has validated environment, bounded serving/shutdown, JSON logs and operational endpoints.

## Stack and public surface

Module `github.com/TuanKietTran/fraud-detection-go`, minimum Go1.25; native evidence uses Go1.27.1 darwin/arm64. Chi/UUID, pgx for optional SQL, embedded PostgreSQL for real parity tests, yaml.v3 for parsed contracts/workflow tests. Money is integer USD cents parsed directly from decimal text, never through float64.

- POST `/v1/score`
- GET `/v1/events/{event_id}/explanation`
- GET `/v1/rules`
- POST `/v1/feedback`
- GET `/healthz`, `/readyz`, `/metrics`; loopback-only `/debug/pprof/*` for runtime profiles

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
| 2026-10-18 | Inferred structuring/scoped memory/retry ambiguity, `97c4c98` | `go test -p 2 ./... -race -count=1 -coverprofile=coverage.out`; domain92.1%, weighted504/558=90.32% |
| 2026-10-18 | Memory application latency, `97c4c98` | `go run ./cmd/latency -samples 10000 -raw`; unique p99=5459ns, replay=4209ns; method/environment in `docs/performance.md`, not HTTP/SQL/S1 |
| 2026-10-18 | Documentation/CI gate card, `9132b35` | `go test ./scripts -count=1`; workflow parsed and weighted thresholds fixture-tested. Hosted workflow UNVERIFIED: not run. |
| 2026-10-19 | Portable Windows coverage gate, `c29854d` (pushed) | `go test -p 2 ./scripts -count=1`; real Go profile with zero-statement entries supported; full native gates passed. |
| 2026-10-19 | R1 production-router replay and runtime sampling, `b68c36e` (pushed) | `go test -p 1 ./internal/playback -run TestR1ProductionRouterTenThousand -count=1 -race -v`; 10,000 logical calls / 9,800 unique decisions in 8.64s; full module race gates passed. |

## Response contract checkpoint

Card7 `2a09ede` (parent `9132b35`): explicit explanation and PascalCase RuleConfig success schemas, empty feedback202, canonical explanation DENY versus score BLOCK, all declared v1 statuses, all reason enums, learned/extreme exact integer policies, negative response mutations and response/schema reference resolution. Verification: `go test -p 2 ./internal/api -race -count=1` plus full native gates. Native resumed gates completed 2026-10-19: API96.0%, domain92.1%, query95.9%, repository93.5%, service90.9%, playback61.5%; weighted non-cmd505/558=90.50%, domain70/76=92.11%. Real PostgreSQL86.635s without skips. This is targeted response checking, not full OpenAPI/request semantic validation.

## Domain invariant checkpoint

Card8 `7c0b261` (parent `2a09ede`), verified 2026-10-19: current source/profile confirmed the six uncovered blocks in ParseMinorUnits and ScoreWithRules/RuleConfig.Validate. Meaningful malformed-money, invalid-policy and REVIEW assertions now reach all six. Tests also cover every configured limit/weight/relationship, exact int64/scale/currency boundaries, 1000 deterministic canonical money round trips, capped reason retention and deterministic/nondecreasing geo-signal properties. No production source, safety check, API or dependency changed.

A native format/build/vet/full-race/tidy/coverage gate passed for the domain-invariant milestone: domain76/76=100.00%, weighted non-cmd511/558=91.58%. API96.0%, query95.9%, repository93.5%, service90.9%, playback61.5%; actual SQL/repository suite89.162s, no skips. No remaining uncovered domain statement block or unreachable-exclusion claim; statement100% is not exhaustive input/path proof.

## Startup policy checkpoint

Parent `7c0b261`: bounded strict FRAUD_RULES_JSON overlays defaults only for initialization. Empty preserves defaults; unknown/duplicate/case-aliased/null/noninteger/overflow/invalid policy fails startup safely. Memory starts fresh; existing PostgreSQL state/feedback/replay wins unchanged across different/empty-seed reopen and actual DB restart. Active corrupt state fails rather than resets. One native gate pass: domain76/76=100%, weighted558/609=91.63%, service91.3%; actual SQL/repository85.354s, no skips. ADR005 defines initialization-before-listen semantics; broader scalability/full-schema/hostedCI acceptance remains Partial.

## Reproducible native gates

```sh
export GOTOOLCHAIN=local
export FRAUD_POSTGRES_CACHE="${TMPDIR:-/tmp}/fraud-postgres-cache"
test -z "$(gofmt -l .)"
go build ./...
go vet ./...
go test -p 2 ./... -race -count=1 -coverprofile=coverage.out
go tool cover -func=coverage.out
go run ./scripts/coveragegate coverage.out
go mod tidy
git diff --exit-code -- go.mod go.sum
```

Structuring fuzz at `97c4c98`: `go test ./internal/domain -run '^$' -fuzz '^FuzzStructuringBand$' -fuzztime=2s -parallel=2`,145053 executions. Other parser fuzz target: FuzzParseMinorUnits. Current Windows/amd64 weighted statement coverage is 608/705=86.24%; profile generation and the portable gate are reproducible from a clean checkout. 

## Open acceptance / limits

The active `openspec/changes/production-fraud` proposal remains **Partial**, not archived as fully production-complete:

- Configurable-policy acceptance is now met via strict FRAUD_RULES_JSON startup seeding and feedback; existing SQL learned state is never reset. No online admin write API is required/provided. GET rules retains PascalCase fields.
- SQL stores one locked full-history JSONB snapshot. Memory scoped operations copy selected account history under one mutex; retention/time indexing/scalable relational SQL are missing.
- Existing v1 response success/error schemas are declared and checked against handlers, including explanation/rules. The targeted helper is not a full semantic OpenAPI/request validator; arbitrary composition, patterns/formats and cross-field schema relations remain outside its scope.
- Cold-clone under-two-minute timing, hosted CI, Linux/Windows, production migrations/backup/restore and real-process SQL HTTP crash drills are not verified.
- Separate-process10k S1 with injected faults belongs to manager integration/load; neither microtiming nor1000-call playback proves it.

No downstream HTTP/write exists in fraud: downstream outbox and downstream circuit-breaker requirements are N/A here. R1 now exercises production routing, while the separate-process 60-window platform soak remains manager-owned. Commits through `b68c36e` are published on this repository's `main`; no release or deployment is claimed. Full Met/Partial/Unmet table: `openspec/dod.md`.

`openspec/gifs/playback-flow.gif` is a preserved historical storyboard, not a capture of the current service. The old77ms/55.9% values are historical only.
