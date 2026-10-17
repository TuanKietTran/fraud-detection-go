## ADDED Requirements

### Requirement: Deterministic exact-money scoring
The service SHALL score submitted events deterministically using exact decimal amount and explicit currency, return a score from 0 through 100, a named decision, and named reasons. The implemented POST /v1/score currently remains a legacy amount-only endpoint; contextual rules and exact-money validation are planned and MUST NOT be represented as implemented until shipped.

#### Scenario: Replay identity
- GIVEN a scoring event with an event_id
- WHEN the same event is scored again without changing its payload
- THEN the same decision is returned and velocity is not counted twice.

#### Scenario: Published endpoint status
- GIVEN the OpenAPI contract
- WHEN consumers inspect its routes
- THEN POST /v1/score is marked implemented and explanation, rules, and feedback routes are marked planned until implemented.
