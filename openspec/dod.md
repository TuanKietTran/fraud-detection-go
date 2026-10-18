# Acceptance status

**Met** requires implemented behavior and an asserting test (or explicit measurement where timing is the criterion). **Partial** means only a stated subset is implemented/verified. **Unmet** means not implemented or no qualifying evidence. A Met row does not imply overall production readiness.

Evidence baseline: `97c4c98`, native Go1.27.1 darwin/arm64, 2026-10-18; full race run `ctl/logs/fraud/card5c-gates.log`, profile `card5c.cover`, cover-function report `card5c-cover-func.log`. These control-directory artifacts are local lab evidence; commands in README regenerate them. Docker, CI execution and production deployment are not verified.

## Met

| Acceptance | Exact test / evidence |
| --- | --- |
| Exact nonnegative USD cents, supported scale/currency and overflow rejection without floats | `domain.TestParseMinorUnits`, `domain.FuzzParseMinorUnits`; API uses raw decimal text |
| Strict amount boundaries and default decision bands | `domain.TestScoreAmountCentsBoundariesAndDecisions`, `domain.TestWeightedRulesAndValidation` (with no extra signals/history) |
| Weighted contextual rules, named reasons and score cap100 | `domain.TestWeightedRulesAndValidation` |
| Same input/event identity returns original result; changed input rejected | `query.TestReplayReturnsSameAndRejectsChangedPayload`, `api.TestOpenAPIParsedRoutesAndStatuses` |
| Replay-safe account/window velocity and account/time isolation | `query.TestVelocityWindowReplayAndAccountIsolation`, `query.TestSharedRepositoryConcurrentUniqueEventsAndReplay` |
| Inferred three-event near-band structuring, exact band/window/replay boundaries | `domain.TestStructuringExactBandAndCountBoundaries`, `query.TestInferredStructuringReplayWindowAndIsolation`; `FuzzStructuringBand` ran145053 executions in2s |
| Explanation, rules read and idempotent bounded feedback learning | `query.TestExplanationRulesAndFeedback`, `query.TestFeedbackNeverInvalidatesRuleThresholds`, `api.TestFeedbackJSONLegacyErrorsAndReplay` |
| Atomic memory rollback/cancellation and defensive copies | `repository.TestMemoryEventsAtomicRollbackCancellationAndCopies`, `query.TestDecisionCopiesAndInvalidInputs` |
| Indexed scoped memory views, rejection of escapes/collisions, preserved full Update | `repository.TestScopedTransactionsIsolationRollbackCopiesAndIndex`, `repository.TestScopedTransactionsRejectEscapesAndCancelAtomically` |
| Real PostgreSQL/memory parity, concurrent writes, rollback, feedback and retained state across actual DB restart | `repository.TestPostgresEventParity`, no skip; PostgreSQL16.9,91.593s recorded; also asserts DATABASE_URL selection reads learned state and inferred-pattern parity |
| Consistent API errors, authoritative request IDs, body limits and trailing JSON rejection | `api.TestJSONErrorsBodyLimitsRequestIDsAndRecovery`, `api.TestStoreFailuresAndCancellationNeverBecomeSuccessfulResponses` |
| Validated environment, bounded serving, health/readiness/metrics | `service.TestConfiguration`, `service.TestOperationalRoutesMetricsAndUnavailableReadiness`, `service.TestRequestTimeoutCancelsStoreAndReturnsJSON` |
| Operational fallback/panic JSON contains matching request ID | `service.TestOperationalPanicIsContained`, `service.TestOperationalFallbackErrorsAreJSON` |
| Real listener graceful cancellation and serve/listen/startup failures | `service.TestServeRealListenerGracefulCancellationAndServeFailure`, `service.TestRunStartupAndListenFailuresAreBoundedAndSanitized` |
| Safe retry after local write commits but HTTP response is canceled | `service.TestRetryAfterCommittedButCanceledHTTPResponse`, memory fault wrapper/channel synchronization; not a real-process SQL crash drill |
| Published route/status sets are parsed and exercised | `api.TestOpenAPIParsedRoutesAndStatuses`, `service.TestOperationalRoutesMetricsAndUnavailableReadiness`, related error-path tables; not a full OpenAPI semantic/schema validator |
| Preserved playback regression:1000 score calls, one503 retry each, exact replay | `playback.TestThousandPerMinuteAndPlayback`; HTTP httptest, not separate service processes or S1 |
| Client400 not retried; transient503 retried with bounded attempts | `playback.TestDoesNotRetryClientError`, `TestRetriesTransientFailure`, `TestStopsAfterMaxAttempts`; playback helper only, not a downstream production client |
| Domain>=90%, weighted non-cmd>=80%, race clean | Recorded domain92.1%, weighted504/558=90.32%; API95.0/query95.9/repository93.5/service90.9/playback61.5. `scripts.TestCoverageGateUsesWeightedStatementsAndDomainThreshold` asserts the new gate's weighting/thresholds/excluded cmd behavior |
| Memory application-handler p99<5ms on measured isolated-account workload | `go run ./cmd/latency -samples 10000 -raw`: unique5459ns, replay4209ns; [environment and raw-sample scope](../docs/performance.md). Not HTTP/SQL latency or concurrent throughput |

