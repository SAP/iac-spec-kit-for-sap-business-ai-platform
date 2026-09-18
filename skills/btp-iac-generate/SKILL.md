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

Reads `specs/tasks.md` (with file path annotations). Writes Terraform files to the `terraform/` directory. Runs `terraform validate` and `terraform fmt` on completion.

---

## Governance Check

**Before writing any Terraform HCL**, locate the btp-iac project root by walking up from the current working directory until a directory containing `specs/`, `memory/`, and `terraform/` is found. Then check whether `<project-root>/memory/governance.md` exists.

**If it does not exist:** proceed without constraints and note: "No governance rules found — proceeding without enforcement."

**If it exists**, load all rules and run a full pre-generation validation pass across all five categories. Check every resource in `specs/tasks.md` against the governance rules before writing a single file.

### Region validation
For each subaccount resource:
- Region must be in `## Regions → Allowed` and not in `Forbidden`
- **STOP** on violation with region, rule violated, and fix instructions

### Naming validation
For each subaccount resource:
- Name must match `## Naming → Subaccount pattern`
- Environment tier must be in `## Naming → Environments`
- **STOP** on violation with name, expected pattern, and fix instructions

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

## Workflow

Read `specs/tasks.md` to get the dependency-ordered task list with file path annotations from `/btp-iac.design`.

For each task in dependency order:
1. Generate the Terraform HCL resource(s) for that task
2. Write to the file path annotated by `/btp-iac.design`
3. Continue to the next task

After all tasks are complete:
- Run `terraform fmt` on the `terraform/` directory
- Run `terraform validate` on the `terraform/` directory
- Report the outcome. If validate fails, identify the failing resource and the likely cause.

---

## Git Safeguards

**Do not commit** the generated Terraform files unless the user explicitly asks (e.g. "commit", "git commit", "commit the changes").

**Do not push** the generated Terraform files unless the user explicitly asks (e.g. "push", "git push").

Default behaviour after a successful generate run is to leave the files as unstaged changes in the working tree so the user can review, iterate, and decide when to commit.
