## ADDED Requirements

### Requirement: Honest reproducible engineering evidence
Documentation SHALL distinguish measured memory application latency, httptest playback and real-process cross-service load. Every Met acceptance row SHALL name an asserting test or actual timing command. Missing runtime policy editing, scalable SQL and unexecuted workflow/deployment evidence SHALL remain explicitly Partial/Unmet.

#### Scenario: Weighted coverage gate
- **GIVEN** a profile containing domain, other non-cmd and uncovered cmd statements
- **WHEN** the coverage gate is executed
- **THEN** non-cmd coverage is weighted by statement counts with a minimum80%
- **AND** domain coverage independently requires90%
- **AND** missing domain/empty profiles fail.

#### Scenario: Workflow versus local evidence
- **GIVEN** the workflow YAML parses and its gate script tests pass locally
- **WHEN** documentation describes CI
- **THEN** hosted workflow execution remains UNVERIFIED until it actually runs.
