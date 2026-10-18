## MODIFIED Requirements

### Requirement: Persist replay-safe scoring decisions
The service SHALL atomically record the input, decision, timestamp, effective recent count and rule snapshot. Identical replay SHALL return the original decision without another velocity contribution. Changed payload SHALL be rejected. The memory adapter provides process-lifetime persistence only.

#### Scenario: Concurrent replay and shared handlers
- **GIVEN** handlers sharing one repository
- **WHEN** identical event identities are submitted concurrently
- **THEN** exactly one event contributes to account velocity
- **AND** every replay returns the original decision.

## ADDED Requirements

### Requirement: Account-window velocity
The service SHALL count prior unique events for the same account in (now-60s, now], using an injected clock. Legacy recent_count SHALL only increase the effective count.

#### Scenario: Exact lower boundary and clock rollback
- **GIVEN** prior events at now-60s or later than now
- **WHEN** a new event is scored
- **THEN** neither boundary-expired nor future events contribute.
