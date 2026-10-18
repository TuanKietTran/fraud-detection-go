## ADDED Requirements

### Requirement: Durable transactional repository
When a PostgreSQL repository is selected, the service SHALL atomically retain event replay identities, original results, rule snapshots, velocity timestamps and feedback labels across a database restart. Memory and PostgreSQL SHALL implement the same transaction port. Callback errors SHALL roll back the complete update.

#### Scenario: Database restart
- **GIVEN** a scored event and confirmed-fraud feedback committed in PostgreSQL
- **WHEN** the database is stopped and started using the same data directory
- **THEN** a newly opened adapter returns the original explanation and learned threshold
- **AND** replay of feedback does not apply it again.
