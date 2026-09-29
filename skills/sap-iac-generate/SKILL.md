---
name: sap-iac-generate
description: Generates complete, validated Terraform HCL by executing each task in dependency order.
license: Apache-2.0
metadata:
  author: SAP
  version: "1.6"
---

# BTP IaC — Generate

Generates the complete, validated Terraform HCL by executing each task in dependency order and writing resources to the directory and file paths defined by `/sap-iac.design`.

Reads `specs/tasks.md` (with path annotations and task metadata). Writes each configuration unit using the standard layout — `main.tf` (resources), `variables.tf`, `outputs.tf`, `providers.tf` (`required_providers`), and `backend.tf` (default local backend) — under the `terraform/` directory. Runs `terraform init`, `terraform fmt --recursive`, and `terraform validate` on each generated directory, fixing any issues and retrying until they pass.

## Platform Name Normalization

The following names all refer to the same platform and are semantically equivalent for all natural-language interpretation in this skill:
- **SAP Business Technology Platform** (and "Business Technology Platform")
- **SAP BTP** (and "BTP" used as a product name in prose)
- **SAP Business AI Platform** (and "Business AI Platform")
- **SAP BAIP** (and "BAIP")

Treat any of these aliases as identical when interpreting user intent. This normalization applies only to natural-language prose. Technical identifiers remain untouched: BTP CLI command tokens (`btp list`, `btp target`), Terraform provider names (`btp`, `hashicorp/btp`), resource type prefixes (`btp_subaccount`, `btp_service_instance`), region codes, and API paths.

## BTP platform validation

### BTP operation safety boundary

The prohibition on state-changing CLI commands excludes the permitted target prelude.

When invoking the BTP CLI or any BTP MCP tool, perform **only read or list retrievals**. For the BTP CLI, invoke only documented read/list commands (for example, `btp list ...`); for BTP MCP, invoke only a tool explicitly documented as a read/list lookup. `btp target --global-account <subdomain>` is the sole permitted account-selection prelude and may be used only immediately before those read/list CLI commands. Never invoke, suggest, or approve a BTP operation that creates, updates, deletes, assigns, unassigns, enables, disables, or otherwise mutates BTP state — even when requested by the user. Do not run login, config, profile, or any other state-changing CLI command.

If a live BTP availability check is required while generating, read `<project-root>/.sap-iac/platform-validation.md` and prefer its recorded CLI route, then this agent's recorded BTP MCP route. If no route is recorded, retain user input without blocking. If a recorded route cannot authenticate, target, or complete its lookup, ask the user to resolve it before relying on platform data. Read the preferred infrastructure provider from `memory/governance.md`; when it is not `none`, target the governed global account and use `btp list accounts/available-region` (or the equivalent MCP lookup), ignoring `NEO` entries, to obtain provider metadata for every generated subaccount region. Do not write provider preferences or response-field metadata to `.sap-iac/platform-validation.md`.

---

## Governance Check

**Before writing any Terraform HCL**, locate the sap-iac project root by walking up from the current working directory until a directory containing `specs/`, `memory/`, and `terraform/` is found. Then check whether `<project-root>/memory/governance.md` exists.

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

## Subdomain Uniqueness

For every `btp_subaccount` resource generated, append a random UUID suffix to the subdomain to guarantee uniqueness:

1. Emit one `random_uuid` resource per `btp_subaccount`, using the same resource label (e.g. `random_uuid "subaccount_dev"`), in the same `main.tf`.
2. Read the base subdomain from the task's `subdomain` metadata field (written there by `/sap-iac.tasks` from `specs/landscape.md`). BTP subdomains are limited to 63 characters. The suffix `-<uuid>` is 37 characters (1 hyphen + 36 UUID chars), so the base **must be truncated to at most 26 characters** before appending. Use a `locals` block with `substr` to enforce this; set the `subdomain` attribute to the local:
   ```hcl
   resource "random_uuid" "subaccount_dev" {}

   locals {
     subaccount_dev_subdomain = "${substr("my-subaccount-dev", 0, 26)}-${random_uuid.subaccount_dev.result}"
   }

   resource "btp_subaccount" "subaccount_dev" {
     name      = "..."
     subdomain = local.subaccount_dev_subdomain
     region    = "..."
   }
   ```
   When the base is already 26 characters or fewer, `substr` is a no-op and the value passes through unchanged. The local name follows the pattern `<resource_label>_subdomain`.
