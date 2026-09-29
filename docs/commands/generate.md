# `sap-iac.generate`

!!! abstract "Summary"
    **Role:** Required — the final step. · **Reads:** `specs/tasks.md` (+ governance) · **Writes:** `terraform/`

Generates complete, validated Terraform HCL by executing each task in dependency order.

## When to run it

Last, after `sap-iac.design` has annotated every task with its target file path.

## Inputs and outputs

| | |
|---|---|
| **Reads** | `specs/tasks.md` (with the file-path annotations from `design`), and `memory/governance.md` if present. |
| **Writes** | Terraform files into `terraform/`, at the paths `design` annotated. Also updates `specs/tasks.md` — each completed task's checkbox is flipped from `- [ ]` to `- [x]`. |

## Behaviour

### Stage filter

At the start of the run, `generate` asks which stage(s) to generate — `Which stage(s) should be generated? (e.g. dev, test, prod — or 'all')`. Only tasks whose stage annotation matches the answer are processed; tasks outside the chosen stage(s) are skipped and remain in `specs/tasks.md` as spec-only, ungenerated entries.

- Runs a full governance validation pass across all six categories **before writing any file**.
- Generates HCL per task in dependency order and writes each configuration unit with the standard layout — `main.tf`, `variables.tf`, `outputs.tf`, `providers.tf` (with `required_providers`), and `backend.tf` (default local backend).
- Generates BTP-located service instances with the BTP provider, CF-located instances with `cloudfoundry/cloudfoundry`, and Kyma resources with `hashicorp/kubernetes`; entitlement-only services generate only their entitlement assignment.
- Emits collected service-instance parameters as `parameters = jsonencode(...)`. It omits the attribute when the task has no parameters.
- Emits an entitlement's explicit `amount`, including the calculated `APPLICATION_RUNTIME` / `MEMORY` amount. Otherwise, `quota_required: true` produces `amount = 1`; tasks with neither field omit `amount`.
- Generates subaccount-level security resources using the split approach: `btp_subaccount_role_collection_base` for each role collection definition and `btp_subaccount_role_collection_role` for each individual role assignment. Never emits `btp_subaccount_role_collection`.
- Generates `btp_subaccount_trust_configuration` with its `identity_provider`. It emits `origin` only when task metadata contains an explicitly captured origin and `origin_explicit = true`; it never derives an origin from the URL.
- For every non-role CF resource scoped to a newly created space, retains its space reference and emits an explicit `depends_on` for every recorded role assignment in that same space. This preserves `space -> roles -> resource` in Terraform's apply graph without affecting organization-scoped, BTP, or other-space resources.
- For a CF/Kyma split, emits the `btp/` `outputs.tf` with the CF API URL (`provider::btp::extract_cf_api_url`) or Kyma kubeconfig URL (`provider::btp::extract_kyma_kubeconfig_url`, URL only), plus a `terraform.tfvars.example` in the consuming `cf/`/`kyma/` directory for the manual handover. When a directory-per-stage layer exists, its `outputs.tf` exposes the directory ID for the BTP config's `parent_id`. No `terraform_remote_state` coupling is generated.
- Emits a `terraform.tfvars.example` in every BTP configuration unit listing the provider-initialization variables (e.g. `globalaccount_subdomain`). If `memory/global-account.md` contains a non-empty `- Subdomain: <value>` line, that value is pre-filled; otherwise a sentinel placeholder is used. When the same directory also receives handover variables (e.g. a `parent_id` from a directory-per-stage layer), a single merged file is written.
- Runs `terraform init`, then `terraform fmt --recursive`, and `terraform validate` on **each generated directory**; on a `fmt` or `validate` failure it fixes the failing resource and retries until both pass.
- If `terraform init` fails it reports the error and does not proceed to `fmt` or `validate`.

!!! warning "Generation stops before writing on a governance violation"
    With a governance file present, a violation in any of the six categories stops `generate` **before any file is written**, unless `- Override: true` is set. A metered-service concern is a warning only, but a missing required cost-centre tag stops generation.

!!! note "Generated files are left uncommitted"
    `generate` leaves the generated files as unstaged working-tree changes; it never commits or pushes them unless you explicitly ask.

## Example

```
/sap-iac.generate
```

`generate` validates against governance, then writes the HCL for the HR leave-request project into `terraform/` using the standard per-unit layout — `main.tf`, `variables.tf`, `outputs.tf`, `providers.tf`, and `backend.tf` (default local backend) — and runs `terraform init`, `terraform fmt --recursive`, and `terraform validate` on each generated directory. The provider version constraints in `providers.tf` are resolved at runtime (via the `terraform` MCP server, falling back to a `WebFetch` against the Terraform registry) and written as `~>` constraints — never hardcoded. Security resources are emitted as `btp_subaccount_role_collection_base`, `btp_subaccount_role_collection_role`, and `btp_subaccount_trust_configuration` blocks; the latter omits `origin` unless it was explicitly captured. When a `specs/connectivity.md` was produced, it also emits `btp_subaccount_destination_generic` and `btp_subaccount_destination_certificate` resources. When it finishes you can review and apply the result yourself:

```sh
cd terraform/btp
terraform init
terraform plan
```

## Related

- Requires [`sap-iac.design`](design.md).
- For applying the output, see the [usage walkthrough](../walkthrough.md).
