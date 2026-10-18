# Infer structuring and scope memory transactions

Infer a named synthetic structuring pattern from at least three unique same-account events in the existing (now-60s, now] window, each amount within the inclusive top 10% of StructuringBandCents. The third qualifying attempt receives STRUCTURING even if the caller omitted/cleared the legacy flag. Preserve that flag as an additional signal and preserve original replay results. Store the qualifying prior count with the decision/rule snapshot. No real-world detection accuracy is claimed.

Avoid cloning unrelated account history in the memory application's atomic transactions. Add an optional scoped repository capability with account-indexed selection, rollback/cancellation, defensive copies and rejection of out-of-scope writes/collisions. Keep the full Update port and PostgreSQL fallback semantics intact. Per-account memory history and the singleton PostgreSQL snapshot remain unbounded scalability limitations.

Measure unique isolated-account and replay application-handler cohorts with a reproducible command retaining raw samples. Do not call this HTTP latency, SQL performance or real-process S1/10k evidence. Prove safe retry after an HTTP request is canceled after its underlying transaction committed, synchronizing by channels rather than sleeps.