3. Add `hashicorp/random` to `required_providers` in `providers.tf` for any configuration unit that contains at least one `btp_subaccount` resource. Look up its latest version at runtime exactly as for other providers — never hardcode it. Omit `hashicorp/random` from units that contain no `btp_subaccount` resources.

---

## CF Environment Landscape Label

For every generated `btp_subaccount_environment_instance` resource where the task metadata field `environment_type` equals `cloudfoundry`, emit the following two blocks **before** the resource block in the same `main.tf`. **Do not apply this pattern to tasks where `environment_type` is `kyma` or any other value.** The `environment_type` field in task metadata is the sole discriminator — never infer it from the resource label or any other heuristic.

### Pattern

```hcl
data "btp_subaccount_environments" "env_info_<label>" {
  subaccount_id = <subaccount_id_reference>
}

resource "terraform_data" "active_env_label_<label>" {
  input = [for env in data.btp_subaccount_environments.env_info_<label>.values : env if env.service_name == "cloudfoundry" && env.environment_type == "cloudfoundry" && env.availability_level == "ACTIVE"][0].landscape_label
}

resource "btp_subaccount_environment_instance" "<label>" {
  subaccount_id    = <subaccount_id_reference>
  landscape_label  = terraform_data.active_env_label_<label>.output
  # ... other attributes
}
```

### Rules

1. **CF only, keyed on task metadata** — apply this pattern only when the task's `environment_type` metadata field equals `cloudfoundry`. For tasks where `environment_type = kyma`, emit `btp_subaccount_environment_instance` directly with no data source or `terraform_data` block and no `landscape_label` attribute. The `environment_type` field is written by `/sap-iac.tasks` and is the authoritative discriminator; do not infer it from the resource label or any other source.
2. **`subaccount_id` is always a reference** — use the same expression (e.g. `btp_subaccount.dev.id`) in both the data source and the environment instance resource. Never hardcode the subaccount ID as a string literal.
3. **Naming convention** — all three blocks share the environment instance's resource label:
   - Data source: `btp_subaccount_environments "env_info_<label>"`
   - `terraform_data`: `terraform_data "active_env_label_<label>"`
   - Resource: `btp_subaccount_environment_instance "<label>"` (unchanged)
4. **`landscape_label` attribute** — always set to `terraform_data.active_env_label_<label>.output`; never hardcode the label string.
5. **Filter criteria** — the `for` expression filters on `service_name == "cloudfoundry"`, `environment_type == "cloudfoundry"`, and `availability_level == "ACTIVE"`. Do not alter these conditions.
6. **Block ordering** — data source first, then `terraform_data`, then the resource. All three go in the same `main.tf`.
7. **`required_version`** — `terraform_data` requires Terraform 1.4+. Whenever at least one CF environment instance is generated in a configuration unit, add `required_version = ">= 1.4"` to the `terraform {}` block in `providers.tf`. Omit it when no CF environment instance is present.

---

## Provider Version

**Before writing `providers.tf`**, look up the latest version of each required Terraform provider. The provider set spans `SAP/btp`, `cloudfoundry/cloudfoundry`, `hashicorp/kubernetes` (Kyma), and `hashicorp/random` (when `btp_subaccount` resources are present) as required by the resolved resource types. Before using `WebFetch` to look up provider versions, check if the `terraform` MCP server is available. If yes, use it. If not, fall back to `WebFetch` against the Terraform registry.

Use the retrieved version as the `~>` constraint in `required_providers` inside `providers.tf`. Never hardcode a version.

---

## Provider Schema Lookup

**Before writing any HCL block** for a resource or data source, query the terraform MCP server to obtain the exact current schema for that resource type. Do not rely on prior knowledge, hardcoded attribute lists, or assumptions about attribute names, types, or required fields.

