## ADDED Requirements

### Requirement: Strict startup policy initialization
FRAUD_RULES_JSON SHALL be at most8192 raw bytes and a single object of canonical PascalCase integer RuleConfig overrides. Empty/whitespace and {} SHALL retain defaults. Unknown/case-aliased/duplicate keys, null/noninteger/overflow values, trailing JSON and invalid policy SHALL fail startup with sanitized errors. Typed seeds SHALL also be validated before store/listener initialization.

#### Scenario: Configured actual scoring
- **GIVEN** a valid seed with high amount100cents, weight20 and ReviewAt20
- **WHEN** a fresh account attempts USD2.00 through HTTP in memory or SQL mode
- **THEN** score is20/REVIEW rather than default0/ALLOW.

### Requirement: Existing durable policy takes precedence
The seed SHALL initialize only new memory stores or the first durable state row. Existing PostgreSQL policy/feedback/replay records SHALL never be merged/reset by changed or empty seeds, including across database restart. Invalid incoming or stored policy SHALL fail startup rather than reset state. Initialization commits before listening; later listen failure SHALL NOT undo an inserted durable row. Feedback validation and original event results SHALL remain preserved.

#### Scenario: Learned durable restart
- **GIVEN** learned ReviewAt21 and an original score20/REVIEW under seedA
- **WHEN** the database restarts and startup supplies different seedB or defaults
- **THEN** learned policy and feedback identity remain unchanged
- **AND** replay remains20/REVIEW while a new equivalent attempt scores20/ALLOW.
