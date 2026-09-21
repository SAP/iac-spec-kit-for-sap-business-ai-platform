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

If this workflow needs a live BTP availability check, read `<project-root>/.btp-iac/platform-validation.md` and prefer its recorded CLI route, then this agent's recorded BTP MCP route. If no route is recorded, retain user input without blocking. If a recorded route cannot authenticate, target, or complete its lookup, ask the user to resolve it before relying on platform data.
