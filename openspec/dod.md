# Acceptance status

**Met** requires implemented behavior and an asserting test (or explicit measurement where timing is the criterion). **Partial** means only a stated subset is implemented/verified. **Unmet** means not implemented or no qualifying evidence. A Met row does not imply overall production readiness.

Evidence baseline: `97c4c98`, native Go1.27.1 darwin/arm64, 2026-10-18; full race run (local gate log, not committed), profile `card5c.cover`, cover-function report `card5c-cover-func.log`. These control-directory artifacts are local lab evidence; commands in README regenerate them. Docker, CI execution and production deployment are not verified.

Response-contract follow-up, 2026-10-19 (parent `9132b35`): full resumed native gates pass in (local gate log, not committed) / `card7.cover`; API96.0%, weighted non-cmd505/558=90.50%, domain70/76=92.11%. Other package percentages unchanged; actual PostgreSQL suite86.635s, no skips. New response-contract tests named below are verified by this run, not inferred from the earlier baseline.

Domain-invariant follow-up, 2026-10-19 (parent `2a09ede`): one native gate pass in (local gate log, not committed) / `card8.cover`; domain76/76=100.00%, weighted non-cmd511/558=91.58%, no uncovered domain statement blocks. Other package percentages unchanged from card7; actual repository/PostgreSQL suite89.162s, no skips. All six current-source gaps were feasible and reached; no safety check was removed/excluded. This is statement coverage plus explicit invariant assertions, not exhaustive input/branch proof.

Startup-policy follow-up, 2026-10-19 (parent `7c0b261`): ONE native gate pass in (local gate log, not committed) / `card9.cover`; domain76/76=100.00%, weighted558/609=91.63%, service91.3%, other package percentages unchanged. Real repository/PostgreSQL85.354s without skips, including configured SQL initialization and actual restart retaining learned policy under changed/empty seeds.

## Met

| Acceptance | Exact test / evidence |
| --- | --- |
| Exact nonnegative USD cents, supported scale/currency and overflow rejection without floats | `domain.TestParseMinorUnits`, `FuzzParseMinorUnits`, `TestMoneyRejectsMalformedTextWithoutPartialValue`, `TestMoneyExactScaleAndInt64Boundaries`; typed error/no-partial-value assertions, MaxInt64 and 1000 deterministic integer-only canonical round trips |
| Strict amount boundaries and default decision bands | `domain.TestScoreAmountCentsBoundariesAndDecisions`, `domain.TestWeightedRulesAndValidation` (with no extra signals/history) |
| Validated executable startup policy in memory/SQL, durable learned policy never reset | `service.TestStartupRulesDefaultsOverlayAndExactIntegerBoundaries`, `TestStartupRulesRejectUnknownDuplicateNullNonIntegerOverflowAndInvalidPolicy`, `TestStartupPolicyDrivesMemoryHTTPScoringAndFreshRestart`; `repository.TestPostgresEventParity` startup-policy subtests exercise actual configured HTTP scoring, changed/empty-seed reopen/database restart, feedback/replay identity and corrupt stored-policy startup rejection. FRAUD_RULES_JSON seeds new stores only; no online admin write API is required/provided. |
| Every configured validation threshold/relationship/weight and fail-closed scoring | `domain.TestEveryRuleConfigThresholdAndRelationshipInvariant`, `TestEveryRuleWeightRejectsOutOfRangeAndAcceptsEndpoints`, `TestScoringRejectsNegativeSignalsWithoutPartialDecision`; all seven weights reject -1/101 and accept0/100 |
| Weighted rules, exact REVIEW/DENY boundaries, score cap and retained deterministic reasons | `domain.TestWeightedRulesAndValidation`, `TestWeightedDecisionAndRuleActivationBoundaries`, `TestScoreCapRetainsAllReasonsAndDeterministicSignalProperties`; equal decision limits prioritize DENY, all seven reasons survive cap, adding a nonnegative geo signal cannot reduce score |
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
| Published route/status sets are parsed and exercised | `api.TestOpenAPIParsedRoutesAndStatuses`, `service.TestOperationalRoutesMetricsAndUnavailableReadiness`, related error-path tables |
| Existing v1 success/error bodies match declared response shapes | `api.TestV1ResponsesMatchParsedSchemasForEveryDeclaredStatus` exercises every declared v1 status, all reason enums, canonical explanation DENY versus score BLOCK, empty feedback202, request-ID equality and learned policy |
| Exact PascalCase policy integer serialization and negative contract drift detection | `api.TestRuleConfigSchemaMatchesExtremeExactIntegerSerialization`, `TestResponseSchemaChecksRejectPayloadDrift`, `TestPublishedResponseReferencesAndSchemaDefinitionsResolve`; int64-max amount configuration is preserved without floats |
| Preserved playback regression:1000 score calls, one503 retry each, exact replay | `playback.TestThousandPerMinuteAndPlayback`; HTTP httptest, not separate service processes or S1 |
| Client400 not retried; transient503 retried with bounded attempts | `playback.TestDoesNotRetryClientError`, `TestRetriesTransientFailure`, `TestStopsAfterMaxAttempts`; playback helper only, not a downstream production client |
| Domain>=90%, weighted non-cmd>=80%, race clean | Latest domain76/76=100.00%, weighted558/609=91.63%; API96.0/query95.9/repository93.5/service91.3/playback61.5. `scripts.TestCoverageGateUsesWeightedStatementsAndDomainThreshold` asserts weighting/thresholds/excluded cmd behavior |
| Memory application-handler p99<5ms on measured isolated-account workload | `go run ./cmd/latency -samples 10000 -raw`: unique5459ns, replay4209ns; [environment and raw-sample scope](../docs/performance.md). Not HTTP/SQL latency or concurrent throughput |

## Partial

| Acceptance | Implemented subset / missing work |
| --- | --- |
| Scalable durable repository | SQL durability/parity proven, but singleton JSONB loads full history and serializes readers/writers. No relational per-event/account-time indexing or retention. Memory scoped account history is also unbounded. |
| Full semantic OpenAPI validation | Explanation/rules success schemas and targeted v1 response-shape checks are implemented. The helper supports objects/arrays/strings/exact integer text, enums, required/extra properties, min/max/minLength and local references. It is not a full request/schema validator (patterns, general formats, cross-field relations and arbitrary OpenAPI composition are not verified). |
| Under-two-minute clean-clone run | Minimal Go-only memory instructions supplied; cold download/clean-clone wall time is not measured. Local warm-cache compiled-process smoke checked12 HTTP operations and SIGTERM exit0 ((local gate log, not committed)); this is not a cold-clone claim. |
| All error/shutdown branches and exhaustive input/path proof | Domain has100% statement coverage and explicit validation/money/decision invariants asserted, but that is not a proof over all inputs/branches. Service remains91.3%; shutdown/error-path completeness is not claimed. |
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
