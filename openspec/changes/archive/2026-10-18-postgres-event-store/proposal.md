# Durable transactional event state

Add an opt-in PostgreSQL event-state repository implementing the existing transaction port. A singleton JSONB row locked with SELECT FOR UPDATE commits event inputs, explanations, rule snapshots, velocity timestamps and feedback in one transaction. Database transactions provide cross-process serialization and rollback.

This deliberately simple adapter loads/writes the entire event history and serializes all operations. It is durable but not a scalable production event model; retention, indexed account-window queries and per-event rows are a future migration. No SQL throughput claim is made.

Use real embedded PostgreSQL to prove memory/SQL parity, concurrent handlers, replay conflicts, rollback, feedback identity, cancellation and data retained across an actual database restart. Startup/download failure is a failing test, never a skipped parity success. Server DATABASE_URL wiring belongs to the subsequent runtime card.
