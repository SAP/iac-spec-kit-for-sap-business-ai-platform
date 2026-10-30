# BTP Platform Validation Capability

## Purpose

Defines the optional live SAP BTP validation route recorded at initialization and shared by generated agent skills.

## Requirements

### Requirement: initialize local capabilities
`sap-iac init` SHALL detect the `btp` CLI and BTP MCP availability per selected agent, then write an untracked `.sap-iac/platform-validation.md` record. The CLI is preferred when both routes are available. Missing routes SHALL warn but SHALL NOT block initialization.

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

### Requirement: Gemini BTP MCP discovery
For selected Gemini projects, initialization SHALL discover BTP MCP servers under `mcpServers` in `~/.gemini/settings.json` and the target project's `.gemini/settings.json`. BTP server identification SHALL use the existing case-insensitive server-name matching convention. Gemini availability SHALL be recorded per agent in `.sap-iac/platform-validation.md` under the existing CLI-first routing contract, without persisting settings or credentials.

#### Scenario: user-scoped Gemini BTP server
- **WHEN** user Gemini settings contain a discoverable BTP MCP server and the BTP CLI is unavailable
- **THEN** the record marks Gemini MCP available and supports the MCP validation route

#### Scenario: target-project Gemini BTP server
- **WHEN** Gemini settings in the initialized target project contain a discoverable BTP MCP server
- **THEN** Gemini availability is detected from that target project even when initialization was launched from its parent directory

#### Scenario: CLI takes precedence
- **WHEN** both the BTP CLI and Gemini BTP MCP are detected
- **THEN** the record marks both available and selects the CLI as the preferred route

#### Scenario: missing Gemini BTP route
- **WHEN** neither the BTP CLI nor a Gemini BTP MCP server is detected
- **THEN** the record marks Gemini MCP unavailable
- **AND** Gemini workflows retain user input without blocking under the existing fallback contract

#### Scenario: configuration remains private
- **WHEN** Gemini settings include credentials, environment variables, or server endpoints
- **THEN** the capability record contains route availability only and does not copy those settings
