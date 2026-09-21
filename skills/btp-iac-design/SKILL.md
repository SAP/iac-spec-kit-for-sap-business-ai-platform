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

If this workflow needs a live BTP availability check, read `<project-root>/.btp-iac/platform-validation.md` and prefer its recorded CLI route, then this agent's recorded BTP MCP route. Do not block when no route is recorded; retain user input. If a recorded route fails to authenticate, target, or answer, ask the user to resolve it before relying on the result.
