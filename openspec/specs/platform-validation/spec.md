# BTP Platform Validation Capability

## Purpose

Defines the optional live SAP BTP validation route recorded at initialization and shared by generated agent skills.

## Requirements

### Requirement: initialize local capabilities
`btp-iac init` SHALL detect the `btp` CLI and BTP MCP availability per selected agent, then write an untracked `.btp-iac/platform-validation.md` record. The CLI is preferred when both routes are available. Missing routes SHALL warn but SHALL NOT block initialization.

#### Scenario: CLI available
- **WHEN** `btp` is on PATH during initialization
- **THEN** the record marks the CLI available and selects it as the preferred route

### Requirement: shared skill routing
Every generated skill SHALL read the capability record before a live platform validation it requires. It SHALL use the recorded CLI first, then an equivalent BTP MCP operation for the current configured agent. If no route is recorded, it SHALL accept user input without blocking. A failed recorded route SHALL require the user to resolve setup; a successful unavailable result SHALL require a replacement value.

#### Scenario: no route recorded
- **WHEN** a skill requires a live platform check and its record has no available route
- **THEN** it retains the user-provided value without blocking

### Requirement: region validation sources
CLI region checks SHALL target the global account subdomain and use `btp list accounts/available-region`; all NEO-labelled entries SHALL be ignored. If no live region route exists, the SAP Help Cloud Foundry region list SHALL be advisory only.

#### Scenario: NEO region returned
- **WHEN** a region lookup returns an entry labelled NEO
- **THEN** the command excludes it from its validation result
