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

Read `<project-root>/.btp-iac/platform-validation.md` after locating the project root. Before writing `specs/services.md`, validate every resolved entitlement, subscription, service offering, and plan against the global-account subdomain in `specs/landscape.md`. Prefer the recorded CLI route: run `btp target --global-account <subdomain>` and inspect `btp list accounts/entitlement`. If CLI was unavailable at initialization, use an equivalent scoped BTP MCP entitlement/subscription operation only when this agent is recorded as having BTP MCP support.

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

## Workflow

Read `specs/scenario.md` and `specs/landscape.md` to understand the application requirements and the account topology.

For each subaccount, resolve:
- Required entitlements (service + plan)
- Subscriptions (SaaS applications)
- Service instances with configuration parameters
- Dependencies between services (ordered)

Write `specs/services.md` with the full dependency-ordered list. This file is the direct input to `/btp-iac.tasks`.