**Lookup sequence (per distinct resource type):**

1. Call `search_providers` with the resource type's provider name (e.g., `btp`, `cloudfoundry`, `kubernetes`) to obtain the `provider_doc_id`.
2. Call `get_provider_details` with that `provider_doc_id` to read the exact schema: all attributes, their types, and which are required vs. optional.
3. Use the returned schema as the authoritative source when writing the HCL block for that resource type.

Perform this lookup once per **distinct resource type** encountered during a generation run — not once per resource instance. A typical run involves 3–6 distinct types.

**Fallback**: If the terraform MCP server is unavailable, fetch the provider documentation from the Terraform public registry via WebFetch. Note once to the user that the MCP server was unavailable and the fallback was used.

---

## Service instance, subscription, and entitlement provider selection

Each service task carries a `resource_type` set by `/sap-iac.tasks` from the service's `consumption_type`:

- `btp_subaccount_service_instance` — generate with the BTP provider (`btp_subaccount_service_instance`). Use the `service_offering_name` and `service_plan_name` attributes directly on the resource. Do **not** generate a `btp_subaccount_service_plan` data source or any other data source to resolve a technical plan ID. Pair with `btp_subaccount_entitlement` as needed.
- `cloudfoundry_service_instance` — generate with the Cloud Foundry provider (`cloudfoundry_service_instance`) scoped to the `cf_space` recorded on the task. Use the `service_offering_name` and `service_plan_name` attributes directly on the resource. Do **not** generate a `cloudfoundry_service_plan` data source or any other data source to resolve the plan. Use `cloudfoundry/cloudfoundry` as its `required_providers` source.
- `btp_subaccount_subscription` — generate with the BTP provider (`btp_subaccount_subscription`), paired with its `btp_subaccount_entitlement`. No `location` or `cf_space` applies.
- `btp_subaccount_entitlement` (entitlement-only) — generate only the entitlement assignment resource; no instance or subscription resource.

Kyma-provider resources use the `hashicorp/kubernetes` provider. Include whichever providers the resolved resource types require in `providers.tf`'s `required_providers`.

### Service instance parameters

When a `btp_subaccount_service_instance` or `cloudfoundry_service_instance` task in `specs/tasks.md` has a `parameters` field in its task metadata block, emit a `parameters = jsonencode({...})` attribute on the generated resource using those key-value pairs. When the `parameters` field is absent from the task metadata, omit the `parameters` attribute entirely — do not emit an empty `parameters` attribute or a placeholder.

```hcl
# Example: task entry has parameters: { data: { memory: 32, edition: "cloud", generateSystemPassword: true } }
# Task metadata provides no concrete name, so the generated stand-in is lifted to a variable.
resource "btp_subaccount_service_instance" "hana" {
  subaccount_id         = btp_subaccount.dev.id
  service_offering_name = "hana-cloud"
  serviceplan_name      = "hana"
  name                  = var.btp_subaccount_service_instance_hana_name
  parameters = jsonencode({
    data = {
      memory                 = 32
      edition                = "cloud"
      generateSystemPassword = true
    }
  })
}

# Corresponding variables.tf entry:
variable "btp_subaccount_service_instance_hana_name" {
  type        = string
  description = "Name of the hana service instance."
}

# Corresponding terraform.tfvars.example entry:
# btp_subaccount_service_instance_hana_name = "<hana-instance-name>"

# Example: task entry has no parameters block
resource "btp_subaccount_service_instance" "alert_notification" {
  subaccount_id         = btp_subaccount.dev.id
  service_offering_name = "alert-notification"
  serviceplan_name      = "free"
  name                  = var.btp_subaccount_service_instance_alert_notification_name
  # no parameters attribute
}

# Corresponding variables.tf entry:
variable "btp_subaccount_service_instance_alert_notification_name" {
  type        = string
  description = "Name of the alert_notification service instance."
}

# Corresponding terraform.tfvars.example entry:
# btp_subaccount_service_instance_alert_notification_name = "<alert-notification-instance-name>"
```

### Entitlement amount

