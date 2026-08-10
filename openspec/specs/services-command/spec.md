# Services Command Capability

## Purpose

Defines the behaviour of `/btp-iac.services`: resolving BTP service dependencies with governance validation, producing `specs/services.md`.

## Requirements

### Requirement: read inputs
The command SHALL read `specs/scenario.md`, `specs/landscape.md`, and if present `memory/governance.md` before resolving any services.

#### Scenario: inputs read
- **WHEN** the command starts
- **THEN** it reads both spec files and optionally governance

### Requirement: resolve service dependencies
The command SHALL produce a dependency-ordered list of BTP entitlements, subscriptions, and service instances per subaccount.

#### Scenario: services resolved
- **WHEN** inputs are read
- **THEN** the command resolves required entitlements, SaaS subscriptions, service instances with configuration, and inter-service dependencies in order

### Requirement: validate service plans against governance
The command SHALL validate each service instance plan against governance rules if `memory/governance.md` exists.

#### Scenario: permitted plan
- **WHEN** the plan is in `## Service Plans → <tier> → Permitted` and not in `Forbidden`
- **THEN** the command continues

#### Scenario: forbidden plan — hard block
- **WHEN** the plan is not permitted or is explicitly forbidden for the environment tier
- **THEN** the command stops with the plan, tier, permitted plans, and fix instructions
- **UNLESS** `- Override: true` is set, in which case a warning is logged and the command continues

### Requirement: write services file
The command SHALL write `specs/services.md` with the full dependency-ordered service list.

#### Scenario: services file written
- **WHEN** the command completes successfully
- **THEN** `specs/services.md` exists and is the direct input to `/btp-iac.tasks`
