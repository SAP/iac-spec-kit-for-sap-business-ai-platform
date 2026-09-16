# Security Command Capability

## Purpose

Defines the behaviour of `/btp-iac.security`: defining authentication and authorisation configuration with governance validation, producing `specs/trust.md`.

## Requirements

### Requirement: read inputs
The command SHALL read `specs/scenario.md`, `specs/landscape.md`, and if present `memory/governance.md` before defining any security configuration.

#### Scenario: inputs read
- **WHEN** the command starts
- **THEN** it reads both spec files and optionally governance

### Requirement: define security configuration
The command SHALL define for each subaccount: IdP trust configurations, role collections with role template assignments, user and group assignments.

#### Scenario: security configuration defined
- **WHEN** inputs are read
- **THEN** the command produces a complete security configuration for each subaccount

### Requirement: validate IdP against governance
The command SHALL validate the custom IdP configuration against governance rules if `memory/governance.md` exists.

#### Scenario: custom IdP required and configured
- **WHEN** `Custom IdP: required` and a custom IdP trust is being configured
- **THEN** the command continues

#### Scenario: custom IdP required but missing — hard block
- **WHEN** `Custom IdP: required` and no custom IdP is configured
- **THEN** the command stops with the rule and fix instructions
- **UNLESS** `- Override: true` is set

### Requirement: validate role collections against governance
The command SHALL validate default role collection assignments against governance rules.

#### Scenario: required assignments missing — hard block
- **WHEN** `Default role collections` specifies assignments that are not included
- **THEN** the command stops with the required assignments and fix instructions
- **UNLESS** `- Override: true` is set

### Requirement: write trust file
The command SHALL write `specs/trust.md` with the complete security configuration.

#### Scenario: trust file written
- **WHEN** the command completes successfully
- **THEN** `specs/trust.md` exists and is the direct input to `/btp-iac.tasks`
