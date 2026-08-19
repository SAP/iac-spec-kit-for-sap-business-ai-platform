---
name: btp-iac-tasks
description: Consolidates landscape, services, and trust specs into a single dependency-ordered task list.
license: Apache-2.0
metadata:
  author: SAP
  version: "1.0"
---

# BTP IaC — Tasks

Consolidates `specs/landscape.md`, `specs/services.md`, and `specs/trust.md` into a single dependency-ordered task list with IDs and parallel execution markers.

Produces `specs/tasks.md`, which is the direct input to `/btp-iac.design` and `/btp-iac.generate`.
