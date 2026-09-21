---
name: btp-iac-generate
description: Generates complete, validated Terraform HCL by executing each task in dependency order.
license: Apache-2.0
metadata:
  author: SAP
  version: "1.0"
---

# BTP IaC — Generate

Generates the complete, validated Terraform HCL by executing each task in dependency order and writing resources to the file paths defined by `/btp-iac.design`.

Reads `specs/tasks.md` (with file path annotations). Writes Terraform files to the `terraform/` directory. Runs `terraform init`, `terraform fmt --recursive`, and `terraform validate` on completion, fixing any issues and retrying until both pass.

## BTP platform validation

If a live BTP availability check is required while generating, read `<project-root>/.btp-iac/platform-validation.md` and prefer its recorded CLI route, then this agent's recorded BTP MCP route. If no route is recorded, retain user input without blocking. If a recorded route cannot authenticate, target, or complete its lookup, ask the user to resolve it before relying on platform data.

---

## Governance Check

**Before writing any Terraform HCL**, locate the btp-iac project root by walking up from the current working directory until a directory containing `specs/`, `memory/`, and `terraform/` is found. Then check whether `<project-root>/memory/governance.md` exists.

**If it does not exist:** proceed without constraints and note: "No governance rules found — proceeding without enforcement."

**If it exists**, load all rules and run a full pre-generation validation pass across all six categories. Check every resource in `specs/tasks.md` against the governance rules before writing a single file.

### Region validation
For each subaccount resource:
- Region must be in `## Regions → Allowed` and not in `Forbidden`
- **STOP** on violation with region, rule violated, and fix instructions

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

**Before writing `versions.tf`**, look up the latest version of each required Terraform provider. Before using `WebFetch` to look up provider versions, check if the `terraform` MCP server is available. If yes, use it. If not, fall back to `WebFetch` against the Terraform registry.

Use the retrieved version as the `~>` constraint in `required_providers`. Never hardcode a version.

---

## Workflow

Read `specs/tasks.md` to get the dependency-ordered task list with file path annotations from `/btp-iac.design`.

For each task in dependency order:
1. Generate the Terraform HCL resource(s) for that task
2. Write to the file path annotated by `/btp-iac.design`
3. Continue to the next task

After all tasks are complete:
1. Run `terraform init` on the `terraform/` directory
2. Run `terraform fmt --recursive` on the `terraform/` directory
3. Run `terraform validate` on the `terraform/` directory
4. If `terraform fmt` or `terraform validate` fails: fix the reported issues in the affected files, then re-run `terraform fmt --recursive` and `terraform validate` until both pass
5. Report the final outcome

---

## Git Safeguards

**Do not commit** the generated Terraform files unless the user explicitly asks (e.g. "commit", "git commit", "commit the changes").

**Do not push** the generated Terraform files unless the user explicitly asks (e.g. "push", "git push").

Default behaviour after a successful generate run is to leave the files as unstaged changes in the working tree so the user can review, iterate, and decide when to commit.
