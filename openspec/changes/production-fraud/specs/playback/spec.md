## MODIFIED Requirements

### Requirement: Playback throughput and retry behavior
The playback client SHALL complete 10,000 logical calls inside 60 seconds using bounded concurrency and at most six attempts per call. It SHALL retry transport failures, per-attempt timeouts, HTTP 408, 429 and 5xx with exponential backoff, full jitter, a 50ms base and 2s cap; it SHALL NOT retry business 4xx. Its production-router test SHALL inject at least 5% 503 (including some 429), 1% timeouts and 2% identical duplicate submissions, then assert terminal successes and one stored decision per unique event key. The existing 1000-call fixture SHALL remain as a separate regression.

#### Scenario: Production router fault-injected replay
- **GIVEN** 10,000 logical scoring calls including 2% identical duplicates
- **WHEN** the production HTTP router is exercised with 503, 429 and per-attempt timeouts
- **THEN** all calls terminate successfully inside 60 seconds
- **AND** there is exactly one stored decision for each unique event key.
