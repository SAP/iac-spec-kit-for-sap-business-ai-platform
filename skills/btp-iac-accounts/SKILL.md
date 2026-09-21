---
name: btp-iac-accounts
description: Defines the BTP account topology — subaccounts, regions, and directory groupings.
license: Apache-2.0
metadata:
  author: SAP
  version: "1.0"
---

# BTP IaC — Accounts

Defines the BTP account topology — subaccounts, regions, and directory groupings.

Reads `specs/scenario.md` and (if present) `memory/governance.md`. Produces `specs/landscape.md` as the authoritative account structure that all service and trust configuration will reference.

## Tool preferences

Before using `WebFetch` to look up SAP documentation, check if the `sap-docs` MCP server is available (tools prefixed `mcp__sap-docs__*`). If yes, use it. If not, fall back to `WebFetch`.

## BTP platform validation

Read `<project-root>/.btp-iac/platform-validation.md` after locating the project root. For regions and any explicitly named service offering, subscription, or service-plan pair in the topology inputs, use the recorded BTP CLI first or this agent's recorded BTP MCP route second. Target the global-account subdomain from governance or the landscape decision using `btp target --global-account <subdomain>` before CLI checks. Use `btp list accounts/entitlement` for entitlement/subscription checks and `btp list accounts/available-region` for regions; ignore all `NEO` region entries. When governance selects a preferred infrastructure provider other than `none`, use returned region provider metadata to check it. On the first successful lookup that exposes this metadata, record its exact field name as `- Region provider field: <field>` in `.btp-iac/platform-validation.md`. Map only unambiguous provider labels to the governed values and never infer a provider from the region code; report a mismatch as a warning, and report unavailable or unmappable provider metadata as advisory only.

If no recorded live route exists, retain user input without blocking. If a live route cannot authenticate, target, or complete a lookup, ask the user to resolve it. A completed unavailable result requires a replacement before output is written. With no live region route, consult the SAP Help Cloud Foundry region list as advisory only and retain user input.

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
- If `## Regions → Preferred infrastructure provider` is present and not `none`, compare it with the provider metadata from the platform region lookup. A known mismatch logs a warning naming the region, preferred provider, and returned provider, but does not stop. A missing field in a legacy governance file means `none`; missing or unmappable metadata is advisory only.

### Naming validation

For each subaccount name you are about to generate:
- If the name does not match the pattern in `## Naming → Subaccount pattern`: **STOP**
  > "GOVERNANCE VIOLATION: Subaccount name `<name>` does not match the required pattern `<pattern>`. Adjust the name or add `- Override: true` to memory/governance.md."
- If the environment tier used is not in `## Naming → Environments`: **STOP**
  > "GOVERNANCE VIOLATION: Environment tier `<tier>` is not defined in governance (defined: `<environments>`). Add the tier to governance or use a defined one."

### Account environment validation

For each account environment you are about to define:
- Its type must be listed in `## Account Setup → Allowed environments`.
- A Cloud Foundry organization name must match `## Naming → Cloud Foundry org pattern`; a Kyma environment name must match `## Naming → Kyma environment pattern`; and every Cloud Foundry space name must match `## Naming → Cloud Foundry space pattern`.
- If a required pattern or allowed-environment rule is absent from an older governance file, collect the missing decision from the user instead of inventing one.
- **STOP** on a governed violation and identify the selected type or name, the expected rule, and the `- Override: true` escape hatch.

**If `- Override: true` is set in `<project-root>/memory/governance.md`:** log a warning for each violation and continue instead of stopping.

---

## Workflow

Read `specs/scenario.md` to understand the application's environment and deployment requirements.

Define the account topology:
- Global account subdomain. Use `## Account Setup → Global account subdomain` when present. Otherwise ask for it before writing the landscape. It must be the global account's subdomain, not its GUID; if the supplied value is a GUID, ask for the subdomain instead.
- Directory groupings (if applicable)
- Subaccounts per environment tier — name, region, description, subdomain
- Per subaccount, determine whether to create no runtime environment, Cloud Foundry, Kyma, or both. Respect any decision already supplied in the scenario or governance. If it is not known, ask the user.
- For each selected Cloud Foundry environment, collect its organization name. Ask whether Cloud Foundry spaces should be created; if yes, collect their concrete names and validate them against governance when a space pattern exists.
- For each selected Kyma environment, collect its environment name.

Write `specs/landscape.md` with the complete account structure, including the global account subdomain and each subaccount's runtime environment types, names, and Cloud Foundry spaces. This file is the authoritative input for `/btp-iac.services`, `/btp-iac.security`, and `/btp-iac.generate`.
