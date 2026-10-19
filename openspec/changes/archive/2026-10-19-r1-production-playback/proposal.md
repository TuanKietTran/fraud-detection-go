# Production-router R1 playback

Prove the fraud service can replay 10,000 logical score requests inside a minute through its production HTTP router with transient failures, per-attempt timeouts and duplicate event identities. Keep the existing 1,000-call test as a separate regression. Expose resource samples and loopback-only pprof for later soak diagnosis. This change does not claim the manager-owned separate-process platform soak or scalable SQL.
