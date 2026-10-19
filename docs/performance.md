# Measured memory application latency

Measured 2026-10-18 on Apple M1, 17,179,869,184 bytes RAM, Go 1.27.1 darwin/arm64, GOMAXPROCS=8. Command:

```sh
GOTOOLCHAIN=local go run ./cmd/latency -samples 10000 -raw
```

| Cohort | Calls | p50 (ns) | p95 (ns) | p99 (ns) |
| --- | ---: | ---: | ---: | ---: |
| Unique event, isolated account | 10,000 | 1,291 | 3,292 | 5,459 |
| Exact replay of those events | 10,000 | 1,334 | 2,750 | 4,209 |

Both p99 values are below the 5ms in-process target (unique p99 = 0.005459ms). These are nearest-rank percentiles of each chronological cohort, without discarded samples. Timing covers `ScoreEventHandler.Handle`, including the scoped memory transaction, but excludes caller payload construction. Every call is checked for successful ALLOW/score0. Unique calls grow retained history from zero through 10,000 distinct accounts; replay runs against that retained state. The command emits environment metadata and optional raw samples and exits nonzero if either p99 is at least 5ms.

Raw evidence for this run: (local gate log, not committed) (20,000 chronological samples, 99,954 bytes) in the lab control directory. The implementation/report commit is recorded in `STATUS-fraud.md`; its parent is `26757f9`. Re-running the command regenerates the raw evidence without a checked-in machine-specific artifact.

**Scope:** this is not HTTP latency, a PostgreSQL benchmark, concurrent throughput or the real-process S1 10k-request load target. It does not establish long-running same-account performance. Scoped memory transactions avoid cloning unrelated accounts but still retain/copy selected account history and serialize under one mutex. The full memory Update path and singleton PostgreSQL adapter still copy/load full history. Retention, per-account indexing by time and a scalable relational SQL model remain limitations.

Separate correctness evidence: race tests cover concurrent replay/unique events, scope escape/collision rejection, rollback/cancellation, defensive copies, memory/SQL inferred-structuring parity and retry after a committed-but-canceled HTTP response. Those tests are not substituted for this timing measurement.
