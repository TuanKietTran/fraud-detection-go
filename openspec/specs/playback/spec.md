# Playback

## Purpose

The legacy playback fixture sends 1000 events across representative amounts. R1 playback sends 10,000 logical requests through the production HTTP router, with 2% duplicate submissions, transient HTTP faults and timeouts. Scoring is replay-safe and retains one decision per event identity.

## Requirements

### Requirement: Finish R1 requests inside one minute
The playback client SHALL complete 10,000 logical calls inside 60 seconds using bounded concurrency and at most six attempts per call. It SHALL retry transport failures, per-attempt timeouts, HTTP 408, 429 and 5xx with exponential backoff, full jitter, a 50ms base and 2s cap; it SHALL NOT retry business 4xx. Its production-router test SHALL inject at least 5% 503 (including some 429), 1% timeouts and 2% identical duplicate submissions, then assert terminal successes and one stored decision per unique event key.

### Requirement: Preserve legacy playback regression
The playback client SHALL continue to complete the 1000-call fixture in less than 60 seconds, including transient retries.

#### Scenario: Legacy burst with one 503 per call
- **GIVEN** the server returns HTTP 503 on the first attempt of each distinct body and then succeeds
- **WHEN** the client plays 1000 calls with at most 3 attempts
- **THEN** every call ends on the success status
- **AND** each call took 2 attempts
- **AND** the play finishes in under 60 seconds

### Requirement: Retry only transient failures
The client SHALL retry a transport error or HTTP 408, 429, or 5xx, and SHALL NOT retry any other status.

#### Scenario: HTTP 400 is final
- **WHEN** the server returns HTTP 400
- **THEN** the client stops after 1 attempt

#### Scenario: Repeated 503 stops at the attempt limit
- **WHEN** the server returns HTTP 503 on every attempt and the limit is 2
- **THEN** the client stops after 2 attempts
- **AND** the outcome status is 503

### Requirement: Playback does not repeat the effect
Playing the same call list again SHALL NOT apply a successful idempotent write a second time. A stateless read SHALL return the same body.

#### Scenario: Second play
- **GIVEN** the first play of 1000 calls succeeded
- **WHEN** the client plays that same list again
- **THEN** stored totals are unchanged
- **AND** each outcome is either the same success body or HTTP 400 containing `idempotency key already used`