For a `btp_subaccount_entitlement` task with an explicit `amount` metadata value, emit that value as the resource's `amount` attribute. This applies to the derived `APPLICATION_RUNTIME` / `MEMORY` runtime-memory entitlement; generate only the entitlement resource, never a service instance or subscription.

Otherwise, when a task's metadata contains `quota_required: true` (set by `/sap-iac.tasks` from the plan category recorded by `/sap-iac.services`), add `amount = 1` to every `btp_subaccount_entitlement` resource generated for that task. When neither field is present, omit the `amount` attribute.

```hcl
resource "btp_subaccount_entitlement" "my_service" {
  subaccount_id = btp_subaccount.dev.id
  service_name  = "my-service"
  plan_name     = "standard"
  amount        = 1   # shown for quota_required: true; explicit metadata uses its own value
}
```

For example, runtime-memory task metadata with `service_offering_name = APPLICATION_RUNTIME`, `service_plan_name = MEMORY`, and `amount = 2` MUST generate `amount = 2`.

---

## Custom IdP trust configuration generation (subaccount level)

For tasks with `resource_type = btp_subaccount_trust_configuration`, emit:

```hcl
resource "btp_subaccount_trust_configuration" "<resource-label>" {
  subaccount_id     = <subaccount_id reference>
  identity_provider = "<identity-provider-url>"
  origin            = "<origin>" # omit unless origin_explicit = true
}
```

Rules:
- `subaccount_id` MUST reference the matching subaccount resource.
- `identity_provider` is the URL from task metadata. It is sufficient when no explicit origin was captured.
- Emit `origin` only when task metadata contains both `origin` and `origin_explicit = true`.
- Never derive `origin` from `identity_provider`, and never use a derived origin from another task or metadata field.
- This rule applies only to `btp_subaccount_trust_configuration`. Keep origin handling for every other resource type unchanged.

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

## Role collection assignment generation (subaccount level)

For tasks with `resource_type = btp_subaccount_role_collection_assignment`, emit:

```hcl
resource "btp_subaccount_role_collection_assignment" "<resource-label>" {
  subaccount_id        = <subaccount_id reference>
  role_collection_name = btp_subaccount_role_collection_base.<base-resource-label>.name
  user_name            = "<username>"   # user assignment — omit when group assignment
  group_name           = "<group-name>" # group assignment — omit when user assignment
  origin               = "<origin>"     # omit when absent from task metadata
}
```

Rules:
- `role_collection_name` MUST reference the corresponding `btp_subaccount_role_collection_base` resource via a Terraform expression, not a literal string.
- Emit `user_name` for user assignments; emit `group_name` for group assignments. Never emit both.
- Emit `origin` only when the task metadata contains an `origin` field; omit it otherwise.
- This resource MUST be placed in the same configuration unit as its `btp_subaccount_role_collection_base` dependency. Add `SAP/btp` to `required_providers` (already present for any unit with BTP resources).

---

## Cloud Foundry space and role generation

For tasks with `resource_type = cloudfoundry_space`, emit:

```hcl
resource "cloudfoundry_space" "<resource-label>" {
  name = "<space-name>"
  org  = var.cf_org_id
}
```

Rules:
- `name` is the space name from task metadata.
- `org` MUST use `var.cf_org_id`, the Cloud Foundry organization ID handed over from the corresponding BTP configuration unit.
- This resource uses the `cloudfoundry/cloudfoundry` provider and MUST be placed in the CF configuration unit. Add that provider to `required_providers` in any unit containing this resource.

For tasks with `resource_type = cloudfoundry_space_role`, emit:

```hcl
resource "cloudfoundry_space_role" "<resource-label>" {
  space    = <cf_space_id reference>
  type     = "<role_type>"
  username = "<username>"
  origin   = "<origin>"
}
```

