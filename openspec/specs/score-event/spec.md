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

### Requirement: Reject invalid JSON
The service SHALL reject malformed or trailing JSON with HTTP 400.

#### Scenario: Not JSON
- **WHEN** a score body is not JSON
- **THEN** the response is HTTP 400.
