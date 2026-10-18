# ADR 005: Strict startup seed, durable policy remains authoritative

Status: accepted for executable configuration; no online admin API.

`FRAUD_RULES_JSON` is optional JSON, bounded to8192 raw bytes including whitespace. Empty/whitespace input and `{}` retain defaults. A partial object overlays defaults using the exact PascalCase RuleConfig names from GET rules, with integer values (monetary thresholds are int64 cents). Unknown/case-aliased or duplicate decoded keys, null, strings/booleans/objects/arrays, fractional/exponent numbers, overflow, trailing values and invalid domain relationships fail configuration. Typed field decoding never converts money through float64. Errors do not echo raw input or decoder details.

`LoadConfig` validates before `Run`; `OpenStore` also validates the typed seed so programmatic callers cannot bypass it. Memory gets a new policy/history instance each startup. Feedback/replay remain stable within that instance; memory restart is explicitly non-durable and receives the new seed/defaults.

For PostgreSQL, the seed is used only when the singleton `fraud_state` row is first inserted. Existing state wins in full: incoming valid seed/defaults are not merged or reapplied, including across actual database restart. Policy, feedback identities, original decisions/rule snapshots and replay identities remain unchanged. Invalid incoming seeds still fail even when an existing valid policy would have ignored them. Startup validates the active stored policy/maps and fails rather than silently replacing corrupt state.

Initialization commits before HTTP listening; if a later bind/start fails, an already-created durable row remains authoritative on retry. This is database initialization, not a hot-reload mechanism or policy update. Startup logs advertise seed-only semantics without logging policy input. GET `/v1/rules` reports the actual active state; changing this environment value does not rewrite an existing database policy. No deletion/reset procedure or unauthenticated policy-write endpoint is introduced.

Evidence: strict parser/configuration rejection tables, real configured memory HTTP scoring and feedback/replay tests, and the existing real PostgreSQL fixture extended with an isolated schema. SQL tests prove a fresh seed changes HTTP scoring, a different/empty seed cannot reset learned thresholds, feedback/replay identities survive pool reopen and actual database stop/start, and invalid stored/incoming policy fails safely.

Consequences: operator control at first initialization closes the configurable-policy gap without risking learned-state resets. General online policy administration is not implemented or required by this decision. Singleton SQL serialization, unbounded selected account history, full OpenAPI/request semantics, hosted workflow and production recovery evidence remain separate limitations.
