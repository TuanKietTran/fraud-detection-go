# Score event

## Purpose

Score synthetic USD events with exact integer cents, validated weighted rules and replay-safe process-lifetime explanations.

## Requirements

### Requirement: Exact weighted scoring
The service SHALL parse decimal strings (and transitional raw numeric tokens without floats), support USD with at most two fractional digits, and reject negative or overflowing amounts. Defaults add 40/HIGH_AMOUNT above 5000 and a further 30/VERY_HIGH_AMOUNT above 10000. Contextual rules add bounded weights for velocity, geo/device mismatch, new payee and supplied structuring within the configured amount band. Scores SHALL be capped at 100, with ALLOW below 40, REVIEW below 70 and DENY otherwise. POST /v1/score retains BLOCK as the documented legacy alias of DENY.

#### Scenario: Amount boundaries
- **GIVEN** default rules and no contextual signals or prior account events
- **WHEN** 5000.00, 5000.01, 10000.00 and 10000.01 are scored
- **THEN** scores are respectively 0, 40, 40 and 70.

### Requirement: Persist replay-safe scoring decisions
The service SHALL atomically record input, decision, timestamp, effective recent count and rule snapshot. Identical replay SHALL return the original decision; changed payload SHALL be rejected (legacy HTTP 400). The memory adapter is not durable across process restart.

#### Scenario: Replay after feedback or window expiry
- **GIVEN** an event already scored
- **WHEN** its exact input is replayed after rules or the clock change
- **THEN** the original decision and reasons are returned without adding another velocity event.

### Requirement: Account-window velocity
The service SHALL count prior unique events for the same account in (now-60s, now], using an injected clock. Legacy recent_count SHALL only increase the effective count.

#### Scenario: Concurrent unique events
- **GIVEN** multiple handlers sharing a repository and a velocity limit of one
- **WHEN** distinct events for the same account are scored concurrently at the same time
- **THEN** only one event is scored without velocity weight.

#### Scenario: Account isolation and time boundaries
- **GIVEN** events for another account, expired events at exactly now-60s, or future events after clock rollback
- **WHEN** a new event is scored
- **THEN** those events are excluded from velocity.

### Requirement: Explanation, configuration and feedback
GET /v1/events/{event_id}/explanation SHALL return the stored decision or 404. GET /v1/rules SHALL return active configuration. POST /v1/feedback SHALL accept CONFIRMED_FRAUD or FALSE_POSITIVE for an existing event, adjust ReviewAt by minus/plus one bounded by [1,DenyAt], replay the same label without another change, and reject conflicting labels.

#### Scenario: Feedback does not invalidate thresholds
- **GIVEN** ReviewAt equals DenyAt
- **WHEN** a false-positive label is applied
- **THEN** ReviewAt remains equal to DenyAt and configuration validates.

### Requirement: Inferred account-window structuring
The service SHALL infer STRUCTURING for the third and subsequent unique same-account attempt in the active window when the current and at least two prior amounts lie in [band-floor(band/10),band], excluding zero amounts. The qualifying prior count SHALL be stored. Legacy caller structuring=true SHALL remain additive; false SHALL NOT suppress inference. Replay SHALL retain the original result rather than being reclassified.

#### Scenario: Third near-band attempt
- **GIVEN** two unique near-band account attempts and arbitrary identical retries
- **WHEN** another unique near-band attempt is scored
- **THEN** STRUCTURING weight is added exactly once
- **AND** other accounts, expired/future timestamps and out-of-band amounts do not qualify.

### Requirement: Scoped memory atomicity
Application memory transactions SHALL select addressed event/feedback and requested account history, plus global policy. Out-of-scope writes or collisions with excluded events SHALL fail atomically. Full Update and SQL fallback SHALL preserve application behavior.

#### Scenario: Scope collision
- **GIVEN** an existing event belonging to an excluded account
- **WHEN** a callback attempts to overwrite its identity through another account scope
- **THEN** the complete update is rejected and original state remains.

### Requirement: Retry after ambiguous HTTP response
A canceled response after commit SHALL NOT duplicate effects when the same event_id and payload are retried.

#### Scenario: Commit before response cancellation
- **GIVEN** a fault wrapper that commits then waits for request cancellation
- **WHEN** the response fails and the same event is retried
- **THEN** its original result is returned and only one velocity/pattern contribution exists.

### Requirement: Durable transactional repository
The optional PostgreSQL adapter SHALL atomically retain replay identities, decisions, rule snapshots, velocity timestamps and feedback labels across a database restart. Callback errors SHALL roll back complete updates. DATABASE_URL selects this adapter at server startup; unset selects non-durable process memory. It loads/writes one full-history JSONB snapshot and makes no scalable-throughput claim.

#### Scenario: Database restart
- **GIVEN** a scored event and confirmed-fraud feedback committed in PostgreSQL
- **WHEN** the database is stopped and started using the same data directory
- **THEN** a newly opened adapter returns the original explanation and learned threshold
- **AND** replay of feedback does not apply it again.

### Requirement: Reject invalid JSON
Score and feedback SHALL reject malformed/trailing JSON with HTTP 400 and bodies exceeding 1048576 bytes with HTTP 413. Errors SHALL have JSON shape {error:{code,message},request_id}, matching X-Request-ID. IDs SHALL be server-generated rather than reflecting untrusted input.

#### Scenario: Not JSON
- **WHEN** a score body is not JSON
- **THEN** the response is HTTP 400 with the invalid_json code and a request ID.

### Requirement: Store failure and panic containment
A failed store SHALL NOT produce a successful rules/explanation/feedback response. Repository failure SHALL return 503, request cancellation or deadline 408, and recovered panic 500. Internal error details SHALL NOT be exposed. Legacy replay/feedback conflicts remain HTTP 400 as explicitly documented in OpenAPI.

