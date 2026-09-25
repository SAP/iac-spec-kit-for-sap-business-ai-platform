# Security Command Capability

## Purpose

Defines the behaviour of `/sap-iac.security`: defining authentication and authorisation configuration with governance validation, producing `specs/trust.md`.

## Requirements

### Requirement: read inputs
The command SHALL read `specs/scenario.md`, `specs/landscape.md`, and if present `memory/governance.md` before defining any security configuration.

#### Scenario: inputs read
- **WHEN** the command starts
- **THEN** it reads both spec files and optionally governance

### Requirement: define security configuration
The command SHALL confirm for each subaccount whether role collections are required. When role collections are required, the command SHALL optionally ask which individual roles should be assigned to each collection (each role identified by `role_name`, `role_template_name`, and `role_template_app_id`). Role collections with no roles specified are valid. The command SHALL write each role collection to `specs/trust.md` with its name, optional description, and a structured list of individually named roles (which may be empty). The command SHALL NOT write unstructured "role template assignments" to the trust file.

The command SHALL also define for each subaccount: IdP trust configurations and user and group assignments to role collections.

#### Scenario: role collections required, roles specified
- **WHEN** the user confirms role collections are required and provides individual role entries (role_name, role_template_name, role_template_app_id) for a collection
- **THEN** `specs/trust.md` records that collection with a structured roles list containing each named role

#### Scenario: role collections required, no roles specified
- **WHEN** the user confirms role collections are required but does not specify individual roles
- **THEN** `specs/trust.md` records the collection with an empty roles list and no role assignments

#### Scenario: role collections not required
- **WHEN** the user confirms role collections are not required for a subaccount
- **THEN** no role collection entries appear in `specs/trust.md` for that subaccount

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
- **THEN** `specs/trust.md` exists and is the direct input to `/sap-iac.tasks`
