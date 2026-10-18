## ADDED Requirements

### Requirement: Inferred account-window structuring
The service SHALL infer STRUCTURING for the third and subsequent unique same-account attempt in the active window when the current and at least two prior amounts lie in [band-floor(band/10), band], excluding zero amounts. The prior qualifying count SHALL be stored. Legacy caller structuring=true SHALL remain an additional signal, never a way to suppress inferred detection.

#### Scenario: Replay is not a third event
- **GIVEN** two unique near-band account attempts and arbitrary identical retries
- **WHEN** the second attempt is replayed
- **THEN** its original result remains unchanged
- **AND** the next unique near-band attempt receives STRUCTURING exactly once.

### Requirement: Scoped memory atomicity
Application memory transactions SHALL select only addressed event/feedback and requested account history, with global policy included. Out-of-scope writes or collisions with excluded events SHALL fail atomically. Full Update and SQL fallback SHALL preserve application semantics.

#### Scenario: Scope collision
- **GIVEN** an event belonging to an excluded account
- **WHEN** a callback attempts to overwrite its identity through another account scope
- **THEN** the complete update is rejected without losing original state.

### Requirement: Retry after ambiguous HTTP response
A canceled response after commit SHALL NOT cause duplicate effects when the same input and event_id are retried.

#### Scenario: Commit before cancellation
- **GIVEN** a fault wrapper that commits an event before waiting for request cancellation
- **WHEN** the response fails and the same event is retried
- **THEN** the original result is returned and the repository contains one event with one velocity/pattern contribution.
