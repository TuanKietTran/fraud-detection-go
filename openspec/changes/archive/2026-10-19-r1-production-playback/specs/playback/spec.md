## MODIFIED Requirements

### Requirement: Finish R1 requests inside one minute
The playback client SHALL complete 10,000 logical calls inside 60 seconds using bounded concurrency and at most six attempts per call. It SHALL retry transport failures, per-attempt timeouts, HTTP 408, 429 and 5xx with exponential backoff, full jitter, a 50ms base and 2s cap; it SHALL NOT retry business 4xx. Its production-router test SHALL inject at least 5% transient HTTP failures including 503 and some 429, 1% timeouts and 2% identical duplicate submissions, then assert terminal success and one stored decision per unique event key.

#### Scenario: Production router fault-injected replay
- **GIVEN** 10,000 logical scoring calls including 2% identical duplicates
- **WHEN** the production HTTP router is exercised with transient failures and per-attempt timeouts
- **THEN** all calls terminate successfully inside 60 seconds
- **AND** there is exactly one stored decision for each unique event key.

## ADDED Requirements

### Requirement: Preserve legacy playback regression
The playback client SHALL continue to complete the 1000-call fixture in less than 60 seconds, including transient retries.

#### Scenario: Legacy burst with one 503 per call
- **GIVEN** the server returns HTTP 503 on the first attempt of each distinct body and then succeeds
- **WHEN** the client plays 1000 calls with at most 3 attempts
- **THEN** every call ends on the success status
- **AND** each call took 2 attempts
- **AND** the play finishes in under 60 seconds.
