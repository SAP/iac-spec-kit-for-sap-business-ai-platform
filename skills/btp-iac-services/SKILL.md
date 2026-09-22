---
name: btp-iac-services
description: Resolves infrastructure requirements into a dependency-ordered list of BTP entitlements, subscriptions, and service instances.
license: Apache-2.0
metadata:
  author: SAP
  version: "1.0"
---

# BTP IaC — Services

Resolves the infrastructure requirements from `specs/scenario.md` into a dependency-ordered list of BTP entitlements, subscriptions, and service instances per subaccount.

Reads `specs/scenario.md` and `specs/landscape.md`. Produces `specs/services.md` for the team to review and adjust before any code is generated.

## Tool preferences

Before using `WebFetch` to look up SAP documentation, check if the `sap-docs` MCP server is available (tools prefixed `mcp__sap-docs__*`). If yes, use it. If not, fall back to `WebFetch`.

## BTP platform validation

### BTP operation safety boundary

The prohibition on state-changing CLI commands excludes the permitted target prelude.

When invoking the BTP CLI or any BTP MCP tool, perform **only read or list retrievals**. For the BTP CLI, invoke only documented read/list commands (for example, `btp list ...`); for BTP MCP, invoke only a tool explicitly documented as a read/list lookup. `btp target --global-account <subdomain>` is the sole permitted account-selection prelude and may be used only immediately before those read/list CLI commands. Never invoke, suggest, or approve a BTP operation that creates, updates, deletes, assigns, unassigns, enables, disables, or otherwise mutates BTP state — even when requested by the user. Do not run login, config, profile, or any other state-changing CLI command.

Read `<project-root>/.btp-iac/platform-validation.md` and `<project-root>/memory/global-account.md` after locating the project root. Before writing `specs/services.md`, validate every resolved entitlement, subscription, service offering, and plan against the configured global-account subdomain. Prefer the recorded CLI route: when the memory record has a subdomain, run `btp target --global-account <subdomain>` and inspect `btp list accounts/entitlement`. If CLI was unavailable at initialization, use an equivalent scoped BTP MCP entitlement/subscription operation only when this agent is recorded as having BTP MCP support. Do not ask for a subdomain.

If no route is recorded, retain user input without blocking. If a recorded route cannot authenticate, target the account, or complete the lookup, ask the user to resolve it. If the lookup completes and an item is unavailable, require a valid replacement before writing output.

---

## Governance Check

**Before resolving any service dependencies**, locate the btp-iac project root by walking up from the current working directory until a directory containing `specs/`, `memory/`, and `terraform/` is found. Then check whether `<project-root>/memory/governance.md` exists.

**If it does not exist:** proceed without constraints and note: "No governance rules found — proceeding without enforcement."

**If it exists**, load all rules and validate every service plan decision against the `## Service Plans` section.

### Service plan validation
For each service instance you are about to define, identify its environment tier from `specs/landscape.md` and check the plan:
- If the plan is **not** in the `Permitted` list for that tier: **STOP**
  > "GOVERNANCE VIOLATION: Plan `<plan>` is not permitted for environment `<tier>` (permitted: `<permitted-plans>`). Change the plan or add `- Override: true` to memory/governance.md."
- If the plan appears in the `Forbidden` list for that tier: **STOP** with the same message.

**If `- Override: true` is set in `<project-root>/memory/governance.md`:** log a warning for each violation and continue instead of stopping.

---

## Service Type Classification

**Before writing `specs/services.md`**, determine how each service will be consumed. This drives both the output structure and which Terraform provider downstream commands use.

### Step 1 — Check governance for pre-stated decisions

Read `<project-root>/memory/governance.md` if it exists. Look for any section or note that states:
- Which services are entitlement-only
- Which services require a subscription vs. a service instance
- Whether service instances are created on BTP or in a Cloud Foundry space

If any of these decisions are already recorded there, use them without asking the user again. Note which decisions were sourced from governance.

### Step 2 — Classify each service

For every service identified from `specs/scenario.md`, determine its `consumption_type`, one of: `instance` (service instance), `subscription` (app subscription), or `entitlement-only` (entitlement assigned, nothing created).

The **default** is:
- SaaS applications → `subscription`
- Technical services → `instance`

Ask the user to confirm or override the type for **each** service (one question per service; do not batch), presenting the default so a confirmation is a single keystroke:

> "For `<service-name>`: (1) service instance, (2) app subscription, or (3) entitlement only? Default is (`<default>`)."

Skip the question only for a service whose type is already fixed by governance (Step 1). This guarantees the user can always mark any service — including a clearly technical one — as entitlement-only, as issue #62 requires.

If a service is classified as `entitlement-only`, record `consumption_type: entitlement-only` — no instance or subscription is created; the entitlement is assigned to the subaccount for future manual use.

### Step 3 — Determine service instance location

For every service with `consumption_type: instance`, determine where the instance is created:

- **BTP** (`location: btp`) — created directly in the subaccount using the BTP Terraform provider (`SAP/btp`)
- **Cloud Foundry** (`location: cf`) — created inside a specific CF space using the Cloud Foundry Terraform provider (`cloudfoundry-community/cloudfoundry` or `SAP/cloudfoundry`)

Check `specs/landscape.md` for the Cloud Foundry environment and its spaces in the relevant subaccount:

- If **no** Cloud Foundry environment exists for the subaccount, record `location: btp` without asking.
- If a Cloud Foundry environment exists, ask:
  > "Service instance `<service-name>` in subaccount `<subaccount>`: create on BTP (BTP provider) or inside a Cloud Foundry space (CF provider)? Default is BTP."
- If the user chooses CF and the subaccount has **more than one** space, ask which space (list the spaces from `specs/landscape.md`); if it has exactly one, use that space without asking. Record the chosen space as `cf_space: <name>`.
- If the user chooses CF but the subaccount's Cloud Foundry environment has **no space** defined in `specs/landscape.md`, **STOP**: a CF service instance requires a space. Tell the user to add a Cloud Foundry space via `/btp-iac.accounts` (the authoritative source for spaces), then re-run `/btp-iac.services`. Do not invent a space name here.

Record `location` for every service instance, plus `cf_space` when `location: cf`. These are passed through to `specs/services.md` and consumed by `/btp-iac.tasks`, `/btp-iac.design`, and `/btp-iac.generate` to select the correct provider and CF space.

---

## Workflow

Read `specs/scenario.md` and `specs/landscape.md` to understand the application requirements and the account topology. Run Service Type Classification before resolving dependencies.

For each subaccount, resolve:
- Required entitlements (service + plan) — includes `entitlement-only` services
- Subscriptions (services with `consumption_type: subscription`)
- Service instances (`consumption_type: instance`) with configuration parameters, `location`, and `cf_space` when CF-located
- Dependencies between services (ordered)

Write `specs/services.md` with the full dependency-ordered list. Each service entry **must** record `consumption_type` (`instance` | `subscription` | `entitlement-only`). Each `instance` entry **must** also record `location` (`btp` | `cf`) and, when `location: cf`, `cf_space: <name>`. This file is the direct input to `/btp-iac.tasks`.

## Next step

Next: `/btp-iac.security` — set up IdP trust, roles, and role collection assignments.
