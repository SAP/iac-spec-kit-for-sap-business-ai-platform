# Accounts Command Capability

## Purpose

Defines the behaviour of `/sap-iac.accounts`: defining the BTP account topology with governance validation, producing `specs/landscape.md`.

## Requirements

### Requirement: read inputs
The command SHALL read `specs/scenario.md`, `memory/global-account.md`, and, if present, `memory/governance.md` before defining any account topology.

#### Scenario: inputs read
- **WHEN** the command starts
- **THEN** it reads `specs/scenario.md`, `memory/global-account.md`, and optionally `memory/governance.md`

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
The command SHALL use the global account subdomain from `memory/global-account.md` when present, without requesting it, and validate each selected runtime environment type and applicable name against Account Setup and Naming rules.

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

#### Scenario: preferred provider mismatch
- **WHEN** a preferred infrastructure provider is configured and the targeted platform lookup returns different unambiguous provider metadata for a subaccount region
- **THEN** the command warns with the region and both providers
- **AND** continues without requiring an override

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
- **THEN** `specs/landscape.md` exists and is the authoritative input for `/sap-iac.services`, `/sap-iac.security`, and `/sap-iac.generate`

### Requirement: validate platform availability when values are supplied
The command SHALL apply the shared platform-validation capability to topology regions and any named service offering, subscription, or plan pair. CLI checks SHALL target the resolved global-account subdomain and ignore NEO regions.

#### Scenario: available Cloud Foundry region
- **WHEN** a topology region is returned by the targeted platform lookup and is not labelled NEO
- **THEN** the command accepts the region

### Requirement: infer usage and beta_enabled from tier
The command SHALL infer `usage` and `beta_enabled` for each subaccount from its environment tier name using case-insensitive, hyphen-delimited token matching, unless an exact governance classification is present.

Matching rules:
- Tier contains token `prod` or `production` → `usage = USED_FOR_PRODUCTION`, `beta_enabled = false`
- Tier contains token `dev` or `development` → `usage = NOT_USED_FOR_PRODUCTION`, `beta_enabled = true`
- Tier contains token `test`, `staging`, or `qa` → `usage = NOT_USED_FOR_PRODUCTION`, `beta_enabled = true`
- Tier does not match any rule → unmatched; see clarification requirement below

An exact, case-insensitive normalized tier name in `memory/governance.md` under `## Account Setup → Tier classifications` overrides token matching. If more than one token rule matches, use this explicit precedence: production, development, then test/staging/qa.

#### Scenario: production tier inferred
- **WHEN** a subaccount's tier contains the `prod` or `production` token and no governance classification is present
- **THEN** `usage` is set to `USED_FOR_PRODUCTION` and `beta_enabled` is set to `false`

#### Scenario: development tier inferred
- **WHEN** a subaccount's tier contains the `dev` or `development` token and no governance classification is present
- **THEN** `usage` is set to `NOT_USED_FOR_PRODUCTION` and `beta_enabled` is set to `true`

#### Scenario: test/staging/qa tier inferred
- **WHEN** a subaccount's tier contains a `test`, `staging`, or `qa` token and no governance classification is present
- **THEN** `usage` is set to `NOT_USED_FOR_PRODUCTION` and `beta_enabled` is set to `true`

#### Scenario: governance classification applied
- **WHEN** `memory/governance.md` contains an exact `Tier classifications` entry for a tier
- **THEN** its explicit `usage` and `beta_enabled` values override keyword inference for that tier

#### Scenario: compound tier matches multiple rules
- **WHEN** a tier such as `dev-prod` contains more than one classification token and has no governance classification
- **THEN** the command applies the documented production, development, then test/staging/qa precedence

### Requirement: clarify unmatched tiers before proceeding
The command SHALL stop and ask the user to specify `usage` and `beta_enabled` for any tier that does not match any inference rule. All unmatched tiers SHALL be presented in a single question before the confirmation table is shown.

#### Scenario: one or more tiers unmatched
- **WHEN** one or more subaccount tiers do not match any keyword rule and have no governance override
- **THEN** the command presents all unmatched tiers in a single question asking the user to specify `usage` and `beta_enabled` for each
- **AND** does not proceed to the confirmation table until the user has answered

#### Scenario: all tiers matched
- **WHEN** every subaccount tier matches a keyword rule or has a governance override
- **THEN** the command proceeds directly to the confirmation table without asking

### Requirement: confirm usage and beta_enabled before writing
The command SHALL present a single summary table of all subaccounts with their inferred or clarified `usage` and `beta_enabled` values and require user confirmation before writing `specs/landscape.md`.

#### Scenario: user accepts inferred values
- **WHEN** the confirmation table is shown and the user accepts
- **THEN** the command writes `usage` and `beta_enabled` for each subaccount into `specs/landscape.md` as confirmed

#### Scenario: user overrides a row
- **WHEN** the user specifies a correction for one or more rows in the confirmation table
- **THEN** the command applies those corrections and writes the updated values into `specs/landscape.md`
