---
name: btp-iac-design
description: Translates the task list into a concrete Terraform folder structure and annotates each task with its target file path.
license: Apache-2.0
metadata:
  author: SAP
  version: "1.0"
---

# BTP IaC — Design

Translates the task list into a concrete Terraform folder structure — how resources are split across files, whether modules are introduced, and how environment-specific variable files are organised.

Reads `specs/tasks.md`. Annotates each task in `specs/tasks.md` with the file path it will be written to.

## BTP platform validation

### BTP operation safety boundary

The prohibition on state-changing CLI commands excludes the permitted target prelude.

When invoking the BTP CLI or any BTP MCP tool, perform **only read or list retrievals**. For the BTP CLI, invoke only documented read/list commands (for example, `btp list ...`); for BTP MCP, invoke only a tool explicitly documented as a read/list lookup. `btp target --global-account <subdomain>` is the sole permitted account-selection prelude and may be used only immediately before those read/list CLI commands. Never invoke, suggest, or approve a BTP operation that creates, updates, deletes, assigns, unassigns, enables, disables, or otherwise mutates BTP state — even when requested by the user. Do not run login, config, profile, or any other state-changing CLI command.

If this workflow needs a live BTP availability check, read `<project-root>/.btp-iac/platform-validation.md` and prefer its recorded CLI route, then this agent's recorded BTP MCP route. Do not block when no route is recorded; retain user input. If a recorded route fails to authenticate, target, or answer, ask the user to resolve it before relying on the result.

## Next step

Next: `/btp-iac.generate` — write and validate all Terraform HCL.
