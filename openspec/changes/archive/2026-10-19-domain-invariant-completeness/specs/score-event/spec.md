## ADDED Requirements

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
Tests SHALL assert score bounds, deterministic repeated results, distinct retained reasons at score cap and nondecreasing score when adding a nonnegative geo signal. Statement coverage SHALL be reported separately from claims of exhaustive input/branch correctness.

#### Scenario: Cap retains explanation
- **GIVEN** valid configured signals triggering all seven rules
- **WHEN** the sum exceeds100
- **THEN** score is100 and all seven named reasons remain in deterministic order.
