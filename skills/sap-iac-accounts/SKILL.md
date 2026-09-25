---
name: sap-iac-accounts
description: Defines the BTP account topology — subaccounts, regions, and directory groupings.
license: Apache-2.0
metadata:
  author: SAP
  version: "1.1"
---

# BTP IaC — Accounts

Defines the BTP account topology — subaccounts, regions, and directory groupings.

Reads `specs/scenario.md`, `memory/global-account.md`, and (if present) `memory/governance.md`. Produces `specs/landscape.md` as the authoritative account structure that all service and trust configuration will reference.

## Tool preferences

Before using `WebFetch` to look up SAP documentation, check if the `sap-docs` MCP server is available (tools prefixed `mcp__sap-docs__*`). If yes, use it. If not, fall back to `WebFetch`.

## BTP platform validation

### BTP operation safety boundary

The prohibition on state-changing CLI commands excludes the permitted target prelude.

When invoking the BTP CLI or any BTP MCP tool, perform **only read or list retrievals**. For the BTP CLI, invoke only documented read/list commands (for example, `btp list ...`); for BTP MCP, invoke only a tool explicitly documented as a read/list lookup. `btp target --global-account <subdomain>` is the sole permitted account-selection prelude and may be used only immediately before those read/list CLI commands. Never invoke, suggest, or approve a BTP operation that creates, updates, deletes, assigns, unassigns, enables, disables, or otherwise mutates BTP state — even when requested by the user. Do not run login, config, profile, or any other state-changing CLI command.

Read `<project-root>/.sap-iac/platform-validation.md`, `<project-root>/memory/global-account.md`, and `<project-root>/memory/governance.md` after locating the project root. For regions and any explicitly named service offering, subscription, or service-plan pair in the topology inputs, use the recorded BTP CLI first or this agent's recorded BTP MCP route second. When `memory/global-account.md` contains a subdomain, target it using `btp target --global-account <subdomain>` before CLI checks. Use `btp list accounts/entitlement` for entitlement/subscription checks and `btp list accounts/available-region` for regions; ignore all `NEO` region entries. Read the preferred infrastructure provider from `memory/governance.md`; when it is not `none`, use returned region provider metadata to check it. Map only unambiguous provider labels to the governed values and never infer a provider from a region code; report a mismatch as a warning, and report unavailable or unmappable provider metadata as advisory only. Do not write provider preferences or response-field metadata to `.sap-iac/platform-validation.md`.

If no recorded live route exists, retain user input without blocking. If a live route cannot authenticate, target, or complete a lookup, ask the user to resolve it. A completed unavailable result requires a replacement before output is written. With no live region route, consult the SAP Help Cloud Foundry region list as advisory only and retain user input.

---

## Governance Check

**Before defining any account topology**, locate the sap-iac project root by walking up from the current working directory until a directory containing `specs/`, `memory/`, and `terraform/` is found. Then check whether `<project-root>/memory/governance.md` exists.

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

## Subaccount usage and beta inference

For every subaccount in the topology, determine two BTP platform attributes:

- `usage` — `USED_FOR_PRODUCTION` or `NOT_USED_FOR_PRODUCTION`
- `beta_enabled` — `true` or `false`

### Governance classifications (read first)

Before applying inference, check `memory/governance.md` for the following optional mapping under `## Account Setup`:

- `- Tier classifications:` — exact tier names, each with explicit `usage` and `beta_enabled` values. For example:
  ```markdown
  - Tier classifications:
    - integration: usage = NOT_USED_FOR_PRODUCTION, beta_enabled = false
  ```

Match a classification entry to a tier name exactly after case normalization. It overrides keyword inference. If the same normalized tier is listed more than once, stop and ask the user to correct `memory/governance.md`; do not choose one entry. This is a governance-data error, not a policy violation: `- Override: true` does not bypass it. Absent entries mean inference applies.

### Inference rules (when no governance override matches)

Apply case-insensitive, hyphen-delimited token matching to the tier name. A token is bounded by the start/end of the name or `-`; do not match arbitrary substrings. If multiple rows match, apply them in table order.

| Tier matches | `usage` | `beta_enabled` |
|---|---|---|
| `(?:^|-)prod(?:uction)?(?:-|$)` | `USED_FOR_PRODUCTION` | `false` |
| `(?:^|-)dev(?:elopment)?(?:-|$)` | `NOT_USED_FOR_PRODUCTION` | `true` |
| `(?:^|-)(?:test|staging|qa)(?:-|$)` | `NOT_USED_FOR_PRODUCTION` | `true` |
| _(no match)_ | — ask user — | — ask user — |

### Unmatched tiers — clarification (ask before the confirmation table)

If one or more tier names do not match any governance override or inference rule, **stop and ask the user in a single question** listing all unmatched tiers:

> "The following tier names could not be automatically classified. For each, please specify `usage` (USED_FOR_PRODUCTION / NOT_USED_FOR_PRODUCTION) and `beta_enabled` (true / false):"
> - `<tier-name-1>`
> - `<tier-name-2>`

Do not show the confirmation table until the user has answered for every unmatched tier.

### Confirmation table

After inference and any clarification, present a single summary table before writing `specs/landscape.md`:

```
Subaccount usage and beta settings (confirmed):

  Subaccount            Tier      Usage                       Beta
  ────────────────────  ────────  ──────────────────────────  ─────
  <subaccount-name>     <tier>    <USED/NOT_USED>             <true/false>
  ...

Accept these, or type changes (e.g. "myapp-test: beta=false")?
```

The user may accept all rows or correct individual rows inline. Apply any corrections before writing.

---

## Workflow

Read `specs/scenario.md` and `memory/global-account.md` to understand the application's environment, deployment requirements, and configured global account.

Define the account topology:
- Global account subdomain. Use the value in `memory/global-account.md`. If it is blank or the file is absent, leave it unspecified and do not ask for one.
- Directory groupings (if applicable)
- Subaccounts per environment tier — name, region, description, subdomain
- Per subaccount, determine whether to create no runtime environment, Cloud Foundry, Kyma, or both. Respect any decision already supplied in the scenario or governance. If it is not known, ask the user.
- For each selected Cloud Foundry environment, collect its organization name. Ask whether Cloud Foundry spaces should be created; if yes, collect their concrete names and validate them against governance when a space pattern exists.
- For each selected Kyma environment, collect its environment name.

Write `specs/landscape.md` with the complete account structure, including the global account subdomain and each subaccount's runtime environment types, names, and Cloud Foundry spaces. For each subaccount, also record the confirmed `usage` (`USED_FOR_PRODUCTION` or `NOT_USED_FOR_PRODUCTION`) and `beta_enabled` (`true` or `false`) values. This file is the authoritative input for `/sap-iac.services`, `/sap-iac.security`, and `/sap-iac.generate`.

## Next step

Next: `/sap-iac.services` — resolve which BTP services each subaccount needs.
