---
name: btp-iac-generate
description: Generates complete, validated Terraform HCL by executing each task in dependency order.
license: Apache-2.0
metadata:
  author: SAP
  version: "1.3"
---

# BTP IaC — Generate

Generates the complete, validated Terraform HCL by executing each task in dependency order and writing resources to the directory and file paths defined by `/btp-iac.design`.

Reads `specs/tasks.md` (with path annotations and task metadata). Writes each configuration unit using the standard layout — `main.tf` (resources), `variables.tf`, `outputs.tf`, `providers.tf` (`required_providers`), and `backend.tf` (default local backend) — under the `terraform/` directory. Runs `terraform init`, `terraform fmt --recursive`, and `terraform validate` on each generated directory, fixing any issues and retrying until they pass.

## BTP platform validation

### BTP operation safety boundary

The prohibition on state-changing CLI commands excludes the permitted target prelude.

When invoking the BTP CLI or any BTP MCP tool, perform **only read or list retrievals**. For the BTP CLI, invoke only documented read/list commands (for example, `btp list ...`); for BTP MCP, invoke only a tool explicitly documented as a read/list lookup. `btp target --global-account <subdomain>` is the sole permitted account-selection prelude and may be used only immediately before those read/list CLI commands. Never invoke, suggest, or approve a BTP operation that creates, updates, deletes, assigns, unassigns, enables, disables, or otherwise mutates BTP state — even when requested by the user. Do not run login, config, profile, or any other state-changing CLI command.

If a live BTP availability check is required while generating, read `<project-root>/.btp-iac/platform-validation.md` and prefer its recorded CLI route, then this agent's recorded BTP MCP route. If no route is recorded, retain user input without blocking. If a recorded route cannot authenticate, target, or complete its lookup, ask the user to resolve it before relying on platform data. Read the preferred infrastructure provider from `memory/governance.md`; when it is not `none`, target the governed global account and use `btp list accounts/available-region` (or the equivalent MCP lookup), ignoring `NEO` entries, to obtain provider metadata for every generated subaccount region. Do not write provider preferences or response-field metadata to `.btp-iac/platform-validation.md`.

---

## Governance Check

**Before writing any Terraform HCL**, locate the btp-iac project root by walking up from the current working directory until a directory containing `specs/`, `memory/`, and `terraform/` is found. Then check whether `<project-root>/memory/governance.md` exists.

**If it does not exist:** proceed without constraints and note: "No governance rules found — proceeding without enforcement."

**If it exists**, load all rules and run a full pre-generation validation pass across all six categories. Check every resource in `specs/tasks.md` against the governance rules before writing a single file.

### Region validation
For each subaccount resource:
- Region must be in `## Regions → Allowed` and not in `Forbidden`
- **STOP** on violation with region, rule violated, and fix instructions
- When `## Regions → Preferred infrastructure provider` is not `none`, compare it with unambiguous provider metadata from the targeted region lookup. Never infer a provider from a region code. A known mismatch logs a warning naming the region, preferred provider, and returned provider, but does not stop generation; unavailable or unmappable metadata is advisory only. A missing field in a legacy governance file means `none`.

### Naming validation
For each subaccount resource:
- Name must match `## Naming → Subaccount pattern`
- Environment tier must be in `## Naming → Environments`
- **STOP** on violation with name, expected pattern, and fix instructions

### Account environment validation
For each Cloud Foundry or Kyma environment resource:
- Its type must be permitted by `## Account Setup → Allowed environments`
- Its name must match the applicable Cloud Foundry organization or Kyma environment pattern
- Every Cloud Foundry space name must match `## Naming → Cloud Foundry space pattern`
- **STOP** on violation with the selected type or name, the expected rule, and fix instructions

### Service plan validation
For each service instance resource:
- Plan must be in `## Service Plans → <tier> → Permitted` and not in `Forbidden`
- **STOP** on violation with plan, tier, permitted plans, and fix instructions

