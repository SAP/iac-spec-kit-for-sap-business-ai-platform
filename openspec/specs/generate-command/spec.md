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
