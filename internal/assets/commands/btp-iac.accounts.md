# /btp-iac.accounts

Defines the BTP account topology — subaccounts, regions, and directory groupings.

Reads `specs/scenario.md` and (if present) `memory/governance.md`. Produces `specs/landscape.md` as the authoritative account structure that all service and trust configuration will reference.

---

## Governance Check

**Before defining any account topology**, locate the btp-iac project root by walking up from the current working directory until a directory containing `specs/`, `memory/`, and `terraform/` is found. Then check whether `<project-root>/memory/governance.md` exists.

**If it does not exist:** proceed without constraints and note: "No governance rules found — proceeding without enforcement."

**If it exists**, load all rules and validate every decision below against them. Use this enforcement logic:

### Region validation
For each subaccount region you are about to assign:
- If the region is **not** in the `## Regions → Allowed` list: **STOP**
  > "GOVERNANCE VIOLATION: Region `<region>` is not in the approved list (`<allowed>`). Change the region or add `- Override: true` to memory/governance.md to bypass enforcement."
- If the region appears in the `## Regions → Forbidden` list: **STOP** with the same message.

### Naming validation
For each subaccount name you are about to generate:
- If the name does not match the pattern in `## Naming → Subaccount pattern`: **STOP**
  > "GOVERNANCE VIOLATION: Subaccount name `<name>` does not match the required pattern `<pattern>`. Adjust the name or add `- Override: true` to memory/governance.md."
- If the environment tier used is not in `## Naming → Environments`: **STOP**
  > "GOVERNANCE VIOLATION: Environment tier `<tier>` is not defined in governance (defined: `<environments>`). Add the tier to governance or use a defined one."

**If `- Override: true` is set in `<project-root>/memory/governance.md`:** log a warning for each violation and continue instead of stopping.

---

## Workflow

Read `specs/scenario.md` to understand the application's environment and deployment requirements.

Define the account topology:
- Global account reference
- Directory groupings (if applicable)
- Subaccounts per environment tier — name, region, description, subdomain

Write `specs/landscape.md` with the complete account structure. This file is the authoritative input for `/btp-iac.services`, `/btp-iac.security`, and `/btp-iac.generate`.
