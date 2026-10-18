# Validated executable startup policy

Close the existing configurable-policy acceptance gap without an admin write API. FRAUD_RULES_JSON is an optional at-most8192-byte JSON object with exact PascalCase RuleConfig field names and integer values. Partial objects overlay current defaults; empty/whitespace input and {} preserve defaults. Reject unknown/case-aliased/duplicate fields, null/non-integer/overflow values, trailing JSON and invalid policy relationships before opening a store/listener. Errors must not reflect raw policy input.

The parsed policy initializes each new memory store and the first PostgreSQL state row. Existing PostgreSQL policy, feedback, events and replay identities take precedence: a valid new startup seed is ignored for existing durable state, never merged or reapplied. Invalid incoming seeds fail even against existing durable state; invalid stored state fails initialization rather than being reset. Feedback still changes the active policy within domain bounds and original event decisions stay immutable.

Prove actual configured HTTP scoring in memory and PostgreSQL modes, empty/default behavior, fresh-memory restart seeding, durable reopen/database restart retaining learned policy under a different/empty startup seed, and strict/sanitized rejection. Keep SQL scalability, full semantic OpenAPI and hosted CI limitations separate and honest.
