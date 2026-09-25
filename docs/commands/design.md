# `sap-iac.design`

!!! abstract "Summary"
    **Role:** Required. · **Reads:** `specs/tasks.md` · **Writes:** `specs/tasks.md` (in place)

Translates the task list into a concrete Terraform folder structure and annotates each task with the file it will be written to.

## When to run it

After `sap-iac.tasks`, and before `sap-iac.generate`.

## Inputs and outputs

| | |
|---|---|
| **Reads** | `specs/tasks.md`. |
| **Writes** | `specs/tasks.md`, updated in place. It does not create a separate output file. |

## Behaviour

- Emits a standard file layout per configuration unit: `main.tf` (resources), `variables.tf` (input variables), `outputs.tf` (output values), `providers.tf` (provider config and `required_providers`), and `backend.tf` (defaulting to a **local** backend).
- Chooses how stages are modelled by reading `memory/governance.md`; when governance records no choice, it prompts the user. Two modes: **per-stage directories** (one directory per stage, each with the standard layout) or **single configuration** (one directory handling all stages via stage variables in `variables.tf`, with values in `<stage>.tfvars` files).
- When a unit contains a Cloud Foundry or Kyma environment, splits it into a `btp/` subdirectory (all BTP-provider resources) plus a sibling `cf/` or `kyma/` subdirectory. In per-stage-directory mode the split sits inside each stage directory. A BTP-only unit is not split.
- Keeps BTP-provider and Cloud Foundry/Kyma-provider resources apart by each task's `location`, so generation can use the correct provider per directory.
- Only when governance/tasks indicate BTP **directories** model stages, defines a directory-per-stage layer whose `outputs.tf` exposes the directory ID, fed into the BTP configuration's `parent_id`.
- Annotates each task with its target directory and file path, which `generate` then follows.
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
