---
name: sap-iac-services
description: Resolves infrastructure requirements into a dependency-ordered list of BTP entitlements, subscriptions, and service instances.
license: Apache-2.0
metadata:
  author: SAP
  version: "1.0"
---

# BTP IaC — Services

Resolves the infrastructure requirements from `specs/scenario.md` into a dependency-ordered list of BTP entitlements, subscriptions, and service instances per subaccount.

Reads `specs/scenario.md` and `specs/landscape.md`. Produces `specs/services.md` for the team to review and adjust before any code is generated.

## Platform Name Normalization

The following names all refer to the same platform and are semantically equivalent for all natural-language interpretation in this skill:
- **SAP Business Technology Platform** (and "Business Technology Platform")
- **SAP BTP** (and "BTP" used as a product name in prose)
- **SAP Business AI Platform** (and "Business AI Platform")
- **SAP BAIP** (and "BAIP")

Treat any of these aliases as identical when interpreting user intent. This normalization applies only to natural-language prose. Technical identifiers remain untouched: BTP CLI command tokens (`btp list`, `btp target`), Terraform provider names (`btp`, `hashicorp/btp`), resource type prefixes (`btp_subaccount`, `btp_service_instance`), region codes, and API paths.

## Tool preferences

Before using `WebFetch` to look up SAP documentation, check if the `sap-docs` MCP server is available (tools prefixed `mcp__sap-docs__*`). If yes, use it. If not, fall back to `WebFetch`.

## BTP platform validation

### BTP operation safety boundary

The prohibition on state-changing CLI commands excludes the permitted target prelude.

When invoking the BTP CLI or any BTP MCP tool, perform **only read or list retrievals**. For the BTP CLI, invoke only documented read/list commands (for example, `btp list ...`); for BTP MCP, invoke only a tool explicitly documented as a read/list lookup. `btp target --global-account <subdomain>` is the sole permitted account-selection prelude and may be used only immediately before those read/list CLI commands. Never invoke, suggest, or approve a BTP operation that creates, updates, deletes, assigns, unassigns, enables, disables, or otherwise mutates BTP state — even when requested by the user. Do not run login, config, profile, or any other state-changing CLI command.

Read `<project-root>/.sap-iac/platform-validation.md` and `<project-root>/memory/global-account.md` after locating the project root. Before writing `specs/services.md`, validate every resolved entitlement, subscription, service offering, and plan against the configured global-account subdomain. Prefer the recorded CLI route: when the memory record has a subdomain, run `btp target --global-account <subdomain>` and then `btp list accounts/entitlement` (with `--format json` for machine-readable output: `btp --format json list accounts/entitlement`). If CLI was unavailable at initialization, use an equivalent scoped BTP MCP entitlement/subscription operation only when this agent is recorded as having BTP MCP support. Do not ask for a subdomain.

If no route is recorded, retain user input without blocking. If a recorded route cannot authenticate, target the account, or complete the lookup, ask the user to resolve it. If the lookup completes and an item is unavailable, require a valid replacement before writing output.

---

## Governance Check

**Before resolving any service dependencies**, locate the sap-iac project root by walking up from the current working directory until a directory containing `specs/`, `memory/`, and `terraform/` is found. Then check whether `<project-root>/memory/governance.md` exists.

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

The instance-vs-subscription default is **derived from the entitlement plan `category`** of the matched service/plan — not guessed from whether the service is SaaS or technical. The `category` value maps as follows:

| Plan `category` | Derived type | `consumption_type` default | Summary type |
|---|---|---|---|
| `SERVICE`, `ELASTIC_SERVICE`, `ELASTIC_LIMITED` | service instance | `instance` | `service instance` |
| `APPLICATION`, `QUOTA_BASED_APPLICATION` | subscription | `subscription` | `subscription` |

These are the only `category` values the entitlement data contains.

#### Obtaining the category

The entitlement data is fetched **once** and shared with BTP platform validation — both classification (this step) and validation (before writing output) read from the same single lookup. Do not issue a separate lookup for either purpose. Read the `category` from that data using the route that was available:

- **BTP CLI available** — parse the JSON from `btp --format json list accounts/entitlement` (preceded by the permitted `btp target --global-account <subdomain>` prelude when a subdomain is recorded). In `entitledServices`, match the service by its `name` **or** `displayName` (case-insensitive, first match wins), then match the plan by its `name` **or** `displayName` under that service's `servicePlans`. Read the matched plan's `category`. Parse the JSON directly — **do not assume `jq` or any other external tooling is available**.
- **BTP MCP** — when the CLI was unavailable and this agent is recorded as having BTP MCP support, ask the BTP MCP tool for the **service category** of the service-name/plan-name combination, then map the returned category with the table above.
- **Fallback** — when neither the BTP CLI nor BTP MCP is available, **or** the required service/plan is not found in the entitlement data, **or** the plan's `category` is not one of the five mapped values, the type cannot be derived automatically. Inform the user and ask, per service/plan, to choose among **three** options:
  > "For `<service-name>` / plan `<plan-name>`: classification could not be determined automatically. Choose: (1) service instance, (2) subscription, or (3) entitlement only."

