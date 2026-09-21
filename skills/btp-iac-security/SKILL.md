---
name: btp-iac-security
description: Identifies authentication and authorisation requirements — IdP trust, role collections, and role collection assignments.
license: Apache-2.0
metadata:
  author: SAP
  version: "1.0"
---

# BTP IaC — Security

Identifies the authentication and authorisation requirements — role collection definitions, user assignments, IdP trust configurations, and role collection assignments.

Reads `specs/scenario.md` and `specs/landscape.md`. Produces `specs/trust.md`.

## Tool preferences

Before using `WebFetch` to look up SAP documentation, check if the `sap-docs` MCP server is available (tools prefixed `mcp__sap-docs__*`). If yes, use it. If not, fall back to `WebFetch`.

## BTP platform validation

### BTP operation safety boundary

The prohibition on state-changing CLI commands excludes the permitted target prelude.

When invoking the BTP CLI or any BTP MCP tool, perform **only read or list retrievals**. For the BTP CLI, invoke only documented read/list commands (for example, `btp list ...`); for BTP MCP, invoke only a tool explicitly documented as a read/list lookup. `btp target --global-account <subdomain>` is the sole permitted account-selection prelude and may be used only immediately before those read/list CLI commands. Never invoke, suggest, or approve a BTP operation that creates, updates, deletes, assigns, unassigns, enables, disables, or otherwise mutates BTP state — even when requested by the user. Do not run login, config, profile, or any other state-changing CLI command.

If this workflow needs a live BTP availability check, read `<project-root>/.btp-iac/platform-validation.md` and prefer its recorded CLI route, then this agent's recorded BTP MCP route. If no route is recorded, retain user input without blocking. If a recorded route cannot authenticate, target, or complete its lookup, ask the user to resolve it before relying on platform data.

---

## Governance Check

**Before defining any security configuration**, locate the btp-iac project root by walking up from the current working directory until a directory containing `specs/`, `memory/`, and `terraform/` is found. Then check whether `<project-root>/memory/governance.md` exists.

**If it does not exist:** proceed without constraints and note: "No governance rules found — proceeding without enforcement."

**If it exists**, load all rules and validate every security decision against the `## Security` section.

### IdP validation
- If `Custom IdP: required` and no custom IdP is being configured: **STOP**
  > "GOVERNANCE VIOLATION: Governance requires a custom Identity Provider. Configure a custom IdP trust or add `- Override: true` to memory/governance.md."

### Role collection validation
- If `Default role collections` specifies assignments and the configuration does not include them: **STOP**
  > "GOVERNANCE VIOLATION: Governance requires default role collection assignments (`<required>`). Include these assignments or add `- Override: true` to memory/governance.md."

**If `- Override: true` is set in `<project-root>/memory/governance.md`:** log a warning for each violation and continue instead of stopping.

---

## Workflow

Read `specs/scenario.md` and `specs/landscape.md` to understand authentication flows, user roles, and external system integrations.

Define for each subaccount:
- IdP trust configurations (platform and application)
- Role collections with role template assignments
- User and group assignments to role collections

Write `specs/trust.md` with the complete security configuration. This file is the direct input to `/btp-iac.tasks`.