Rules:
- `space` MUST reference the `cloudfoundry_space` resource for the named space via a Terraform expression (e.g. `cloudfoundry_space.dev_space.id`), not a hardcoded ID.
- `type` is the `role_type` value from task metadata (e.g. `space_developer`, `space_auditor`).
- `username` is the `username` from task metadata.
- `origin` is the `origin` from task metadata (either the derived custom-IdP origin or `sap.ids`).
- This resource uses the `cloudfoundry/cloudfoundry` provider. Add `cloudfoundry/cloudfoundry` to `required_providers` in any configuration unit that contains at least one `cloudfoundry_space_role` resource (it is already present when CF service instances or space resources exist in the same unit).
- Each `cloudfoundry_space_role` resource depends on its `cloudfoundry_space` resource; the reference expression creates this dependency implicitly — no `depends_on` is needed.

### CF space role barrier for subsequent resources

For every non-role Cloud Foundry-provider task scoped to a newly created space, inspect its task dependencies. Retain its existing reference to the matching `cloudfoundry_space` resource and emit an explicit `depends_on` listing every dependent `cloudfoundry_space_role` resource for that same space. Resolve role-resource labels from the dependent task IDs and list them in stable task-ID order. For example:

```hcl
resource "cloudfoundry_service_instance" "orders" {
  space = cloudfoundry_space.dev_apps.id
  # ... resource-specific attributes

  depends_on = [
    cloudfoundry_space_role.dev_apps_developer,
    cloudfoundry_space_role.dev_apps_manager,
  ]
}
```

This enforces `cloudfoundry_space -> all cloudfoundry_space_role resources for that space -> other same-space resource` in Terraform's apply graph. Do not infer roles from names or apply a barrier merely because resources share a CF configuration unit: derive it exclusively from the task dependencies. Do not emit these role-based `depends_on` entries for `cloudfoundry_space_role`, BTP-provider resources, CF organization-scoped resources, or resources for a different space.

---

## File layout and cross-directory wiring

Write each configuration unit using the standard layout defined by `/sap-iac.design`: `main.tf`, `variables.tf`, `outputs.tf`, `providers.tf`, `backend.tf`. `required_providers` goes in `providers.tf`; `backend.tf` defaults to a local backend.

### BTP outputs for CF / Kyma

When a unit is split into `btp/` + `cf/`|`kyma/`, emit in the `btp/` `outputs.tf` the connection details the downstream provider needs, derived from the environment instance `labels`:

- Cloud Foundry — the API endpoint via `provider::btp::extract_cf_api_url(<environment-instance>.labels)` **and** the CF Org ID via `provider::btp::extract_cf_org_id(<environment-instance>.labels)`. Both outputs are mandatory whenever a CF environment instance is generated.
- Kyma — the kubeconfig URL via `provider::btp::extract_kyma_kubeconfig_url(<environment-instance>.labels)`. Expose **only the URL**; do not generate the kubeconfig download or parsing, and do not emit provider configuration (kubeconfig) for the kubernetes provider.

### Directory-ID output

When `/sap-iac.design` defined a BTP directory-per-stage layer, emit that configuration's `outputs.tf` exposing the directory ID, intended to feed the BTP configuration's `parent_id`.

### Manual tfvars handover

Separate directories are independent Terraform roots on the local backend, so cross-directory values move by **manual tfvars handover** — never `terraform_remote_state`. For each consuming directory (a `cf/`/`kyma/` directory consuming BTP outputs, or a BTP configuration consuming a directory ID), emit the producing directory's `outputs.tf` and a `terraform.tfvars.example` placeholder in the consuming directory that names the variables to copy across. The user runs the producer, copies the output values into the consumer's tfvars, then runs the consumer.

When the producing BTP directory contains a Cloud Foundry environment, the consuming directory's `terraform.tfvars.example` **must list both `cf_api_url` and `cf_org_id`** as handover variables — never one without the other.

### Provider-initialization tfvars example

For every configuration unit that contains a `provider "btp"` block, emit a `terraform.tfvars.example` in that unit's directory listing each variable referenced by the `provider "btp"` block as a placeholder entry. Always use the file name `terraform.tfvars.example` — never `terraform.tfvars` — so Terraform does not load it automatically.

