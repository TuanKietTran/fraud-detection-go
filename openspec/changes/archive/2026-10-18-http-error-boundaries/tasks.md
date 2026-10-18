# Tasks

- [x] Return JSON errors and authoritative server request IDs on every API route.
- [x] Bound bodies using MaxBytesReader and reject trailing JSON on score and feedback.
- [x] Distinguish typed validation/conflict/missing-event errors from store failure and cancellation.
- [x] Recover synthetic panics without leaking payloads or sensitive error details.
- [x] Expand OpenAPI error schemas/status checks and deterministic handler tests for every error family.
- [x] Run full native gates and record measured coverage before archiving: `ctl/logs/fraud/card5a-gates.log`; API 95.0%, domain 91.9%, query 95.6%, repository 88.9%, playback 61.5%; weighted non-cmd 325/367 = 88.56%. Full build/vet/race/tidy/format gates passed, including real PostgreSQL suite (94.250s). Final OpenAPI description revalidated with `go test -p 2 ./internal/api -race -count=1`.
