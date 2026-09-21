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

<!-- FOLLOW-UPS: max number of clarifying questions and their topics (runtime/sizing, destinations, setup structure) -->
Asks up to three targeted follow-up questions about unresolved runtime and Cloud Foundry sizing, destinations, and setup structure, then writes the results to `specs/scenario.md`.

<!-- OUTPUT: path where the structured requirements are written -->
`specs/scenario.md`

## Tool preferences

Before using `WebFetch` to look up SAP documentation, check if the `sap-docs` MCP server is available (tools prefixed `mcp__sap-docs__*`). If yes, use it. If not, fall back to `WebFetch`.

## BTP platform validation

Locate the project root and read `.btp-iac/platform-validation.md`, `memory/global-account.md`, and, when it exists, `memory/governance.md`. When scenario input names regions, service offerings, subscriptions, or service-plan pairs, apply the same platform-validation contract as `/btp-iac.govern`: use the recorded CLI first, otherwise this agent's recorded BTP MCP route, and accept user input without blocking when no route is recorded. Use only the subdomain configured at initialization in `memory/global-account.md`; if it is absent or blank, retain user input without a targeted live check and do not ask for a subdomain.

Target the subdomain with `btp target --global-account <subdomain>` before CLI entitlement or region commands. Check named offerings, subscriptions, and plan pairs with `btp list accounts/entitlement`; check regions with `btp list accounts/available-region` and ignore `NEO` entries. Read `## Regions → Preferred infrastructure provider` from `memory/governance.md`; when it is not `none`, compare it with unambiguous provider metadata returned for each named region. Never infer a provider from a region code. Log a warning for a known mismatch, and treat missing or unmappable provider metadata as advisory only. Do not write provider preferences or response-field metadata to `.btp-iac/platform-validation.md`. Use equivalent scoped BTP MCP operations when CLI is unavailable. A completed negative result requires a user-selected replacement; an authentication, targeting, or lookup failure requires the user to resolve setup. If no live region route exists, use the SAP Help region list as advisory only and retain user input.

## Governance-aware follow-ups

Before asking any follow-up, use `memory/governance.md` when it exists. Treat `## Account Setup → Allowed environments` as the resolved runtime constraint and `## Naming → Environments` as the resolved environment-tier set. Do not ask an environment question for information already supplied by the scenario input or either governance field.

Ask an environment follow-up only when the scenario input and governance together do not sufficiently determine a required runtime or setup-structure decision. If governance permits exactly one runtime, use it without asking. If it permits multiple runtimes, ask only when the scenario does not select one and the choice is required; ask only about that unresolved choice. Do not ask for a global-account subdomain.

## Follow-up questions

Ask at most one question for each unresolved topic, in this order:

1. **Runtime and sizing** — determine the required runtime. When Cloud Foundry is selected or permitted, ask for the memory allocation or sizing required by each Cloud Foundry application. Do not ask for Cloud Foundry memory when Cloud Foundry is not selected.
2. **Destinations** — ask whether integrations require destinations; if they do, collect each required destination and its purpose.
3. **Setup structure** — ask how the setup should be structured, for example whether each stage has its own subaccount, rather than asking about “environment isolation.”

Skip a question when the scenario input and governance already provide a sufficient answer.