**Pre-fill from memory**: Before writing the file, check whether `<project-root>/memory/global-account.md` exists. Look for a line matching `- Subdomain: <value>` (the canonical format written by `sap-iac init`). If the value after `- Subdomain:` is non-empty, use it as the value for `globalaccount_subdomain`. If the file is absent or the `- Subdomain:` line is blank, use the sentinel `<your-globalaccount-subdomain>`.

**Merged file when both rules apply**: If a configuration unit qualifies for both this rule and the manual tfvars handover rule (e.g. a BTP configuration consuming a directory `parent_id`), emit **one** `terraform.tfvars.example` containing the union of all handover variables and all provider-initialization variables. Never write the file twice.

Example output (`terraform.tfvars.example` in the BTP configuration unit directory):
```hcl
# Copy this file to terraform.tfvars and fill in the values before running terraform apply.
globalaccount_subdomain = "<your-globalaccount-subdomain>"
```

When pre-filled from memory:
```hcl
# Copy this file to terraform.tfvars and fill in the values before running terraform apply.
globalaccount_subdomain = "my-actual-subdomain"
```

---

## Resource Prohibitions

The following Terraform resource or data source types MUST NEVER appear in any generated file. The positive mapping from intent to resource type is owned by `/sap-iac.tasks`.

| Prohibited resource | Use instead |
|---|---|
| `btp_subaccount_destination` | `btp_subaccount_destination_generic` |
| `btp_subaccount_role_collection` | `btp_subaccount_role_collection_base` + `btp_subaccount_role_collection_role` |
| `data "btp_subaccount_service_plan"` | `service_offering_name` + `service_plan_name` attributes on `btp_subaccount_service_instance` directly |
| `data "cloudfoundry_service_plan"` | `service_offering_name` + `service_plan_name` attributes on `cloudfoundry_service_instance` directly |

**Guard**: Before generating HCL for any task, check its `resource_type`. If it is `btp_subaccount_destination` or `btp_subaccount_role_collection`, **STOP** and report:
> "Task <ID> carries a prohibited resource type `<type>`. Re-run `/sap-iac.tasks` to correct the mapping before generating."

Before generating any data source block, check whether it appears in the prohibited table above. If it does, **STOP** and report:
> "Data source `<type>` is prohibited. Use the direct attribute approach shown in the table."

---

## Provider Authentication

**Ask once per run, after the stage-filter question and before writing any file.** The questions are independent — BTP auth and CF auth are separate prompts. The BTP prompt is only shown when the stage-filtered task set contains at least one BTP resource. The CF prompt is only shown when the stage-filtered task set contains at least one Cloud Foundry resource.

### BTP Provider Authentication

Ask the user: "How should BTP provider authentication be performed?"

| Method | Variables emitted in `variables.tf` |
|---|---|
| username/password | `login_name`, `password` |
| username/password with custom IdP | `login_name`, `password`, `idp` |
| SSO / token | `idp` |
| mTLS (client certificate) | `x509_private_key`, `x509_cert_chain` |

Rules:
- **No `default` value** on `login_name`, `password`, `idp`, `x509_private_key`, or `x509_cert_chain`.
- All credential variables (`login_name`, `password`, `idp`, `x509_private_key`, `x509_cert_chain`) **must** have `sensitive = true`.
- Declare all auth variables in `variables.tf` for **every** BTP configuration unit.
- The `provider "btp"` block in `providers.tf` references **only** the variables required by the selected method — do not emit unused auth attributes.
- Add placeholder entries for all auth variables to `terraform.tfvars.example` (merged with existing entries per the provider-initialization tfvars example rule).
- Apply the same selection to all BTP configuration units — do not re-prompt per unit.

Example `variables.tf` for username/password:
```hcl
variable "globalaccount_subdomain" {
  type        = string
  description = "Subdomain of the global account."
}

variable "login_name" {
  type        = string
  description = "BTP login name (email address)."
  sensitive   = true
}

variable "password" {
  type        = string
  description = "BTP password."
  sensitive   = true
}
```

Example `provider "btp"` block for username/password:
```hcl
provider "btp" {
  globalaccount = var.globalaccount_subdomain
  username      = var.login_name
  password      = var.password
}
```

