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
Every generated skill SHALL read the capability record before a live platform validation it requires. It SHALL use the recorded CLI first, then an equivalent BTP MCP read/list operation for the current configured agent. If no route is recorded, it SHALL accept user input without blocking. A failed recorded route SHALL require the user to resolve setup; a successful unavailable result SHALL require a replacement value.

#### Scenario: no route recorded
- **WHEN** a skill requires a live platform check and its record has no available route
- **THEN** it retains the user-provided value without blocking

### Requirement: BTP access is read-only
Every generated skill SHALL restrict BTP CLI and BTP MCP activity to read or list retrievals. An MCP tool may be invoked only when it is explicitly documented as a read/list lookup. The sole CLI exception is `btp target --global-account <subdomain>` as the account-selection prelude immediately before read/list CLI commands. No skill SHALL invoke, suggest, or approve BTP create, update, delete, assign, unassign, enable, disable, or other mutating operations, including when a user requests them.

#### Scenario: user asks for a BTP mutation
- **WHEN** a user asks an agent command to create, update, or delete BTP state through the CLI or MCP
- **THEN** the command does not invoke a BTP mutation and confines any live lookup to read/list activity

### Requirement: region validation sources
CLI region checks SHALL target the global account subdomain and use `btp list accounts/available-region`; all NEO-labelled entries SHALL be ignored. When the returned region data includes infrastructure-provider metadata, skills MAY compare it with the preferred provider read from `memory/governance.md` after mapping only unambiguous labels; they SHALL not infer the provider from a region code. The local capability record SHALL contain only route availability, not provider preferences or response-field metadata. If no live region route exists, the SAP Help Cloud Foundry region list SHALL be advisory only.

#### Scenario: NEO region returned
- **WHEN** a region lookup returns an entry labelled NEO
- **THEN** the command excludes it from its validation result
