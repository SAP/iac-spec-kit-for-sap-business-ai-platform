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
The command SHALL execute each task from `specs/tasks.md` in dependency order, writing resources to the directory and file paths annotated by `/btp-iac.design`, using the standard file layout per configuration unit: resources in `main.tf`, input variables in `variables.tf`, output values in `outputs.tf`, provider configuration and `required_providers` in `providers.tf`, and a `backend.tf` defaulting to a local backend.

#### Scenario: resources generated
- **WHEN** the pre-generation pass passes
- **THEN** the command writes Terraform HCL for each task to its annotated path using the standard `main.tf` / `variables.tf` / `outputs.tf` / `providers.tf` / `backend.tf` layout

#### Scenario: local backend emitted
- **WHEN** a configuration unit is written
- **THEN** it includes a `backend.tf` configured for a local backend by default

### Requirement: select provider by service resource type
The command SHALL generate each service resource with the provider indicated by its `resource_type`:

- `btp_subaccount_service_instance` — BTP provider (`btp_subaccount_service_instance`, using `btp_subaccount_entitlement` / `btp_subaccount_service_plan` as needed).
- `cloudfoundry_service_instance` — Cloud Foundry provider (`cloudfoundry_service_instance`) scoped to the `cf_space` recorded on the task, resolving offering/plan via CF data sources. Use `cloudfoundry/cloudfoundry` as its `required_providers` source.
- `btp_subaccount_subscription` — BTP provider (`btp_subaccount_subscription`) paired with its `btp_subaccount_entitlement`. No `location` or `cf_space` applies.
- `btp_subaccount_entitlement` (entitlement-only) — generate only the entitlement assignment resource; no instance or subscription resource.

Kyma-provider resources SHALL use the `hashicorp/kubernetes` provider. The `required_providers` block in `providers.tf` SHALL include every provider the resolved resource types require.

#### Scenario: btp service instance
- **WHEN** a service instance task has `location: btp`
- **THEN** the command generates a BTP-provider service instance resource

#### Scenario: cf service instance
- **WHEN** a service instance task has `location: cf`
- **THEN** the command generates a Cloud Foundry-provider service instance resource scoped to the task's `cf_space` using the `cloudfoundry/cloudfoundry` provider

#### Scenario: subscription service
- **WHEN** a service task has `resource_type: btp_subaccount_subscription`
- **THEN** the command generates a `btp_subaccount_subscription` resource paired with its `btp_subaccount_entitlement`, both via the BTP provider, with no `location` or `cf_space`

#### Scenario: entitlement-only service
- **WHEN** a service is classified as entitlement-only
- **THEN** the command generates only the entitlement assignment and no instance or subscription resource

#### Scenario: kyma resources
- **WHEN** a configuration unit contains Kyma-provider resources
- **THEN** those resources use the `hashicorp/kubernetes` provider

### Requirement: emit BTP outputs for CF and Kyma provider wiring
When a configuration unit is split for a Cloud Foundry or Kyma environment, the command SHALL emit in the BTP directory's `outputs.tf` the connection values the downstream provider needs, derived from the environment instance labels: the Cloud Foundry API endpoint via `provider::btp::extract_cf_api_url(...)` and the Kyma kubeconfig URL via `provider::btp::extract_kyma_kubeconfig_url(...)`. For Kyma the command SHALL expose only the kubeconfig URL; it SHALL NOT generate the download or parsing of the kubeconfig for the `hashicorp/kubernetes` provider.

#### Scenario: cf api url output
- **WHEN** a unit contains a Cloud Foundry environment
- **THEN** the BTP `outputs.tf` exposes the CF API endpoint using `provider::btp::extract_cf_api_url` against the environment instance labels

#### Scenario: kyma kubeconfig url output
- **WHEN** a unit contains a Kyma environment
- **THEN** the BTP `outputs.tf` exposes the kubeconfig URL using `provider::btp::extract_kyma_kubeconfig_url` against the environment instance labels
- **AND** the command does not generate kubeconfig download or parsing for the kubernetes provider

