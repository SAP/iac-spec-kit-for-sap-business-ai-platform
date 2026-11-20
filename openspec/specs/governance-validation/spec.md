# Governance Validation Capability

## Purpose

Defines how downstream commands (`/sap-iac.accounts`, `/sap-iac.services`, `/sap-iac.security`, `/sap-iac.generate`) read and enforce governance rules from `memory/governance.md`.
## Requirements
### Requirement: governance file read before generating output
Each downstream command (`/sap-iac.accounts`, `/sap-iac.services`, `/sap-iac.security`, `/sap-iac.generate`) SHALL read `memory/governance.md` at the start of execution if it exists.

#### Scenario: governance file present
- **WHEN** `memory/governance.md` exists
- **THEN** the command loads all governance rules before making any decisions

#### Scenario: governance file absent
- **WHEN** `memory/governance.md` does not exist
- **THEN** the command proceeds without governance constraints and notes that no governance rules are in effect

### Requirement: hard block on violation
A downstream command SHALL stop and report a violation when a decision contradicts a governance rule, unless override is active.

#### Scenario: region violation blocked
- **WHEN** a subaccount is being placed in a region not in the allowed list
- **THEN** the command stops with a message identifying the forbidden region, the governance rule violated, and instructions to either change the region or set `- Override: true` in governance.md

#### Scenario: naming violation blocked
- **WHEN** a generated subaccount name does not match the required pattern
- **THEN** the command stops with a message showing the expected pattern and the non-conforming name

#### Scenario: account-environment violation blocked
- **WHEN** a Cloud Foundry or Kyma environment is selected outside the allowed environment list, or its name or a CF space name violates the applicable pattern
- **THEN** the command stops with a message showing the violated Account Setup or Naming rule

#### Scenario: service plan violation blocked
- **WHEN** a service instance is assigned a plan not permitted for its environment tier
- **THEN** the command stops with a message identifying the plan, the tier, and the permitted plans for that tier

#### Scenario: security violation blocked
- **WHEN** a configuration decision contradicts a security rule (e.g. no custom IdP configured when required)
- **THEN** the command stops with a message identifying the specific security rule violated

### Requirement: advisory provider preference validation
When a preferred infrastructure provider other than `none` is set and region provider metadata is available, commands that validate regions SHALL compare it to the preference. They SHALL not infer a provider from a region code.

#### Scenario: provider mismatch
- **WHEN** an unambiguous region provider differs from the preferred provider
- **THEN** the command logs a warning naming the region, preferred provider, and returned provider
- **AND** continues without requiring `- Override: true`

#### Scenario: provider metadata unavailable
- **WHEN** the platform route, provider metadata, or its mapping to a governed provider is unavailable
- **THEN** the command retains the existing allowed/forbidden-region outcome and reports the provider check as advisory

### Requirement: override bypasses hard block
When `memory/governance.md` contains `- Override: true`, downstream commands SHALL warn about violations but continue rather than blocking.

#### Scenario: override active with violation
- **WHEN** a governance violation is detected and `- Override: true` is set
- **THEN** the command logs a warning identifying the violation and the override
- **AND** continues generating output

### Requirement: reusable module consumption
Downstream governance-aware commands SHALL reuse a module recorded under `## Reusable Modules` in `memory/governance.md` only in the context recorded for that module, and SHALL retain the module's source or documentation reference in the specifications or Terraform output they generate. This guidance is advisory: it SHALL NOT hard-block generation, and commands SHALL NOT be required to emit a Terraform `module` block.

#### Scenario: module reused in approved context
- **WHEN** a command generates output that matches the approved context of a recorded reusable module
- **THEN** the command reuses the recorded module rather than recreating an equivalent pattern
- **AND** carries the module's `Source` or `Documentation` reference into the generated specification or Terraform output

#### Scenario: module not applicable to the context
- **WHEN** no recorded reusable module applies to the context of the output being generated
- **THEN** the command generates the output normally without reusing a module
- **AND** does not block or warn solely because a module exists for a different context

#### Scenario: reuse section records none
- **WHEN** `## Reusable Modules` contains `- Reuse: none`, or the section is absent
- **THEN** downstream commands generate output without reusing any module

### Requirement: subaccount label application
When `memory/governance.md` records a labeling strategy under `## Subaccount Labels` (anything other than `- Labels: none`), `/sap-iac.generate` SHALL emit a `labels` map on each generated `btp_subaccount` resource, applying every label whose applicability is `all` plus any label scoped to that subaccount, and honouring the recorded per-subaccount assignments and exceptions. A `required` label SHALL be present on every subaccount it applies to. The governed `cost-center` label SHALL satisfy the `## Cost Controls` cost-centre-tag requirement rather than being emitted as a separate attribute. This guidance is advisory: it SHALL respect the `- Override: true` flag and SHALL NOT introduce a hard block beyond the existing cost-centre-tag check.

#### Scenario: labels emitted on generated subaccounts
- **WHEN** generate produces a `btp_subaccount` resource and `## Subaccount Labels` governs labels that apply to it
- **THEN** the resource carries a `labels` map containing those labels with their governed values

#### Scenario: per-subaccount exception honoured
- **WHEN** a subaccount has a per-subaccount assignment or an `except <label>` exception recorded under `## Subaccount Labels`
- **THEN** generate applies the override value or omits the excepted label for that subaccount

#### Scenario: cost-centre tag satisfied by the cost-center label
- **WHEN** `## Cost Controls` requires a cost-centre tag and a `cost-center` label is governed
- **THEN** the cost-centre-tag requirement is satisfied by the `cost-center` entry of the subaccount `labels` map
- **AND** generate does not emit a separate cost-centre attribute

#### Scenario: no labels governed
- **WHEN** `## Subaccount Labels` contains `- Labels: none`, or the section is absent
- **THEN** generate does not emit a governed `labels` map on that basis

