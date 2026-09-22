# Generate Command Capability

## Purpose

Defines the behaviour of `/btp-iac.generate`: running a full pre-generation governance validation pass, generating Terraform HCL in dependency order, and validating the output.

## Requirements

### Requirement: run pre-generation governance validation
The command SHALL run a full pre-generation validation pass across all six governance categories before writing any Terraform HCL if `memory/governance.md` exists.

#### Scenario: governance file present
- **WHEN** `memory/governance.md` exists
- **THEN** the command validates every resource in `specs/tasks.md` against all six categories before writing a single file

#### Scenario: governance file absent
- **WHEN** `memory/governance.md` does not exist
- **THEN** the command proceeds without governance constraints and notes this

### Requirement: validate regions before generating
The command SHALL validate each subaccount resource region against governance rules during the pre-generation pass.

#### Scenario: region violation — hard block
- **WHEN** a subaccount region is not allowed or is forbidden
- **THEN** the command stops with the region, rule violated, and fix instructions
- **UNLESS** `- Override: true` is set

#### Scenario: preferred provider mismatch
- **WHEN** a preferred infrastructure provider is configured and the targeted platform lookup returns different unambiguous provider metadata for a generated subaccount region
- **THEN** the command logs a warning with the region and both providers
- **AND** continues generation without requiring an override

### Requirement: validate naming before generating
The command SHALL validate each subaccount resource name and environment tier against governance naming rules.

#### Scenario: naming violation — hard block
- **WHEN** a name does not match the required pattern or tier is undefined
- **THEN** the command stops with the name, expected pattern, and fix instructions
- **UNLESS** `- Override: true` is set

### Requirement: validate account environments before generating
The command SHALL validate each Cloud Foundry and Kyma environment resource against the permitted runtime environment list and its applicable naming pattern, including Cloud Foundry spaces.

#### Scenario: account-environment violation — hard block
- **WHEN** an environment type is not allowed or an organization, environment, or space name does not match its required pattern
- **THEN** the command stops with the violated rule and fix instructions
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

### Requirement: select provider by service instance location
The command SHALL generate each service instance with the provider indicated by its `location`: BTP provider for `btp`, Cloud Foundry provider for `cf`. Entitlement-only services SHALL generate only the entitlement assignment. The `required_providers` block SHALL include every provider the resolved locations require.

#### Scenario: btp service instance
- **WHEN** a service instance task has `location: btp`
- **THEN** the command generates a BTP-provider service instance resource

#### Scenario: cf service instance
- **WHEN** a service instance task has `location: cf`
- **THEN** the command generates a Cloud Foundry-provider service instance resource scoped to the task's `cf_space`

#### Scenario: entitlement-only service
- **WHEN** a service is classified as entitlement-only
- **THEN** the command generates only the entitlement assignment and no instance or subscription resource

### Requirement: emit available subaccount classification attributes
For each `btp_subaccount` task selected by the stage filter, the command SHALL read `usage` and `beta_enabled` from task metadata and emit every present, valid value on the generated `btp_subaccount` resource. It SHALL NOT infer defaults or substitute governance values. Missing values are supported for legacy tasks and SHALL NOT stop generation; invalid values that are present SHALL stop before writing that resource.

#### Scenario: classified subaccount generated
- **WHEN** a selected `btp_subaccount` task contains valid `usage` and `beta_enabled` metadata
- **THEN** the generated `btp_subaccount` resource contains the same `usage` and `beta_enabled` values

#### Scenario: subaccount classification metadata missing
- **WHEN** a selected `btp_subaccount` task omits `usage` or `beta_enabled` metadata
- **THEN** the command omits the unavailable Terraform attribute, identifies the task as using legacy classification metadata, and continues

#### Scenario: missing metadata on an unselected stage
- **WHEN** a `btp_subaccount` task outside the selected stages omits `usage` or `beta_enabled` metadata
- **THEN** the command skips that task without checking its classification metadata

#### Scenario: invalid subaccount classification metadata
- **WHEN** a selected `btp_subaccount` task contains an invalid `usage` or `beta_enabled` value
- **THEN** the command stops before writing that resource and identifies the invalid task metadata

### Requirement: resolve latest provider versions at runtime
The command SHALL look up the current latest version of each required Terraform provider before writing `versions.tf`, and use those versions as `~>` constraints in `required_providers`. It SHALL NOT hardcode any version. Before using WebFetch, it SHALL check if the `terraform` MCP server is available and prefer it.

#### Scenario: provider version resolved
- **WHEN** generating `versions.tf`
- **THEN** the command looks up the latest version for each provider via the terraform MCP server or WebFetch fallback, and uses it as the `~>` constraint

### Requirement: run terraform init before fmt and validate
The command SHALL run `terraform init` on the `terraform/` directory before `terraform fmt` and `terraform validate`.

#### Scenario: init succeeds
- **WHEN** all files are written and `terraform init` succeeds
- **THEN** the command proceeds to `terraform fmt --recursive` then `terraform validate`

#### Scenario: init fails
- **WHEN** `terraform init` fails
- **THEN** the command reports the error and does not proceed to fmt or validate

### Requirement: run terraform fmt and validate on completion
The command SHALL run `terraform fmt --recursive` and `terraform validate` after `terraform init` and report the outcome.

#### Scenario: fmt and validate pass
- **WHEN** all files are written and both commands succeed
- **THEN** the command reports success

#### Scenario: fmt or validate fails — fix and retry
- **WHEN** `terraform fmt --recursive` or `terraform validate` fails
- **THEN** the command fixes the reported issues in the affected files, then re-runs `terraform fmt --recursive` and `terraform validate` until both pass

### Requirement: do not commit generated code by default
The command SHALL NOT run `git commit` (or stage files) after generating Terraform HCL unless the user explicitly requests a commit.

#### Scenario: default post-generate state
- **WHEN** generation and validation succeed
- **THEN** the generated files are left as unstaged working-tree changes for the user to review

#### Scenario: user requests commit
- **WHEN** the user explicitly asks to commit (e.g. "commit", "git commit", "commit the changes")
- **THEN** the command may stage and commit the generated files

### Requirement: do not push generated code by default
The command SHALL NOT run `git push` after generating Terraform HCL unless the user explicitly requests a push.

#### Scenario: default post-generate state
- **WHEN** generation and validation succeed
- **THEN** no push is performed

#### Scenario: user requests push
- **WHEN** the user explicitly asks to push (e.g. "push", "git push")
- **THEN** the command may push the changes
