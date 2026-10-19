# Architecture decision records

Common platform decisions and fraud-specific decisions are indexed here. These decisions describe a synthetic scoring demonstration, not production fraud accuracy or production deployment.

- [Platform 01: CQRS-inspired boundary](platform-01-cqrs.md) (also [ADR 004](004-cqrs-inspired-application-boundary.md))
- [Platform 02: exact money](platform-02-exact-money.md) (also [ADR 002](002-exact-money-and-replay.md))
- [Platform 03: idempotency](platform-03-idempotency.md) (also [ADR 002](002-exact-money-and-replay.md))
- [Platform 04: outbox and resilience](platform-04-outbox-and-resilience.md)
- [Platform 05: persistence and testing](platform-05-persistence-and-testing.md) (also [ADR 001](001-transactional-event-state.md))
- [ADR 003: runtime deadlines and readiness](003-runtime-deadlines-and-readiness.md)
- [ADR 005: startup policy seeding](005-startup-policy-seeding.md)

## Fraud explainability trade-off

Scoring returns named rule reasons with the immutable decision. A model score without stable explanations would make replay behavior harder to audit; deterministic weighted rules keep the demonstration inspectable. This is explainability of the implemented synthetic rules, not evidence of predictive accuracy.