### Security validation
For each trust configuration:
- Custom IdP presence matches `## Security → Custom IdP` requirement
- Required role collection assignments are included
- **STOP** on violation with specific rule and fix instructions

### Cost controls validation
For each metered service resource:
- If `## Cost Controls → Metered service warning: enabled`, output a warning listing the metered services before proceeding (do not block)
- If `## Cost Controls → Cost centre tag: required`, verify each subaccount resource includes the cost centre tag attribute — **STOP** if missing
  > "GOVERNANCE VIOLATION: Cost centre tag is required on all subaccounts. Add the tag to `<subaccount>` or add `- Override: true` to memory/governance.md."

**If `- Override: true` is set in `<project-root>/memory/governance.md`:** log a warning for each violation and continue instead of stopping.

---

## Provider Version

**Before writing `providers.tf`**, look up the latest version of each required Terraform provider. The provider set spans `SAP/btp`, `cloudfoundry/cloudfoundry`, and `hashicorp/kubernetes` (Kyma) as required by the resolved resource types. Before using `WebFetch` to look up provider versions, check if the `terraform` MCP server is available. If yes, use it. If not, fall back to `WebFetch` against the Terraform registry.

Use the retrieved version as the `~>` constraint in `required_providers` inside `providers.tf`. Never hardcode a version.

---

## Service instance, subscription, and entitlement provider selection

Each service task carries a `resource_type` set by `/btp-iac.tasks` from the service's `consumption_type`:

- `btp_subaccount_service_instance` — generate with the BTP provider (`btp_subaccount_service_instance`, using `btp_subaccount_entitlement` / `btp_subaccount_service_plan` as needed).
- `cloudfoundry_service_instance` — generate with the Cloud Foundry provider (`cloudfoundry_service_instance`) scoped to the `cf_space` recorded on the task, resolving the offering/plan via CF data sources. Use `cloudfoundry/cloudfoundry` as its `required_providers` source.
- `btp_subaccount_subscription` — generate with the BTP provider (`btp_subaccount_subscription`), paired with its `btp_subaccount_entitlement`. No `location` or `cf_space` applies.
- `btp_subaccount_entitlement` (entitlement-only) — generate only the entitlement assignment resource; no instance or subscription resource.

Kyma-provider resources use the `hashicorp/kubernetes` provider. Include whichever providers the resolved resource types require in `providers.tf`'s `required_providers`.

---

## Role collection and role generation (subaccount level)

For tasks with `resource_type = btp_subaccount_role_collection_base`, emit:

```hcl
resource "btp_subaccount_role_collection_base" "<resource-label>" {
  subaccount_id = <subaccount_id reference>
  name          = "<collection-name>"
  description   = "<optional description>"   # omit when not specified
}
```

For tasks with `resource_type = btp_subaccount_role_collection_role`, emit:

```hcl
resource "btp_subaccount_role_collection_role" "<resource-label>" {
  subaccount_id        = <subaccount_id reference>
  name                 = btp_subaccount_role_collection_base.<base-resource-label>.name
  role_name            = "<role-name>"
  role_template_name   = "<role-template-name>"
  role_template_app_id = "<role-template-app-id>"
}
```

The `name` attribute MUST reference the corresponding `btp_subaccount_role_collection_base` resource via a Terraform expression (`btp_subaccount_role_collection_base.<base-resource-label>.name`), not a literal string. This creates an implicit graph dependency so Terraform will not attempt to create the role before the collection exists. Do NOT emit `btp_subaccount_role_collection` in any generated file.

---

## File layout and cross-directory wiring

Write each configuration unit using the standard layout defined by `/btp-iac.design`: `main.tf`, `variables.tf`, `outputs.tf`, `providers.tf`, `backend.tf`. `required_providers` goes in `providers.tf`; `backend.tf` defaults to a local backend.

### BTP outputs for CF / Kyma

When a unit is split into `btp/` + `cf/`|`kyma/`, emit in the `btp/` `outputs.tf` the connection details the downstream provider needs, derived from the environment instance `labels`:

