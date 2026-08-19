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
