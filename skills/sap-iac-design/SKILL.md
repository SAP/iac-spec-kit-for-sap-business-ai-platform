---
name: sap-iac-design
description: Translates the task list into a concrete Terraform folder structure and annotates each task with its target directory and file path.
license: Apache-2.0
metadata:
  author: SAP
  version: "1.3"
---

# BTP IaC — Design

Translates the task list into a concrete Terraform folder structure and annotates each task in `specs/tasks.md` with the directory and file path it will be written to, preserving each task's metadata unchanged.

Reads `specs/tasks.md`. Performs no BTP mutations — see the safety boundary below.

## Platform Name Normalization

The following names all refer to the same platform and are semantically equivalent for all natural-language interpretation in this skill:
- **SAP Business Technology Platform** (and "Business Technology Platform")
- **SAP BTP** (and "BTP" used as a product name in prose)
- **SAP Business AI Platform** (and "Business AI Platform")
- **SAP BAIP** (and "BAIP")

Treat any of these aliases as identical when interpreting user intent. This normalization applies only to natural-language prose. Technical identifiers remain untouched: BTP CLI command tokens (`btp list`, `btp target`), Terraform provider names (`btp`, `hashicorp/btp`), resource type prefixes (`btp_subaccount`, `btp_service_instance`), region codes, and API paths.

## Standard file layout

Every configuration unit (a directory Terraform is run in) uses the standard layout:

- `main.tf` — resources
- `variables.tf` — input variables
- `outputs.tf` — output values
- `providers.tf` — provider configuration and the `required_providers` block
- `backend.tf` — backend configuration, defaulting to a **local** backend

Do not reintroduce ad-hoc resource-type files. `required_providers` lives in `providers.tf`; there is no separate `versions.tf`.

## Stage modelling

Determine how stages are modelled by reading `memory/governance.md`. If governance records a stage-modelling choice, use it. If governance is absent or records no choice, **prompt the user** to pick one of:

- **Per-stage directories** — one directory per stage, named after the stage (`terraform/<stage>/…`), each containing the standard layout.
- **Single configuration** — one directory whose configuration handles all stages via stage-specific variables declared in `variables.tf`, with values supplied through per-stage tfvars files named after the stage (e.g. `<stage>.tfvars`). These stage variables propagate into the configuration to satisfy naming conventions.

## BTP / CF / Kyma split

When a configuration unit contains a Cloud Foundry or Kyma environment, split it into subdirectories: all BTP-provider resources under `btp/`, and all Cloud Foundry- or Kyma-provider resources under a sibling `cf/` or `kyma/`. Drive service-instance placement by `location` (`btp` or `cf`). Subscription tasks (`resource_type = btp_subaccount_subscription`) and entitlement-only tasks (`resource_type = btp_subaccount_entitlement`) carry no `location` but are always BTP-provider resources — place them under `btp/`.

When no CF or Kyma environment is present, do not split: use the standard layout directly in the unit directory.

Per-stage-directory mode, CF present:

```
terraform/
  dev/
    btp/     { main.tf variables.tf outputs.tf providers.tf backend.tf terraform.tfvars.example }
    cf/      { main.tf variables.tf outputs.tf providers.tf backend.tf terraform.tfvars.example }
  prod/
    btp/ …
    cf/  …
```

Single-configuration mode, CF present:

```
terraform/
  btp/   { main.tf variables.tf outputs.tf providers.tf backend.tf <stage>.tfvars terraform.tfvars.example }
  cf/    { main.tf variables.tf outputs.tf providers.tf backend.tf <stage>.tfvars terraform.tfvars.example }
```

BTP-only (no CF/Kyma): the standard five files plus `terraform.tfvars.example` directly in the unit directory, no `btp/`/`cf/`/`kyma/` split.

The `btp/` `outputs.tf` carries the connection details the downstream provider needs (CF API URL, Kyma kubeconfig URL); the consuming `cf/`/`kyma/` directory receives them by manual tfvars handover (`/sap-iac.generate` emits a `terraform.tfvars.example` there). Directories are independent Terraform roots on the local backend — do not couple them with `terraform_remote_state`.

## BTP directory-per-stage layer

Only when governance or tasks indicate BTP **directories** are used to model stages, define a dedicated directory-per-stage configuration using the same standard layout. Its `outputs.tf` exposes the directory ID, which is handed (via manual tfvars) into the BTP configuration's `parent_id`. When directories are not used for stages, do not define this layer — subaccounts sit directly under the global account.

## Task annotations

Annotate every task in `specs/tasks.md` with the directory and file path it will be written to under the selected structure, while preserving its existing Task metadata unchanged.

## BTP platform validation

### BTP operation safety boundary

The prohibition on state-changing CLI commands excludes the permitted target prelude.

When invoking the BTP CLI or any BTP MCP tool, perform **only read or list retrievals**. For the BTP CLI, invoke only documented read/list commands (for example, `btp list ...`); for BTP MCP, invoke only a tool explicitly documented as a read/list lookup. `btp target --global-account <subdomain>` is the sole permitted account-selection prelude and may be used only immediately before those read/list CLI commands. Never invoke, suggest, or approve a BTP operation that creates, updates, deletes, assigns, unassigns, enables, disables, or otherwise mutates BTP state — even when requested by the user. Do not run login, config, profile, or any other state-changing CLI command.

If this workflow needs a live BTP availability check, read `<project-root>/.sap-iac/platform-validation.md` and prefer its recorded CLI route, then this agent's recorded BTP MCP route. Do not block when no route is recorded; retain user input. If a recorded route fails to authenticate, target, or answer, ask the user to resolve it before relying on the result.

## Next step

Next: `/sap-iac.generate` — write and validate all Terraform HCL.