Example `provider "btp"` block for SSO/token:
```hcl
provider "btp" {
  globalaccount = var.globalaccount_subdomain
  idp           = var.idp
}
```

Example `provider "btp"` block for mTLS:
```hcl
provider "btp" {
  globalaccount    = var.globalaccount_subdomain
  x509_private_key = var.x509_private_key
  x509_cert_chain  = var.x509_cert_chain
}
```

### CF Provider Authentication

Ask the user (only when CF resources are present): "How should Cloud Foundry provider authentication be performed?"

| Method | Variables emitted in `variables.tf` |
|---|---|
| username/password | `cf_user`, `cf_password` |
| username/password with custom origin | `cf_user`, `cf_password`, `cf_origin` |
| SSO / token | `cf_sso_passcode` |

Rules:
- **No `default` value** on any CF auth variable.
- All CF credential variables (`cf_user`, `cf_password`, `cf_origin`, `cf_sso_passcode`) **must** have `sensitive = true`.
- Declare all CF auth variables in `variables.tf` for **every** CF configuration unit.
- The `provider "cloudfoundry"` block in `providers.tf` references **only** the variables required by the selected method.
- Add placeholder entries for all CF auth variables to `terraform.tfvars.example`.
- Apply the same selection to all CF configuration units — do not re-prompt per unit.
- Skip this prompt entirely when the task set contains no Cloud Foundry resources.

Example `provider "cloudfoundry"` block for username/password:
```hcl
provider "cloudfoundry" {
  api_url  = var.cf_api_url
  user     = var.cf_user
  password = var.cf_password
}
```

Example `provider "cloudfoundry"` block for SSO/token:
```hcl
provider "cloudfoundry" {
  api_url      = var.cf_api_url
  sso_passcode = var.cf_sso_passcode
}
```

---

## Placeholder Variables

**Any placeholder value** that would appear as a string literal in generated HCL must instead be expressed as a Terraform `variable` in `variables.tf` and listed as a placeholder entry in `terraform.tfvars.example`. Never emit a raw placeholder string directly inside a resource, data source, or locals block in any `.tf` file.

### What counts as a placeholder

A value is a placeholder when it:
- Matches the pattern `<something>` (angle-bracket sentinel, e.g. `<collection-name>`, `<role-template-app-id>`, `<your-value>`)
- Is a bare keyword: `TODO`, `FIXME`, `TBD`, `CHANGEME`
- Starts with the prefix `my-` or `my_` and no concrete value was supplied in the task metadata for that attribute (e.g. `"my-hana"`, `"my-subaccount"`)

A value is **not** a placeholder when it was explicitly provided in the task metadata — even if it happens to look generic.

### Detection pass

Before writing each `.tf` file, scan every string literal that is about to be emitted. For each placeholder found:

1. **Derive a variable name** using the format `<resource_type>_<resource_label>_<attribute_name>` — where `<resource_type>` and `<resource_label>` are the first and second strings in `resource "type" "label"`, and `<attribute_name>` is the HCL attribute key, lowercased with hyphens replaced by underscores. This produces a unique, predictable name for every resource attribute in a configuration unit. Examples: resource type `"btp_subaccount_role_collection_base"`, label `"dev_admins"`, attribute `name` → `btp_subaccount_role_collection_base_dev_admins_name`; resource type `"btp_subaccount_service_instance"`, label `"hana"`, attribute `name` → `btp_subaccount_service_instance_hana_name`; resource type `"btp_subaccount_role"`, label `"dev_admins"`, attribute `app_id` → `btp_subaccount_role_dev_admins_app_id`.
2. **Declare the variable** in the configuration unit's `variables.tf`:
   ```hcl
   # resource "btp_subaccount_role" "developer" { app_id = "<role-template-app-id>" }
   # → variable name: btp_subaccount_role_developer_app_id
   variable "btp_subaccount_role_developer_app_id" {
     type        = string
     description = "Application ID of the developer role template."
   }
   ```
   - No `default` value — the user must supply it.
   - `sensitive = true` only for values that are credentials or secrets; omit it for names, descriptions, and identifiers.