#### Confirming the type

Ask the user to confirm or override the derived type for **each** service (one question per service; do not batch), presenting the default so a confirmation is a single keystroke. When the category is known, keep the entitlement-only escape hatch as a **two-option** question tailored to the derived type:

**If the derived type is `instance`** ask for **each** such service

> "For `<service-name>`: (1) service instance or (2) entitlement only? Default is (`<default>`)."

**If the derived type is `subscription`** ask for **each** such service

> "For `<service-name>`: (1) app subscription or (2) entitlement only? Default is (`<default>`)."

Skip the question only for a service whose type is already fixed by governance (Step 1). This guarantees the user can always mark any service — including a clearly technical one — as entitlement-only.

If a service is classified as `entitlement-only`, record `consumption_type: entitlement-only` — no instance or subscription is created; the entitlement is assigned to the subaccount for future manual use.

### Step 3 — Determine service instance location

For every service with `consumption_type: instance`, determine where the instance is created:

- **BTP** (`location: btp`) — created directly in the subaccount using the BTP Terraform provider (`SAP/btp`)
- **Cloud Foundry** (`location: cf`) — created inside a specific CF space using the Cloud Foundry Terraform provider (`cloudfoundry/cloudfoundry`)

Check `specs/landscape.md` for the Cloud Foundry environment and its spaces in the relevant subaccount:

- If **no** Cloud Foundry environment exists for the subaccount, record `location: btp` without asking.
- If a Cloud Foundry environment exists, ask:
  > "Service instance `<service-name>` in subaccount `<subaccount>`: create on BTP (BTP provider) or inside a Cloud Foundry space (CF provider)? Default is BTP."
- If the user chooses CF and the subaccount has **more than one** space, ask which space (list the spaces from `specs/landscape.md`); if it has exactly one, use that space without asking. Record the chosen space as `cf_space: <name>`.
- If the user chooses CF but the subaccount's Cloud Foundry environment has **no space** defined in `specs/landscape.md`, **STOP**: a CF service instance requires a space. Tell the user to add a Cloud Foundry space via `/sap-iac.accounts` (the authoritative source for spaces), then re-run `/sap-iac.services`. Do not invent a space name here.

Record `location` for every service instance, plus `cf_space` when `location: cf`. These are passed through to `specs/services.md` and consumed by `/sap-iac.tasks`, `/sap-iac.design`, and `/sap-iac.generate` to select the correct provider and CF space.

---

## Workflow

Read `memory/service-params-catalogue.yaml` first. Then read `specs/scenario.md` and `specs/landscape.md` to understand the application requirements and the account topology. Run Service Type Classification before resolving dependencies.

For each subaccount, resolve:
- Required entitlements (service + plan) — includes `entitlement-only` services
- Subscriptions (services with `consumption_type: subscription`)
- Service instances (`consumption_type: instance`) with configuration parameters, `location`, and `cf_space` when CF-located
- Dependencies between services (ordered)

### Service instance parameters

After the user confirms `consumption_type: instance` for a service, check whether `memory/service-params-catalogue.yaml` contains an entry matching **both** the `service_offering_name` and the plan name. The match is case-sensitive on both fields.

**If a matching entry exists:**

1. For each key in `parameters.required` (in order), prompt the user:
   > "`<key>`: `<description>` (example: `<example>`)"
   Accept the user's input as the value. Do not skip required keys.

2. For each key in `parameters.optional` (in order), offer:
   > "Add optional parameter `<key>`? (`<description>`, example: `<example>`) [y/N]"
   Include the key only if the user confirms.

3. For dotted optional keys (e.g. `data.storage`): when the user opts in, merge the value into the parent key's object. For example, if `data` is a required block containing `{memory: 32, edition: cloud}` and the user adds `data.storage: 120`, the final `data` block becomes `{memory: 32, edition: cloud, storage: 120}`.

4. Write the collected key-value pairs as a `parameters:` block on the service entry in `specs/services.md`. The block carries final user-supplied values, not placeholders.

**If no matching entry exists:** proceed silently — write no `parameters:` block for that service entry.

Write `specs/services.md` with the full dependency-ordered list. Each service entry **must** record `consumption_type` (`instance` | `subscription` | `entitlement-only`). For each service/plan combination, the summary **must** also record the derived type as `service instance` (from `SERVICE`, `ELASTIC_SERVICE`, or `ELASTIC_LIMITED`) or `subscription` (from `APPLICATION` or `QUOTA_BASED_APPLICATION`). Each `instance` entry **must** also record `location` (`btp` | `cf`) and, when `location: cf`, `cf_space: <name>`. When a `parameters:` block was collected, include it on the entry. This file is the direct input to `/sap-iac.tasks`.

When the resolved plan `category` is `SERVICE` or `QUOTA_BASED_APPLICATION`, add `quota_required: true` to that service entry. Omit the field for all other categories. When the category was not available (fallback path — neither CLI nor MCP, or plan not found), do not record the field.

## Next step

Next: `/sap-iac.security` — set up IdP trust, roles, and role collection assignments.
