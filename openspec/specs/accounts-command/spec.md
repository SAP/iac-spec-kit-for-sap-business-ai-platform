# Accounts Command Capability

## Purpose

Defines the behaviour of `/btp-iac.accounts`: defining the BTP account topology with governance validation, producing `specs/landscape.md`.

## Requirements

### Requirement: read inputs
The command SHALL read `specs/scenario.md` and, if present, `memory/governance.md` before defining any account topology.

#### Scenario: inputs read
- **WHEN** the command starts
- **THEN** it reads `specs/scenario.md` and optionally `memory/governance.md`

### Requirement: define account topology
The command SHALL define the BTP account topology: global account subdomain, directory groupings, subaccounts per environment tier with name, region, description, and subdomain, and optional Cloud Foundry and Kyma environments.

#### Scenario: topology defined
- **WHEN** inputs are read
- **THEN** the command produces a complete account topology covering all environment tiers

### Requirement: collect runtime environment setup
For every subaccount, the command SHALL determine whether Cloud Foundry, Kyma, both, or no runtime environment is required. When Cloud Foundry is selected, it SHALL determine whether spaces are required and collect their names.

#### Scenario: environment setup missing from inputs
- **WHEN** the scenario and governance do not fully determine a subaccount's runtime environment setup
- **THEN** the command asks targeted questions before writing `specs/landscape.md`

### Requirement: validate account setup governance
The command SHALL use the governed global account subdomain when present, and validate each selected runtime environment type and applicable name against Account Setup and Naming rules.

#### Scenario: governed environment violation
- **WHEN** a selected environment type is not allowed or an environment, organization, or space name violates its pattern
- **THEN** the command stops with the violated rule and fix instructions
- **UNLESS** `- Override: true` is set, in which case a warning is logged and the command continues

### Requirement: validate regions against governance
The command SHALL validate each subaccount region against governance rules if `memory/governance.md` exists.

#### Scenario: allowed region
- **WHEN** a subaccount region is in the `## Regions → Allowed` list
- **THEN** the command continues

#### Scenario: forbidden region — hard block
- **WHEN** a subaccount region is not in `## Regions → Allowed` or appears in `Forbidden`
- **THEN** the command stops with the region, the rule violated, and fix instructions
- **UNLESS** `- Override: true` is set, in which case a warning is logged and the command continues

### Requirement: validate naming against governance
The command SHALL validate each subaccount name and environment tier against governance naming rules if `memory/governance.md` exists.

#### Scenario: name matches pattern
- **WHEN** the subaccount name matches `## Naming → Subaccount pattern` and the tier is in `## Naming → Environments`
- **THEN** the command continues

#### Scenario: naming violation — hard block
- **WHEN** the name does not match the pattern or the tier is not defined
- **THEN** the command stops with the name, expected pattern, and fix instructions
- **UNLESS** `- Override: true` is set

### Requirement: write landscape file
The command SHALL write `specs/landscape.md` as the authoritative account structure.

#### Scenario: landscape file written
- **WHEN** the command completes successfully
- **THEN** `specs/landscape.md` exists and is the authoritative input for `/btp-iac.services`, `/btp-iac.security`, and `/btp-iac.generate`

### Requirement: validate platform availability when values are supplied
The command SHALL apply the shared platform-validation capability to topology regions and any named service offering, subscription, or plan pair. CLI checks SHALL target the resolved global-account subdomain and ignore NEO regions.

#### Scenario: available Cloud Foundry region
- **WHEN** a topology region is returned by the targeted platform lookup and is not labelled NEO
- **THEN** the command accepts the region
