---
name: btp-iac-scenario
description: Translates a plain-language setup description into structured BTP infrastructure requirements.
license: Apache-2.0
metadata:
  author: SAP
  version: "1.0"
---

# BTP IaC — Scenario

<!-- INPUT: what the user provides as the starting description -->
Takes a plain-language description of the setup to be deployed and translates it into a structured set of BTP infrastructure requirements.

<!-- FOLLOW-UPS: max number of clarifying questions and their topics (runtime, integrations, environment isolation) -->
Asks up to three targeted follow-up questions about runtime, integration points, and environment isolation, then writes the results to `specs/scenario.md`.

<!-- OUTPUT: path where the structured requirements are written -->
`specs/scenario.md`

## Tool preferences

Before using `WebFetch` to look up SAP documentation, check if the `sap-docs` MCP server is available (tools prefixed `mcp__sap-docs__*`). If yes, use it. If not, fall back to `WebFetch`.

## BTP platform validation

Locate the project root and read `.btp-iac/platform-validation.md` and `memory/global-account.md`. When scenario input names regions, service offerings, subscriptions, or service-plan pairs, apply the same platform-validation contract as `/btp-iac.govern`: use the recorded CLI first, otherwise this agent's recorded BTP MCP route, and accept user input without blocking when no route is recorded. Use only the subdomain configured at initialization in `memory/global-account.md`; if it is absent or blank, retain user input without a targeted live check and do not ask for a subdomain.

Target the subdomain with `btp target --global-account <subdomain>` before CLI entitlement or region commands. Check named offerings, subscriptions, and plan pairs with `btp list accounts/entitlement`; check regions with `btp list accounts/available-region` and ignore `NEO` entries. When `## Regions → Preferred infrastructure provider` in `memory/governance.md` is not `none`, compare it with unambiguous provider metadata returned for each named region. On the first successful lookup that exposes this metadata, record its exact field name as `- Region provider field: <field>` in `.btp-iac/platform-validation.md`. Never infer a provider from a region code. Log a warning for a known mismatch, and treat missing or unmappable provider metadata as advisory only. Use equivalent scoped BTP MCP operations when CLI is unavailable. A completed negative result requires a user-selected replacement; an authentication, targeting, or lookup failure requires the user to resolve setup. If no live region route exists, use the SAP Help region list as advisory only and retain user input.
