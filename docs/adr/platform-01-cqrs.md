# Platform ADR 01: CQRS-inspired application boundary

Status: accepted for the current scope.

Scoring reads account history and commits a unique immutable decision, so replay lookup, velocity classification, pure rule evaluation and storage must remain one transaction. `internal/query` retains its existing package name; the service does not deploy separate command/query services or projections. See [ADR 004](004-cqrs-inspired-application-boundary.md).
