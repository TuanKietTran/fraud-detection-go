# Platform ADR 05: persistence and testing

Status: accepted for the current demonstration.

The memory repository supports fast deterministic tests; optional PostgreSQL proves transactional persistence and restart parity. Its current adapter stores one full-history JSONB snapshot and serializes readers/writers, so it is not a scalable relational design. Tests exercise memory invariants, API handlers, real loopback HTTP replay, and embedded PostgreSQL without Docker. See [ADR 001](001-transactional-event-state.md) and [ADR 003](003-runtime-deadlines-and-readiness.md).
