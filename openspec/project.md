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
