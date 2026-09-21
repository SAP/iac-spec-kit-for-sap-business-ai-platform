---
name: btp-iac-tasks
description: Consolidates landscape, services, and trust specs into a single dependency-ordered task list.
license: Apache-2.0
metadata:
  author: SAP
  version: "1.0"
---

# BTP IaC — Tasks

Consolidates `specs/landscape.md`, `specs/services.md`, and `specs/trust.md` into a single dependency-ordered task list with IDs and parallel execution markers. Preserve every Cloud Foundry environment, Kyma environment, and Cloud Foundry space from the landscape as a task with its subaccount, type, and name; order each after its subaccount and Cloud Foundry spaces after their Cloud Foundry environment.

Produces `specs/tasks.md`, which is the direct input to `/btp-iac.design` and `/btp-iac.generate`.

## BTP platform validation

### BTP operation safety boundary

The prohibition on state-changing CLI commands excludes the permitted target prelude.

When invoking the BTP CLI or any BTP MCP tool, perform **only read or list retrievals**. For the BTP CLI, invoke only documented read/list commands (for example, `btp list ...`); for BTP MCP, invoke only a tool explicitly documented as a read/list lookup. `btp target --global-account <subdomain>` is the sole permitted account-selection prelude and may be used only immediately before those read/list CLI commands. Never invoke, suggest, or approve a BTP operation that creates, updates, deletes, assigns, unassigns, enables, disables, or otherwise mutates BTP state — even when requested by the user. Do not run login, config, profile, or any other state-changing CLI command.

If this workflow needs a live BTP availability check, read `<project-root>/.btp-iac/platform-validation.md` and prefer its recorded CLI route, then this agent's recorded BTP MCP route. If no route is recorded, retain user input without blocking. If a recorded route cannot authenticate, target, or complete its lookup, ask the user to resolve it before relying on platform data.
