# Tasks

- [x] Replace README with run/curl, Mermaid layers/consumers, measured facts and explicit operational/compatibility limits.
- [x] Add domain glossary and reconcile DoD with current named tests and Partial/Unmet rows.
- [x] Add CI format/build/vet/race/coverage/module-hygiene workflow, explicitly UNVERIFIED: not run.
- [x] Add locally testable weighted statement coverage gate (domain>=90%, aggregate non-cmd>=80%) and workflow parsing assertions.
- [x] Local compiled-memory process smoke:12 HTTP checks and SIGTERM graceful exit0; warm-cache elapsed0.492086s excluding build, not cold-clone/S1 timing. Results in ctl/logs/fraud/card6-gates.log.
- [x] Native gates passed: Go1.27.1 darwin/arm64, build/vet/race/format/tidy/diff, real PostgreSQL89.015s without skips, domain70/76=92.11%, weighted non-cmd504/558=90.32%. New scripts tests (weighted fixture gates and parsed workflow commands) passed; hosted workflow remains unverified.
- [x] Archive this focused documentation change; retain broader production-fraud proposal as explicitly Partial for unresolved acceptance.
