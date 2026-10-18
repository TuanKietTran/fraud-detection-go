## MODIFIED Requirements

### Requirement: Reject invalid JSON
Score and feedback SHALL reject malformed/trailing JSON with HTTP 400 and bodies exceeding 1048576 bytes with HTTP 413. Errors SHALL have JSON shape {error:{code,message},request_id}, matching X-Request-ID. IDs SHALL be server-generated rather than reflecting untrusted input.

#### Scenario: Trailing feedback JSON
- **GIVEN** a valid feedback object followed by a second JSON object
- **WHEN** it is submitted
- **THEN** the response is 400 invalid_json and no feedback is committed.

## ADDED Requirements

### Requirement: Store failure and panic containment
A failed store SHALL NOT produce a successful rules/explanation/feedback response. Repository failure SHALL return 503, request cancellation or deadline 408, and recovered panic 500. Internal error details SHALL NOT be exposed.

#### Scenario: Failed explanation lookup
- **GIVEN** a repository returning a synthetic internal error
- **WHEN** an explanation is requested
- **THEN** response is 503 unavailable with no internal error detail and a request ID.
