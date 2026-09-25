---
name: sap-iac-next
description: Inspects local project state and recommends the next command in the BTP IaC greenfield flow.
license: Apache-2.0
metadata:
  author: SAP
  version: "1.1"
---

# BTP IaC — Next

Inspects the local project state and recommends the next command in the BTP IaC greenfield flow. Does not invoke any BTP CLI command or BTP MCP tool — all decisions are based solely on local file existence.

## Flow overview

The greenfield flow in dependency order:

```
/sap-iac.govern → /sap-iac.scenario → /sap-iac.accounts → /sap-iac.services
  → /sap-iac.security → [/sap-iac.connectivity] → /sap-iac.tasks
  → /sap-iac.design → /sap-iac.generate
```

`/sap-iac.connectivity` is optional. All other steps are required.

## Workflow

### Step 1 — Locate project root

Walk up from the current working directory until a directory containing `specs/`, `memory/`, and `terraform/` subdirectories is found. All checks below are relative to that root.

If no project root is found, report: "Could not locate a sap-iac project root. Run this command from within a project created by `sap-iac init`." and stop.

### Step 2 — Check file existence

Check for the existence of each of the following, in order:

| # | File / path | Produced by |
|---|---|---|
| 1 | `memory/governance.md` | `/sap-iac.govern` |
| 2 | `specs/scenario.md` | `/sap-iac.scenario` |
| 3 | `specs/landscape.md` | `/sap-iac.accounts` |
| 4 | `specs/services.md` | `/sap-iac.services` |
| 5 | `specs/trust.md` | `/sap-iac.security` |
| 6 | `specs/connectivity.md` | `/sap-iac.connectivity` (optional) |
| 7 | `specs/tasks.md` | `/sap-iac.tasks` |
| 8 | `terraform/**/*.tf` (any file) | `/sap-iac.generate` |

### Step 3 — Determine recommendation

Apply the following decision table (first matching rule wins):

| Condition | Recommendation |
|---|---|
| `memory/governance.md` missing | **Run `/sap-iac.govern`** — establish guardrails before anything else. |
| `specs/scenario.md` missing | **Run `/sap-iac.scenario`** — describe your application. |
| `specs/landscape.md` missing | **Run `/sap-iac.accounts`** — map your scenario to BTP directories and subaccounts. |
| `specs/services.md` missing | **Run `/sap-iac.services`** — resolve which BTP services each subaccount needs. |
| `specs/trust.md` missing | **Run `/sap-iac.security`** — set up IdP trust, roles, and role collection assignments. |
| `specs/tasks.md` missing and `specs/connectivity.md` missing | **Run `/sap-iac.connectivity`** (optional) — define destinations and certificates, or **`/sap-iac.tasks`** to skip connectivity. |
| `specs/tasks.md` missing and `specs/connectivity.md` exists | **Run `/sap-iac.tasks`** — build a dependency-ordered execution plan. |
| `terraform/**/*.tf` missing | **Run `/sap-iac.design`** — plan the Terraform file and module layout, then `/sap-iac.generate`. |
| `terraform/**/*.tf` exists | **Done** — Terraform code is in `terraform/`. Review, commit, and apply. |

### Step 4 — Output

Print a state summary followed by the recommendation:

```
BTP IaC — Project State
───────────────────────
✓ memory/governance.md
✓ specs/scenario.md
✓ specs/landscape.md
✗ specs/services.md
✗ specs/trust.md
✗ specs/connectivity.md
✗ specs/tasks.md
✗ terraform/**/*.tf

Next: /sap-iac.services — resolve which BTP services each subaccount needs.
```

Use `✓` for files that exist and `✗` for files that do not. Always print all eight rows so the user can see their full position in the flow.