#### Scenario: Failed explanation lookup
- **GIVEN** a repository returning a synthetic internal error
- **WHEN** an explanation is requested
- **THEN** response is 503 unavailable with no internal error detail and a request ID.

### Requirement: Validated bounded runtime
HTTP_ADDR defaults to :8103 and SHALL name a valid nonzero port. DATABASE_URL SHALL be a PostgreSQL URL when present. REQUEST_TIMEOUT, STARTUP_TIMEOUT and SHUTDOWN_TIMEOUT SHALL be positive durations at most one minute, defaulting to 5s/10s/10s. All routes SHALL have response deadlines; server read/header/write/idle operations SHALL be bounded. Signal cancellation SHALL initiate graceful shutdown with forced close on deadline. Logs SHALL be structured and omit database connection information.

#### Scenario: Native HTTP lifecycle
- **GIVEN** a server with an injected TCP listener
- **WHEN** readiness is requested then service context is canceled
- **THEN** readiness is 200 and serving terminates without synchronization sleeps.

### Requirement: Operational endpoints
GET /healthz SHALL report process liveness, GET /readyz SHALL probe the selected event store and return 503 when unavailable, and GET /metrics SHALL return process-local requests_total and failures_total (HTTP 5xx) counters.

#### Scenario: Store unavailable
- **GIVEN** a repository returning an error
- **WHEN** health and readiness are queried
- **THEN** health is 200 and readiness is 503.

### Requirement: Honest reproducible engineering evidence
Documentation SHALL distinguish measured memory application latency, httptest playback and real-process cross-service load. Every Met acceptance row SHALL name an asserting test or actual timing command. Missing runtime policy editing, scalable SQL and unexecuted workflow/deployment evidence SHALL remain explicitly Partial/Unmet.

#### Scenario: Weighted coverage gate
- **GIVEN** a profile containing domain, other non-cmd and uncovered cmd statements
- **WHEN** the coverage gate is executed
- **THEN** non-cmd coverage is weighted by statement counts with minimum80%
- **AND** domain coverage independently requires90%
- **AND** missing domain/empty profiles fail.

#### Scenario: Workflow versus local evidence
- **GIVEN** workflow parsing and gate tests pass locally
- **WHEN** documentation describes CI
- **THEN** hosted execution remains UNVERIFIED until it actually runs.

### Requirement: Explicit existing success response schemas
GET explanation200 SHALL declare the original stored event_id, bounded score, canonical ALLOW/REVIEW/DENY decision and named reasons. POST score SHALL retain documented BLOCK compatibility. GET rules200 SHALL declare every existing PascalCase RuleConfig field as a required integer with applicable bounds and exact int64 monetary serialization. Feedback202 SHALL declare an empty body. No runtime policy-edit surface is implied.

#### Scenario: Denial explanation
- **GIVEN** a default high-amount event scored as BLOCK on POST
- **WHEN** its explanation is fetched
- **THEN** the stored decision is DENY and both responses satisfy their schemas.

#### Scenario: Learned and extreme policy values
- **GIVEN** valid injected int64-max thresholds or a feedback-adjusted ReviewAt
- **WHEN** GET rules is called
- **THEN** PascalCase fields remain exact integers and satisfy the declared shape.

### Requirement: Parsed response-contract regression checks
Tests SHALL resolve published response references/schema definitions and exercise every declared v1 operation/status against real handlers. Checks SHALL include JSON media type, required/extra fields, string/array/exact integer types, enums, applicable bounds and error/header request-ID equality. Negative mutations SHALL prove rejection. This is a targeted response subset, not a full semantic OpenAPI/request validator.

#### Scenario: Response drift
- **GIVEN** a missing/wrong/extra field, invalid bounds/integer/enum or dangling reference
- **WHEN** response contract checks run
- **THEN** the defect is rejected.

### Requirement: Fail-closed domain validation
Policy validation SHALL reject ReviewAt below1, DenyAt before ReviewAt or above100, negative high amount, very-high amount before high, nonpositive velocity/band and any weight outside0..100. ScoreWithRules with invalid policy or negative amount/count SHALL return an error without a partial score/decision/reasons result. Valid equality/endpoints SHALL remain accepted. REVIEW SHALL include its lower threshold; DENY SHALL win at its threshold, including equal review/deny limits.

#### Scenario: Invalid inactive rule
- **GIVEN** a negative or above100 weight on any named rule
- **WHEN** a zero-signal event is scored
- **THEN** scoring fails closed even if that rule would not have triggered.

### Requirement: Exact money boundary evidence
Tests SHALL reject malformed whole/fractional text, signs, invalid scale/currency and overflow with the appropriate error sentinel and no partial value. Canonical integer-derived decimal text SHALL round-trip exactly through MaxInt64 cents. No float conversion or safety-check deletion is permitted to increase coverage.

#### Scenario: Maximum minor units
- **GIVEN** decimal92233720368547758.07 USD
- **WHEN** it is parsed
- **THEN** the value is MaxInt64 cents
- **AND** increasing its fractional cent to08 is rejected.

### Requirement: Bounded deterministic score evidence
Tests SHALL assert score bounds, deterministic repeated results, distinct retained reasons at score cap and nondecreasing score when adding a nonnegative geo signal. Statement coverage SHALL be reported separately from exhaustive input/branch claims.

#### Scenario: Cap retains explanation
- **GIVEN** valid configured signals triggering all seven rules
- **WHEN** the sum exceeds100
- **THEN** score is100 and all seven named reasons remain in deterministic order.
