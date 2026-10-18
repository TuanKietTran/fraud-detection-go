## ADDED Requirements

### Requirement: Deterministic exact-money scoring
The service SHALL score submitted events deterministically using exact decimal amount and explicit currency, return a score from 0 through 100, a named decision, and named reasons. Exact USD parsing, weighted contextual scoring, internally accumulated account-window velocity and replay-safe memory/Postgres repositories are implemented. PostgreSQL runtime selection, inferred structuring, HTTP operations and measured latency remain pending.

#### Scenario: Replay identity
- GIVEN a scoring event with an event_id
- WHEN the same event is scored again without changing its payload
- THEN the same decision is returned and velocity is not counted twice.

#### Scenario: Published endpoint status
- GIVEN the OpenAPI contract
- WHEN consumers inspect its routes
- THEN POST /v1/score and explanation, rules, and feedback routes are marked implemented with documented compatibility aliases and statuses.
