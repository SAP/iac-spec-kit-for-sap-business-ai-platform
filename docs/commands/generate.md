# `btp-iac.generate`

!!! abstract "Summary"
    **Role:** Required — the final step. · **Reads:** `specs/tasks.md` (+ governance) · **Writes:** `terraform/`

Generates complete, validated Terraform HCL by executing each task in dependency order.

## When to run it

Last, after `btp-iac.design` has annotated every task with its target file path.

## Inputs and outputs

| | |
|---|---|
| **Reads** | `specs/tasks.md` (with the file-path annotations from `design`), and `memory/governance.md` if present. |
| **Writes** | Terraform files into `terraform/`, at the paths `design` annotated. Also updates `specs/tasks.md` — each completed task's checkbox is flipped from `- [ ]` to `- [x]`. |

## Behaviour

### Stage filter

At the start of the run, `generate` asks which stage(s) to generate — `Which stage(s) should be generated? (e.g. dev, test, prod — or 'all')`. Only tasks whose stage annotation matches the answer are processed; tasks outside the chosen stage(s) are skipped and remain in `specs/tasks.md` as spec-only, ungenerated entries.

- Runs a full governance validation pass across all six categories **before writing any file**.
- Generates HCL per task in dependency order and writes it to the annotated paths.
- Generates BTP-located service instances with the BTP provider and CF-located instances with `SAP/cloudfoundry`; entitlement-only services generate only their entitlement assignment.
- Runs `terraform init`, then `terraform fmt --recursive`, and `terraform validate` on completion; on a `fmt` or `validate` failure it fixes the failing resource and retries until both pass.
- If `terraform init` fails it reports the error and does not proceed to `fmt` or `validate`.

!!! warning "Generation stops before writing on a governance violation"
    With a governance file present, a violation in any of the six categories stops `generate` **before any file is written**, unless `- Override: true` is set. A metered-service concern is a warning only, but a missing required cost-centre tag stops generation.

!!! note "Generated files are left uncommitted"
    `generate` leaves the generated files as unstaged working-tree changes; it never commits or pushes them unless you explicitly ask.

## Example

```
/btp-iac.generate
```

`generate` validates against governance, then writes the HCL for the HR leave-request project into `terraform/` — `versions.tf`, `subaccount.tf`, `services-btp.tf`, `services-cf.tf`, and `security.tf` — and runs `terraform fmt --recursive` and `terraform validate`. The provider version constraints in `versions.tf` are resolved at runtime (via the `terraform` MCP server, falling back to a `WebFetch` against the Terraform registry) and written as `~>` constraints — never hardcoded. When a `specs/connectivity.md` was produced, it also emits `btp_subaccount_destination_generic` and `btp_subaccount_destination_certificate` resources. When it finishes you can review and apply the result yourself:

```sh
cd terraform
terraform init
terraform plan
```

## Related

- Requires [`btp-iac.design`](design.md).
- For applying the output, see the [usage walkthrough](../walkthrough.md).
