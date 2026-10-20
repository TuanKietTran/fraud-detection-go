# Acceptance status

**Met** requires implemented behavior and an asserting test (or explicit measurement where timing is the criterion). **Partial** means only a stated subset is implemented/verified. **Unmet** means not implemented or no qualifying evidence. A Met row does not imply overall production readiness.

Current full native gate evidence: 2026-10-19, Go1.27.2 windows/amd64; `go build ./...`, `go vet ./...`, `go test -p 2 ./... -race -count=1 -cover`, `go mod tidy` and dependency diff check passed. Weighted coverage profile: `go test -p 2 ./... -count=1 -coverprofile=/tmp/fraud-coverage.out` followed by `go run ./scripts/coveragegate /tmp/fraud-coverage.out`; non-cmd608/705=86.24%, domain76/76=100%. Docker, hosted CI and production deployment are not verified.

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
| Validated environment, bounded serving, health/readiness/runtime metrics, loopback-only pprof | `service.TestConfiguration`, `service.TestOperationalRoutesMetricsAndUnavailableReadiness`, `service.TestRequestTimeoutCancelsStoreAndReturnsJSON`, `service.TestMetricsExposeRuntimeResourcesAndPprofIsLoopbackOnly` |
| Operational fallback/panic JSON contains matching request ID | `service.TestOperationalPanicIsContained`, `service.TestOperationalFallbackErrorsAreJSON` |
| Real listener graceful cancellation and serve/listen/startup failures | `service.TestServeRealListenerGracefulCancellationAndServeFailure`, `service.TestRunStartupAndListenFailuresAreBoundedAndSanitized` |
| Safe retry after local write commits but HTTP response is canceled | `service.TestRetryAfterCommittedButCanceledHTTPResponse`, memory fault wrapper/channel synchronization; not a real-process SQL crash drill |
| Published route/status sets are parsed and exercised | `api.TestOpenAPIParsedRoutesAndStatuses`, `service.TestOperationalRoutesMetricsAndUnavailableReadiness`, related error-path tables |
| Existing v1 success/error bodies match declared response shapes | `api.TestV1ResponsesMatchParsedSchemasForEveryDeclaredStatus` exercises every declared v1 status, all reason enums, canonical explanation DENY versus score BLOCK, empty feedback202, request-ID equality and learned policy |
| Exact PascalCase policy integer serialization and negative contract drift detection | `api.TestRuleConfigSchemaMatchesExtremeExactIntegerSerialization`, `TestResponseSchemaChecksRejectPayloadDrift`, `TestPublishedResponseReferencesAndSchemaDefinitionsResolve`; int64-max amount configuration is preserved without floats |
| Production-router R1:10,000 logical score calls, 2% identical duplicates, transient 503/429, 1% attempt timeouts, terminal responses and one decision per unique identity | `playback.TestR1ProductionRouterTenThousand`; actual loopback HTTP production router, Go1.27.2 windows/amd64, 8.64s, 9,800 stored decisions |
| Preserved legacy playback regression:1000 score calls and exact replay | `playback.TestThousandPerMinuteAndPlayback`; HTTP httptest |
| Client400 not retried; transient503 retried with bounded attempts | `playback.TestDoesNotRetryClientError`, `TestRetriesTransientFailure`, `TestStopsAfterMaxAttempts`; playback helper only, not a downstream production client |
| Domain>=90%, weighted non-cmd>=80%, race clean | Latest domain76/76=100.00%, weighted608/705=86.24%; API96.0/query95.9/repository93.5/service90.0/playback74.7. `scripts.TestCoverageGateUsesWeightedStatementsAndDomainThreshold` asserts weighting/thresholds/excluded cmd behavior; Go gate also accepts zero-statement profile entries. |
| Memory application-handler p99<5ms on measured isolated-account workload | `go run ./cmd/latency -samples 10000 -raw`: unique5459ns, replay4209ns; [environment and raw-sample scope](../docs/performance.md). Not HTTP/SQL latency or concurrent throughput |

## Partial

| Acceptance | Implemented subset / missing work |
| --- | --- |
| Scalable durable repository | SQL durability/parity proven, but singleton JSONB loads full history and serializes readers/writers. No relational per-event/account-time indexing or retention. Memory scoped account history is also unbounded. |
| Full semantic OpenAPI validation | Explanation/rules success schemas and targeted v1 response-shape checks are implemented. The helper supports objects/arrays/strings/exact integer text, enums, required/extra properties, min/max/minLength and local references. It is not a full request/schema validator (patterns, general formats, cross-field relations and arbitrary OpenAPI composition are not verified). |
| Under-two-minute clean-clone run | Minimal Go-only memory instructions supplied; cold download/clean-clone wall time has not been measured or independently verified. |
| All error/shutdown branches and exhaustive input/path proof | Domain has100% statement coverage and explicit validation/money/decision invariants asserted, but that is not a proof over all inputs/branches. Service remains91.3%; shutdown/error-path completeness is not claimed. |
| Hiring-manager documentation / CI | README, diagram, curl, glossary, ADRs and measured report supplied; workflow and weighted gate specified/tested locally, but hosted workflow remains UNVERIFIED: not run. |

## Unmet / not claimed

| Acceptance | Reason |
| --- | --- |
| Separate-process cross-service R2 soak: 60 consecutive 10k/min windows, failures, retries, reconciliations and resource trend | Manager-owned integration/load effort; this service's R1 production-router test is not the five-binary platform soak. |
| Production migrations, least-privilege rollout, backup/restore and HTTP/SQL crash recovery | Constructor DDL and synthetic DB restart test only; legacy migration is not the current adapter schema. |
| Authentication/authorization, real PII/payment data or production fraud accuracy | Deliberately synthetic unauthenticated local demonstration; not safe to expose publicly. |
| Hosted CI execution, Linux/Windows runtime verification, Docker build/deploy | Not run here; no Docker dependency/artifact claim. |
| Tag/release/deployment | No tag, release or deployment is claimed; main commits are published through the authorized repository gate. |

## Not applicable

Fraud performs no downstream HTTP calls or downstream writes, so downstream outbox, retry-with-jitter and circuit-breaker requirements do not apply to this service. Playback backoff tests do not change that classification.