- Cloud Foundry — the API endpoint via `provider::btp::extract_cf_api_url(<environment-instance>.labels)`.
- Kyma — the kubeconfig URL via `provider::btp::extract_kyma_kubeconfig_url(<environment-instance>.labels)`. Expose **only the URL**; do not generate the kubeconfig download or parsing, and do not emit provider configuration (kubeconfig) for the kubernetes provider.

### Directory-ID output

When `/btp-iac.design` defined a BTP directory-per-stage layer, emit that configuration's `outputs.tf` exposing the directory ID, intended to feed the BTP configuration's `parent_id`.

### Manual tfvars handover

Separate directories are independent Terraform roots on the local backend, so cross-directory values move by **manual tfvars handover** — never `terraform_remote_state`. For each consuming directory (a `cf/`/`kyma/` directory consuming BTP outputs, or a BTP configuration consuming a directory ID), emit the producing directory's `outputs.tf` and a `terraform.tfvars.example` placeholder in the consuming directory that names the variables to copy across. The user runs the producer, copies the output values into the consumer's tfvars, then runs the consumer.

---

## Resource Prohibitions (subaccount level)

The following Terraform resource types MUST NEVER appear in any generated file. The positive mapping from intent to resource type is owned by `/btp-iac.tasks`.

| Prohibited resource | Use instead |
|---|---|
| `btp_subaccount_destination` | `btp_subaccount_destination_generic` |
| `btp_subaccount_role_collection` | `btp_subaccount_role_collection_base` + `btp_subaccount_role_collection_role` |

**Guard**: Before generating HCL for any task, check its `resource_type`. If it is `btp_subaccount_destination` or `btp_subaccount_role_collection`, **STOP** and report:
> "Task <ID> carries a prohibited resource type `<type>`. Re-run `/btp-iac.tasks` to correct the mapping before generating."

Do not attempt to substitute or remap — stop and require the user to fix the task list.

---

## Workflow

Read `specs/tasks.md` to get the dependency-ordered task list with file path annotations from `/btp-iac.design`.

**Stage filter**: Ask the user: "Which stage(s) should be generated? (e.g. dev, test, prod — or 'all')" Only process tasks whose stage annotation matches the answer. Tasks outside the requested stages are skipped — they remain in `specs/tasks.md` as spec-only and are not generated.

For each selected `btp_subaccount` task, read `usage` and `beta_enabled` from its Task metadata. Emit every value that is present and valid; do not infer defaults or substitute governance values. Legacy tasks may omit either value, in which case omit that Terraform attribute and note the task ID as using legacy classification metadata. Stop and report the task ID only when a present value is invalid. Tasks outside the selected stages are not checked for these attributes.

For each task in dependency order:
1. Generate the Terraform HCL resource(s) for that task
2. Write to the file path annotated by `/btp-iac.design`
3. Mark the task as complete in `specs/tasks.md` by changing its checkbox from `- [ ]` to `- [x]`
4. Continue to the next task

After all tasks are complete, for **each generated directory** (each independent Terraform root — e.g. `btp/`, `cf/`, `kyma/`, per-stage or directory-per-stage directories):
1. Run `terraform init` on the directory
2. Run `terraform fmt --recursive` on the directory
3. Run `terraform validate` on the directory
4. If `terraform fmt` or `terraform validate` fails: fix the reported issues in the affected files, then re-run `terraform fmt --recursive` and `terraform validate` until both pass
5. Report the final outcome per directory

---

## Git Safeguards

**Do not commit** the generated Terraform files unless the user explicitly asks (e.g. "commit", "git commit", "commit the changes").

**Do not push** the generated Terraform files unless the user explicitly asks (e.g. "push", "git push").

Default behaviour after a successful generate run is to leave the files as unstaged changes in the working tree so the user can review, iterate, and decide when to commit.

## Next step

Your Terraform code is in `terraform/`. Review the generated files, then commit and apply when ready.
