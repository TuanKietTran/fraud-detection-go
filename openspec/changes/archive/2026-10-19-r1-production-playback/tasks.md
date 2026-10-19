# Tasks

- [x] Preserve the 1,000-call playback regression.
- [x] Implement bounded concurrent retries with six attempts, per-attempt deadlines, full jitter, exponential 50ms base and 2s cap; do not retry business 4xx.
- [x] Exercise 10,000 logical submissions through production routes on a real loopback HTTP listener with 2% identical duplicates, >=5% transient 503/429 responses and 1% timeout attempts.
- [x] Assert every logical request terminates successfully and the store contains one decision per unique event identity.
- [x] Add runtime goroutine/heap/TCP-connection/GC-pause metrics and loopback-only pprof endpoints; test exposed metric fields and access guard.
- [x] Update current playback spec, API contract, README and DoD. Verified 2026-10-19 on Go1.27.2 windows/amd64: R1 test passes under race in 8.64s; full build/vet/race/tidy gates pass. Manager-owned R2 soak, hosted workflow and Docker remain unverified.
