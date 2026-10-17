# Score event

## Purpose

Turn an event amount into a score, a decision, and reason codes. The HTTP response is the whole result; this route does not store it.

## Requirements

### Requirement: Allow a small amount
The system SHALL allow an event whose amount is 5000 or less, with score 0 and no reason codes.

#### Scenario: Amount at the allow boundary
- **WHEN** a client posts `event_id`, `account`, and amount 5000
- **THEN** the response is HTTP 200
- **AND** `Score` is 0
- **AND** `Decision` is `ALLOW`
- **AND** `ReasonCodes` is empty
- **AND** `EventID` is the posted event id

### Requirement: Review a high amount
The system SHALL add 40 and reason `HIGH_AMOUNT` when the amount is greater than 5000, and SHALL return `REVIEW` when the score is at least 40 and below 70.

#### Scenario: Amount just over 5000
- **WHEN** the amount is greater than 5000 and not greater than 10000
- **THEN** `Score` is 40
- **AND** `Decision` is `REVIEW`
- **AND** `ReasonCodes` is `HIGH_AMOUNT`

#### Scenario: Amount equal to 10000
- **WHEN** the amount is 10000
- **THEN** `Score` is 40
- **AND** `Decision` is `REVIEW`
- **AND** `ReasonCodes` contains `HIGH_AMOUNT` and does not contain `VERY_HIGH_AMOUNT`

### Requirement: Block a very high amount
The system SHALL add a further 30 and reason `VERY_HIGH_AMOUNT` when the amount is greater than 10000, and SHALL return `BLOCK` when the score is at least 70.

#### Scenario: Amount just over 10000
- **WHEN** the amount is greater than 10000
- **THEN** `Score` is 70
- **AND** `Decision` is `BLOCK`
- **AND** `ReasonCodes` contains `HIGH_AMOUNT` and `VERY_HIGH_AMOUNT`

### Requirement: Do not persist the score
The system SHALL NOT write the decision as part of `POST /v1/score`.

#### Scenario: Score then nothing to read
- **WHEN** a client scores an event
- **THEN** the only result is the HTTP body
- **AND** the API exposes no route that returns that decision later

### Requirement: Reject invalid JSON
The system SHALL reject a body that is not JSON.

#### Scenario: Not JSON
- **WHEN** the body is not JSON
- **THEN** the response is HTTP 400 and the body contains `invalid json`
