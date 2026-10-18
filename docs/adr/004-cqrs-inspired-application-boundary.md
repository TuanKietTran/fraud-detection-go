# ADR 004: Keep the scoring application transaction explicit

Status: accepted for the current scope; not a claim of full CQRS deployment.

The baseline named its use case a query. Once a decision depends on account-window history, scoring also records an immutable result and must not be a stateless read. Replay lookup, history classification, pure rule evaluation and event insertion therefore execute in one application transaction. Feedback commits its identity and learned policy together.

The repository keeps the existing `internal/query` name for continuity; domain rules remain pure and HTTP/wiring remain outside the domain. Explanation and rules reads use the same repository boundary. There are no independent read-model processes, asynchronously projected databases or separately deployed command/query services. README calls this CQRS-inspired layering rather than implying those systems exist.

Why not separate projections here: stale velocity or delayed replay identity could change a retried decision or count a duplicate. The demonstration prioritizes observable atomic invariants and adapters/tests over deployment complexity. Consequences are explicit serialization and limited throughput/scaling, especially for the PostgreSQL snapshot model. The legacy `InMemoryFraud` adapter and zero UUID field are not runtime identities; event_id is authoritative.

Online policy administration is not provided or required for startup configuration. Validated FRAUD_RULES_JSON seeds new memory/SQL stores while preserving existing durable learned state (ADR005); deterministic feedback remains transactional. Rule/explanation success schemas and targeted response checks are implemented; broader semantic OpenAPI validation remains partial. Nothing in this decision justifies describing the service as fully production-ready.
