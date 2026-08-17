# Generate Command Capability

## Purpose

Defines the behaviour of `/btp-iac.generate`: running a full pre-generation governance validation pass, generating Terraform HCL in dependency order, and validating the output.

## Requirements

### Requirement: run pre-generation governance validation
The command SHALL run a full pre-generation validation pass across all five governance categories before writing any Terraform HCL if `memory/governance.md` exists.

#### Scenario: governance file present
- **WHEN** `memory/governance.md` exists
- **THEN** the command validates every resource in `specs/tasks.md` against all five categories before writing a single file

#### Scenario: governance file absent
- **WHEN** `memory/governance.md` does not exist
- **THEN** the command proceeds without governance constraints and notes this

### Requirement: validate regions before generating
The command SHALL validate each subaccount resource region against governance rules during the pre-generation pass.

#### Scenario: region violation — hard block
- **WHEN** a subaccount region is not allowed or is forbidden
- **THEN** the command stops with the region, rule violated, and fix instructions
- **UNLESS** `- Override: true` is set

### Requirement: validate naming before generating
The command SHALL validate each subaccount resource name and environment tier against governance naming rules.

#### Scenario: naming violation — hard block
- **WHEN** a name does not match the required pattern or tier is undefined
- **THEN** the command stops with the name, expected pattern, and fix instructions
- **UNLESS** `- Override: true` is set

### Requirement: validate service plans before generating
The command SHALL validate each service instance resource plan against governance service plan rules.

#### Scenario: service plan violation — hard block
- **WHEN** a plan is not permitted or is forbidden for the environment tier
- **THEN** the command stops with the plan, tier, permitted plans, and fix instructions
- **UNLESS** `- Override: true` is set

### Requirement: validate security before generating
The command SHALL validate trust configuration decisions against governance security rules.

#### Scenario: security violation — hard block
- **WHEN** a trust configuration violates a security rule
- **THEN** the command stops with the specific rule and fix instructions
- **UNLESS** `- Override: true` is set

### Requirement: validate cost controls before generating
The command SHALL apply cost controls governance during the pre-generation pass.

#### Scenario: metered service warning
- **WHEN** `## Cost Controls → Metered service warning: enabled`
- **THEN** the command outputs a warning listing metered services before proceeding (does not block)

#### Scenario: cost centre tag missing — hard block
- **WHEN** `## Cost Controls → Cost centre tag: required` and a subaccount resource lacks the tag attribute
- **THEN** the command stops with the subaccount name and fix instructions
- **UNLESS** `- Override: true` is set

### Requirement: generate Terraform HCL in dependency order
The command SHALL execute each task from `specs/tasks.md` in dependency order, writing resources to the file paths annotated by `/btp-iac.design`.

#### Scenario: resources generated
- **WHEN** the pre-generation pass passes
- **THEN** the command writes Terraform HCL for each task to its annotated file path

### Requirement: run terraform fmt and validate on completion
The command SHALL run `terraform fmt` and `terraform validate` after all files are written and report the outcome.

#### Scenario: validate passes
- **WHEN** all files are written and `terraform validate` succeeds
- **THEN** the command reports success

#### Scenario: validate fails
- **WHEN** `terraform validate` fails
- **THEN** the command identifies the failing resource and the likely cause
