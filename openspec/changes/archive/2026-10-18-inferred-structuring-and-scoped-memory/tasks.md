# Tasks

- [x] Infer three-event top-10%-band structuring with exact integer boundaries and no new persisted configuration requirement.
- [x] Test replay, account/time/future-timestamp isolation, snapshots, property/fuzz band invariants and memory/full-store parity.
- [x] Add scoped indexed memory transactions preserving full Update, defensive copies, rollback, cancellation and scope rejection.
- [x] Extend real PostgreSQL parity with inferred structuring.
- [x] Add deterministic ambiguous-after-commit HTTP retry regression.
- [x] Add cmd/latency for raw unique/replay application-handler percentiles.
- [x] Native build/vet/race/tidy/format/diff gates passed; `ctl/logs/fraud/card5c-gates.log`. API95.0%, domain92.1%, query95.9%, repository93.5%, service90.9%, playback61.5%; weighted non-cmd504/558=90.32%. Real PostgreSQL91.593s, no skips. Structuring fuzz2s:145053 executions. 10000 samples/cohort: unique p50/p95/p99=1291/3292/5459ns, replay1334/2750/4209ns; raw report in ctl/logs/fraud/card5c-latency.json and scope in docs/performance.md.
