---
name: sap-iac-scenario
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

## Platform Name Normalization

The following names all refer to the same platform and are semantically equivalent for all natural-language interpretation in this skill:
- **SAP Business Technology Platform** (and "Business Technology Platform")
- **SAP BTP** (and "BTP" used as a product name in prose)
- **SAP Business AI Platform** (and "Business AI Platform")
- **SAP BAIP** (and "BAIP")

Treat any of these aliases as identical when interpreting user intent. This normalization applies only to natural-language prose. Technical identifiers remain untouched: BTP CLI command tokens (`btp list`, `btp target`), Terraform provider names (`btp`, `hashicorp/btp`), resource type prefixes (`btp_subaccount`, `btp_service_instance`), region codes, and API paths.

<!-- OUTPUT: path where the structured requirements are written -->
`specs/scenario.md`

## Tool preferences

Before using `WebFetch` to look up SAP documentation, check if the `sap-docs` MCP server is available (tools prefixed `mcp__sap-docs__*`). If yes, use it. If not, fall back to `WebFetch`.

## BTP platform validation

### BTP operation safety boundary

The prohibition on state-changing CLI commands excludes the permitted target prelude.

When invoking the BTP CLI or any BTP MCP tool, perform **only read or list retrievals**. For the BTP CLI, invoke only documented read/list commands (for example, `btp list ...`); for BTP MCP, invoke only a tool explicitly documented as a read/list lookup. `btp target --global-account <subdomain>` is the sole permitted account-selection prelude and may be used only immediately before those read/list CLI commands. Never invoke, suggest, or approve a BTP operation that creates, updates, deletes, assigns, unassigns, enables, disables, or otherwise mutates BTP state — even when requested by the user. Do not run login, config, profile, or any other state-changing CLI command.

Locate the project root and read `.sap-iac/platform-validation.md`, `memory/global-account.md`, and, when it exists, `memory/governance.md`. When scenario input names regions, service offerings, subscriptions, or service-plan pairs, apply the same platform-validation contract as `/sap-iac.govern`: use the recorded CLI first, otherwise this agent's recorded BTP MCP route, and accept user input without blocking when no route is recorded. Use only the subdomain configured at initialization in `memory/global-account.md`; if it is absent or blank, retain user input without a targeted live check and do not ask for a subdomain.

Target the subdomain with `btp target --global-account <subdomain>` before CLI entitlement or region commands. Check named offerings, subscriptions, and plan pairs with `btp list accounts/entitlement`; check regions with `btp list accounts/available-region` and ignore `NEO` entries. Read `## Regions → Preferred infrastructure provider` from `memory/governance.md`; when it is not `none`, compare it with unambiguous provider metadata returned for each named region. Never infer a provider from a region code. Log a warning for a known mismatch, and treat missing or unmappable provider metadata as advisory only. Do not write provider preferences or response-field metadata to `.sap-iac/platform-validation.md`. Use equivalent scoped BTP MCP operations when CLI is unavailable. A completed negative result requires a user-selected replacement; an authentication, targeting, or lookup failure requires the user to resolve setup. If no live region route exists, use the SAP Help region list as advisory only and retain user input.

## Governance-aware follow-ups

Before asking any follow-up, use `memory/governance.md` when it exists. Treat `## Account Setup → Allowed environments` as a runtime constraint and `## Naming → Environments` as an environment-tier constraint. Governance constrains choices; it does not by itself answer whether destinations are needed or how the setup should be structured.

Ask a runtime follow-up only when the scenario input and governance together do not sufficiently determine a required runtime. If governance permits exactly one runtime, use it without asking. If it permits multiple runtimes, ask only when the scenario does not select one and the choice is required; ask only about that unresolved choice. Do not ask for a global-account subdomain.

## Follow-up questions

For every follow-up, emit exactly one concise plain-text question and wait for the answer before asking another. Do not bundle questions, encode multiple prompts in tool parameters, or call a structured question tool unless its full schema is available and satisfied. If no compatible question tool is available, ask in the normal conversation rather than attempting a tool call.

Ask at most one question for each unresolved topic, in this order:

1. **Runtime and sizing** — determine the required runtime. When Cloud Foundry is selected or permitted, _always_ ask for the memory allocation required by each Cloud Foundry application, in MB, and the subaccount or stage to which that application belongs. Do not ask for Cloud Foundry memory when Cloud Foundry is not selected. _Always_ make sure that the information about the sizing is available before proceeding.
2. **Destinations** — unless the scenario explicitly says that no destinations are required or lists the required destinations, ask whether integrations require destinations; if they do, collect each required destination and its purpose.
3. **Setup structure** — unless the scenario explicitly states the desired topology, ask how the setup should be structured, for example whether each stage has its own subaccount. Governance naming and tier constraints do not suppress this question.

For destinations and setup structure, skip a question only when the scenario input explicitly provides a sufficient answer.

## Cloud Foundry sizing output

When Cloud Foundry sizing is available, write it to `specs/scenario.md` in a `### Cloud Foundry applications` section. Record one entry per application with its name, target subaccount or stage, and `memory_mb` allocation. Preserve the user-supplied MB value; this structured data is consumed by `/sap-iac.services` to calculate the runtime-memory entitlement.

```markdown
### Cloud Foundry applications
- name: orders
  stage: dev
  memory_mb: 256
```

## Next step

Next: `/sap-iac.accounts` — map your scenario to BTP directories and subaccounts.