### Requirement: emit directory ID output when a directory-per-stage layer exists
When `/btp-iac.design` defined a BTP directory-per-stage layer, the command SHALL emit that configuration's `outputs.tf` exposing the directory ID, intended to feed the BTP configuration's `parent_id`.

#### Scenario: directory id output
- **WHEN** a directory-per-stage layer was defined
- **THEN** its `outputs.tf` exposes the directory ID

### Requirement: emit tfvars handover placeholder between directories
Because separate directories are independent Terraform roots on a local backend, the command SHALL move cross-directory values by manual tfvars handover rather than `terraform_remote_state`. For each consuming directory (a `cf/`/`kyma/` directory consuming BTP outputs, or a BTP configuration consuming a directory ID), the command SHALL emit the declaring directory's `outputs.tf` together with a `terraform.tfvars.example` placeholder in the consuming directory that names the variables to copy across. The command SHALL NOT generate `terraform_remote_state` coupling.

#### Scenario: handover scaffolding emitted
- **WHEN** a consuming directory depends on values produced by another directory
- **THEN** the command emits the producing directory's `outputs.tf` and a `terraform.tfvars.example` in the consuming directory naming the variables to copy
- **AND** the command does not generate a `terraform_remote_state` data source

### Requirement: emit provider-initialization tfvars example
For each configuration unit that contains a `provider "btp"` block, the command SHALL emit a `terraform.tfvars.example` file listing the variables referenced by the `provider "btp"` block in that unit as placeholder entries. The file SHALL be named `terraform.tfvars.example` (never `terraform.tfvars`) so Terraform does not load it automatically. If `memory/global-account.md` exists in the project root and contains a non-empty `- Subdomain: <value>` line, the command SHALL use that value as the pre-filled entry for `globalaccount_subdomain`; a blank `- Subdomain:` record or an absent file SHALL produce the sentinel `<your-globalaccount-subdomain>`. When both this requirement and the handover requirement apply to the same directory (e.g. a BTP configuration unit that also consumes a directory ID), the command SHALL emit a single `terraform.tfvars.example` containing the union of all entries from both requirements.

#### Scenario: tfvars example emitted with sentinel placeholder
- **WHEN** a BTP configuration unit is generated and `memory/global-account.md` is absent or its `- Subdomain:` line is blank
- **THEN** the command writes `terraform.tfvars.example` in that unit's directory with one entry per provider-initialization variable, each set to a descriptive sentinel (e.g. `globalaccount_subdomain = "<your-globalaccount-subdomain>"`)

#### Scenario: tfvars example pre-filled from memory
- **WHEN** a BTP configuration unit is generated and `memory/global-account.md` contains a non-empty `- Subdomain: <value>` line
- **THEN** the command writes `terraform.tfvars.example` in that unit's directory with `globalaccount_subdomain` set to that value

#### Scenario: file is never named terraform.tfvars
- **WHEN** any provider-initialization example file is written
- **THEN** the file is named `terraform.tfvars.example`, not `terraform.tfvars`

#### Scenario: merged file when both handover and provider-init apply
- **WHEN** a directory qualifies for both the handover placeholder (it consumes cross-directory outputs such as a `parent_id`) and the provider-initialization example
- **THEN** the command writes exactly one `terraform.tfvars.example` containing the union of all handover variables and all provider-initialization variables

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
The command SHALL look up the current latest version of each required Terraform provider before writing `providers.tf`, and use those versions as `~>` constraints in `required_providers`. The provider set SHALL include, as required by the resolved resource types, each of `SAP/btp`, `cloudfoundry/cloudfoundry`, and `hashicorp/kubernetes`. It SHALL NOT hardcode any version. Before using WebFetch, it SHALL check if the `terraform` MCP server is available and prefer it.

#### Scenario: provider version resolved
- **WHEN** generating `providers.tf`
- **THEN** the command looks up the latest version for each required provider via the terraform MCP server or WebFetch fallback, and uses it as the `~>` constraint