3. **Replace the placeholder** in the `.tf` file with `var.<variable_name>`.
4. **Add an entry** to `terraform.tfvars.example` (merged per the existing merge rule):
   ```hcl
   btp_subaccount_role_developer_app_id = "<role-template-app-id>"
   ```
   Preserve the original placeholder text as the example value so the user knows what to fill in.

### Scope

Apply this rule to **all generated `.tf` files** in all configuration units — `main.tf`, `variables.tf`, `outputs.tf`, `providers.tf`. The auth-variable and globalaccount rules already satisfy this requirement for their respective attributes; do not duplicate those variables.

### Example

Before (placeholder in `main.tf`):
```hcl
resource "btp_subaccount_role_collection_base" "dev_admins" {
  subaccount_id = btp_subaccount.dev.id
  name          = "<collection-name>"
  description   = "<optional description>"
}
```

After (variable reference in `main.tf`):
```hcl
resource "btp_subaccount_role_collection_base" "dev_admins" {
  subaccount_id = btp_subaccount.dev.id
  name          = var.btp_subaccount_role_collection_base_dev_admins_name
  description   = var.btp_subaccount_role_collection_base_dev_admins_description
}
```

Corresponding `variables.tf` additions:
```hcl
variable "btp_subaccount_role_collection_base_dev_admins_name" {
  type        = string
  description = "Name of the dev_admins role collection."
}

variable "btp_subaccount_role_collection_base_dev_admins_description" {
  type        = string
  description = "Description of the dev_admins role collection."
}
```

Corresponding `terraform.tfvars.example` additions:
```hcl
btp_subaccount_role_collection_base_dev_admins_name        = "<collection-name>"
btp_subaccount_role_collection_base_dev_admins_description = "<optional description>"
```

---

## Workflow

Read `specs/tasks.md` to get the dependency-ordered task list with file path annotations from `/sap-iac.design`.

**Stage filter**: Ask the user: "Which stage(s) should be generated? (e.g. dev, test, prod — or 'all')" Only process tasks whose stage annotation matches the answer. Tasks outside the requested stages are skipped — they remain in `specs/tasks.md` as spec-only and are not generated.

For each selected `btp_subaccount` task, read `usage` and `beta_enabled` from its Task metadata. Emit every value that is present and valid; do not infer defaults or substitute governance values. Legacy tasks may omit either value, in which case omit that Terraform attribute and note the task ID as using legacy classification metadata. Stop and report the task ID only when a present value is invalid. Tasks outside the selected stages are not checked for these attributes.

For each task in dependency order:
1. Generate the Terraform HCL resource(s) for that task
2. Write to the file path annotated by `/sap-iac.design`
3. Mark the task as complete in `specs/tasks.md` by changing its checkbox from `- [ ]` to `- [x]`
4. Continue to the next task

After all tasks are complete, for **each generated directory** (each independent Terraform root — e.g. `btp/`, `cf/`, `kyma/`, per-stage or directory-per-stage directories):
1. Run `terraform init` on the directory
2. Run `terraform fmt --recursive` on the directory
3. Run `terraform validate` on the directory
4. If `terraform fmt` or `terraform validate` fails: fix the reported issues in the affected files, then re-run `terraform fmt --recursive` and `terraform validate`. Repeat until both pass — there is no retry limit. **Generation is not complete and success MUST NOT be reported until `terraform validate` passes on every generated directory.** The only exit from this loop (other than all directories passing) is explicit user cancellation. If the user cancels, report which directories have not yet passed validation. If the same directory fails three consecutive times with the same error, pause and ask the user how to proceed before retrying further.
5. Report the final outcome per directory

---

## Git Safeguards

**Do not commit** the generated Terraform files unless the user explicitly asks (e.g. "commit", "git commit", "commit the changes").

**Do not push** the generated Terraform files unless the user explicitly asks (e.g. "push", "git push").

Default behaviour after a successful generate run is to leave the files as unstaged changes in the working tree so the user can review, iterate, and decide when to commit.

## Next step

Your Terraform code is in `terraform/`. Review the generated files, then commit and apply when ready.
