# `sap-iac.design`

!!! abstract "Summary"
    **Role:** Required. · **Reads:** `specs/tasks.md` (+ `specs/state-architecture.md`, governance) · **Writes:** `specs/state-architecture.md`, `specs/tasks.md` (in place)

Guides the Terraform state-chunking architecture, persists it to `specs/state-architecture.md`, and annotates each task with the file it will be written to.

## When to run it

After `sap-iac.tasks`, and before `sap-iac.generate`.

## Inputs and outputs

| | |
|---|---|
| **Reads** | `specs/tasks.md`; `specs/state-architecture.md` and `memory/governance.md` when present. |
| **Writes** | `specs/state-architecture.md` (the approved state-chunking decision) and `specs/tasks.md`, updated in place. |

## Behaviour

- Guides the user through the **state-chunking architecture** first: presents the mandatory BTP / CF / Kyma provider split as fixed, recommends optional finer-grained boundaries (global account & directory, subaccount, entitlements/instances/subscriptions, CF org & space, team/app), and states each boundary's trade-offs (deployment independence, ownership, blast radius, lifecycle, operational overhead).
- Lets the user **accept, refine, or override** the recommendation in free text, and respects that structure for the rest of the run. An override that conflicts with the fixed BTP/CF/Kyma split is flagged for the user to resolve.
- **Persists** the approved architecture to `specs/state-architecture.md` — stage-modelling mode, per-chunk boundary/root/backend/ownership/dependencies, stage ordering, rationale, and assumptions. On a re-run it summarizes the existing file, lets the user revise it, and rewrites it.
- Emits a standard file layout per configuration unit: `main.tf` (resources), `variables.tf` (input variables), `outputs.tf` (output values), `providers.tf` (provider config and `required_providers`), and `backend.tf` (defaulting to a **local** backend).
- Chooses how stages are modelled by reading `memory/governance.md`; when governance records no choice, it prompts the user. Two modes: **per-stage directories** (one directory per stage, each with the standard layout) or **single configuration** (one directory handling all stages via stage variables in `variables.tf`, with values in `<stage>.tfvars` files).
- When a unit contains a Cloud Foundry or Kyma environment, splits it into a `btp/` subdirectory (all BTP-provider resources) plus a sibling `cf/` or `kyma/` subdirectory. In per-stage-directory mode the split sits inside each stage directory. A BTP-only unit is not split.
- Keeps BTP-provider and Cloud Foundry/Kyma-provider resources apart by each task's `location`, so generation can use the correct provider per directory.
- Places entitlement-only tasks, including `APPLICATION_RUNTIME` / `MEMORY`, in the BTP configuration and preserves their amount and other task metadata.
- Only when governance/tasks indicate BTP **directories** model stages, defines a directory-per-stage layer whose `outputs.tf` exposes the directory ID, fed into the BTP configuration's `parent_id`.
- Annotates each task with its target directory and file path, preserving all metadata and dependencies. A CF space, its role assignments, and subsequent resources scoped to that space share the CF `main.tf`, retaining the `space -> roles -> resource` barrier without coupling separate spaces.
- Performs no BTP mutations for its core function — it only reads and rewrites `specs/tasks.md`, plus `memory/governance.md` for the stage-mode choice.

## Example

```
/sap-iac.design
```

`design` picks the stage-modelling mode (from governance or by prompting) and annotates each task in `specs/tasks.md` with its target directory and file. For a per-stage-directory layout with a Cloud Foundry environment:

```markdown
1. [T1] Create subaccount hr-leave-prod          -> prod/btp/main.tf
2. [T2] Create xsuaa (application, BTP)          -> prod/btp/main.tf
3. [T3] Create hana-cloud (hana, CF)             -> prod/cf/main.tf
4. [T4] Create role collection HR_Leave_Employee -> prod/btp/main.tf
```

Each directory carries the standard `main.tf` / `variables.tf` / `outputs.tf` / `providers.tf` / `backend.tf` layout. The `prod/btp/outputs.tf` exposes the CF API URL that `prod/cf/` consumes via a manual tfvars handover.

## Related

- Requires [`sap-iac.tasks`](tasks.md).
- Feeds [`sap-iac.generate`](generate.md).
