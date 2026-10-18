# ADR 001: Atomic event, policy and feedback state

Status: accepted for this synthetic demonstration.

A decision depends on account history and a mutable learned policy. Independent writes would allow concurrent decisions to miss velocity, or feedback retries to adjust policy twice. The repository therefore serializes one transaction spanning the replay lookup, account history, rule snapshot and recorded result. Replays preserve the original result even after policy/time changes.

The memory implementation uses isolated snapshots and commits only successful callbacks. The PostgreSQL implementation locks one JSONB state row and commits the entire state atomically. Its schema is initialized idempotently by the adapter constructor; the legacy `migrations/001_init.sql` does not describe this adapter.

Consequences: memory loses everything on process restart; PostgreSQL retains state and serializes across handlers/processes. Both currently copy/load full history and PostgreSQL serializes readers as well as writers. This is a correctness-first educational adapter, not a high-throughput or unbounded-history production design. Large histories require per-event storage, an indexed account/time query, bounded retention and a policy version row. Do not infer SQL latency from an in-memory measurement.

Only synthetic fixtures are supported. SQL connection information must not be logged. Production migrations, backup/restore and privileges have not been exercised.