#### Scenario: kyma provider included
- **WHEN** the task set contains a Kyma environment
- **THEN** the resolved provider set includes `hashicorp/kubernetes` at its latest version

### Requirement: look up provider schema via MCP before writing HCL
Before writing any HCL block for a resource or data source, the command SHALL query the terraform MCP server for the exact current schema of that resource type. The lookup sequence is: (1) `search_providers` to obtain the `provider_doc_id` for the resource type's provider, (2) `get_provider_details` to read the exact schema including all attributes, types, and required fields. The command SHALL use the returned schema as the authoritative source for attribute names, types, and required/optional classification. If the terraform MCP server is unavailable, the command SHALL fall back to WebFetch against the Terraform registry and note the fallback.

#### Scenario: schema retrieved from MCP
- **WHEN** the terraform MCP server is available
- **THEN** the command calls `search_providers` then `get_provider_details` for each resource type before writing its HCL, and uses the returned schema as the authoritative attribute set

#### Scenario: schema fallback to WebFetch
- **WHEN** the terraform MCP server is unavailable
- **THEN** the command fetches the provider documentation from the Terraform registry via WebFetch and notes that the MCP server was unavailable

### Requirement: resolve service instance plan via named attributes, not data sources
When generating a `btp_subaccount_service_instance` resource, the command SHALL use the `service_offering_name` and `service_plan_name` attributes directly on the resource. It SHALL NOT generate a `btp_subaccount_service_plan` data source or any other data source to look up a technical plan ID. When generating a `cloudfoundry_service_instance` resource, the command SHALL use the `service_offering_name` and `service_plan_name` attributes directly on the resource. It SHALL NOT generate a `cloudfoundry_service_plan` data source or any other data source to resolve the plan.

#### Scenario: BTP service instance uses named attributes
- **WHEN** a task has `resource_type: btp_subaccount_service_instance`
- **THEN** the generated resource contains `service_offering_name` and `service_plan_name` attributes
- **AND** no `data "btp_subaccount_service_plan"` block is generated for that service instance

#### Scenario: CF service instance uses named attributes
- **WHEN** a task has `resource_type: cloudfoundry_service_instance`
- **THEN** the generated resource contains `service_offering_name` and `service_plan_name` attributes
- **AND** no `data "cloudfoundry_service_plan"` block is generated for that service instance

### Requirement: run terraform init before fmt and validate
The command SHALL run `terraform init` on each generated directory before `terraform fmt` and `terraform validate` for that directory.

#### Scenario: init succeeds
- **WHEN** all files are written and `terraform init` succeeds for a generated directory
- **THEN** the command proceeds to `terraform fmt --recursive` then `terraform validate` for that directory

#### Scenario: init fails
- **WHEN** `terraform init` fails for a generated directory
- **THEN** the command reports the error and does not proceed to fmt or validate for that directory

### Requirement: run terraform fmt and validate on completion
The command SHALL run `terraform fmt --recursive` and `terraform validate` on each generated directory after its `terraform init` and report the outcome per directory. Generation is not complete until `terraform validate` passes on every generated directory. The command SHALL NOT report generation success if `terraform validate` has not passed on all directories.

#### Scenario: fmt and validate pass
- **WHEN** all files are written and both commands succeed for a generated directory
- **THEN** the command reports success for that directory

#### Scenario: fmt or validate fails — fix and retry until passing
- **WHEN** `terraform fmt --recursive` or `terraform validate` fails for a generated directory
- **THEN** the command fixes the reported issues in the affected files, re-runs `terraform fmt --recursive` and `terraform validate`, and repeats until `terraform validate` passes
- **AND** the command does not report generation success until all directories pass

#### Scenario: user cancels during retry loop
- **WHEN** the user explicitly cancels while the command is in the fix-and-retry loop
- **THEN** the command stops and reports the directories that have not yet passed validation

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