## Partial

| Acceptance | Implemented subset / missing work |
| --- | --- |
| Configurable validated policy | Domain accepts injected RuleConfig and deterministic feedback adjusts ReviewAt; executable seeds defaults. Runtime rule-edit API/policy env is not exposed. GET rules uses PascalCase exported field names. |
| Scalable durable repository | SQL durability/parity proven, but singleton JSONB loads full history and serializes readers/writers. No relational per-event/account-time indexing or retention. Memory scoped account history is also unbounded. |
| HTTP contract completeness | Actual route/status/error families covered, but success explanation/rules schemas are not fully specified/validated and status tests are not a comprehensive OpenAPI validator. |
| Under-two-minute clean-clone run | Minimal Go-only memory instructions supplied; cold download/clean-clone wall time is not measured. Local warm-cache compiled-process smoke checked12 HTTP operations and SIGTERM exit0 (`ctl/logs/fraud/card6-gates.log`); this is not a cold-clone claim. |
| All error/shutdown branches and domain invariants exhaustively covered | Measured domain92.1%, service90.9%, not100%; no claim of exhaustive branch/property coverage. |
| Hiring-manager documentation / CI | README, diagram, curl, glossary, ADRs and measured report supplied; workflow and weighted gate specified/tested locally, but hosted workflow remains UNVERIFIED: not run. |

## Unmet / not claimed

| Acceptance | Reason |
| --- | --- |
| Real-process10k payment S1 with faults under60s | Manager-owned integration/load effort; application microtiming and1000 httptest playback are not substitutes. |
| Production migrations, least-privilege rollout, backup/restore and HTTP/SQL crash recovery | Constructor DDL and synthetic DB restart test only; legacy migration is not the current adapter schema. |
| Authentication/authorization, real PII/payment data or production fraud accuracy | Deliberately synthetic unauthenticated local demonstration; not safe to expose publicly. |
| Hosted CI execution, Linux/Windows runtime verification, Docker build/deploy | Not run here; no Docker dependency/artifact claim. |
| Push/tag/PR/publication/deployment | Owner-only and not performed. |

## Not applicable

Fraud performs no downstream HTTP calls or downstream writes, so downstream outbox, retry-with-jitter and circuit-breaker requirements do not apply to this service. Playback backoff tests do not change that classification.

## Historical storyboard

`openspec/gifs/playback-flow.gif` is the preserved baseline storyboard, not a current screen capture. Old claims that scores are stateless/unpersisted, float-based or only amount-scored are superseded by the current spec. The old77ms/55.9% baseline is historical, not a current measurement.
