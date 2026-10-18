## ADDED Requirements

### Requirement: Explicit existing success response schemas
GET explanation200 SHALL declare the original stored event_id, bounded score, canonical ALLOW/REVIEW/DENY decision and named reasons. POST score SHALL retain its documented BLOCK compatibility. GET rules200 SHALL declare every existing PascalCase RuleConfig field as a required integer with applicable bounds and exact int64 monetary serialization. Feedback202 SHALL declare an empty body. No new route or runtime policy-edit behavior is implied.

#### Scenario: Denial explanation differs only by canonical decision alias
- **GIVEN** a default high-amount event scored as BLOCK on POST
- **WHEN** its explanation is fetched
- **THEN** its stored decision is DENY and both responses satisfy their published schemas.

#### Scenario: Learned and extreme policy values
- **GIVEN** valid injected int64-max thresholds or a feedback-adjusted ReviewAt
- **WHEN** GET rules is called
- **THEN** its PascalCase fields remain exact integers and satisfy the published policy shape.

### Requirement: Parsed response-contract regression checks
Tests SHALL resolve published response references/schema definitions and exercise every declared v1 operation/status against real handlers. Checks SHALL include JSON media type, required/extra fields, string/array/integer types, enums, applicable bounds and error/header request-ID equality. Negative mutations SHALL demonstrate rejection. These checks SHALL be described as a targeted subset, not a full semantic OpenAPI/request validator.

#### Scenario: Response drift
- **GIVEN** a missing field, unexpected field, wrong casing, invalid integer/bounds/enum or dangling reference
- **WHEN** the response contract checks run
- **THEN** the defect is rejected.
