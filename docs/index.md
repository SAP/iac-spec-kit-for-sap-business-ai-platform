# Infrastructure-as-Code Specification Toolkit for SAP Business AI Platform

The Infrastructure-as-Code Specification Toolkit for SAP Business AI Platform is an open-source CLI tool that scaffolds AI-assisted Terraform projects for [SAP Business AI Platform](https://www.sap.com/products/technology-platform.html). It generates a project structure with AI-agent skills that guide you through the full infrastructure-as-code lifecycle — from describing your scenario to generating production-ready [Terraform](https://terraform.io/) HCL.

## How it works

1. **Install** the `sap-iac` CLI and scaffold a project — see [Installation & Setup](setup.md).
2. **Describe** your application in plain language, then let the AI agent drive the command workflow — see [Usage](usage.md).
3. **Generate and review** the Terraform, then apply it against your account yourself using Terraform.

!!! tip "New here?"
    Start with the [usage walkthrough](walkthrough.md) — it runs a single application end to end, from `sap-iac init` to a `terraform plan`.

## Where to go next

- [Installation & Setup](setup.md) — prerequisites and how to install the CLI.
- [Usage](usage.md) — the `sap-iac init` command and the AI-agent workflow.
- [Command reference](commands/index.md) — every `sap-iac.*` command in detail.
- [Usage walkthrough](walkthrough.md) — a complete end-to-end run.
- [Scenario examples](scenarios.md) — ready-made scenario descriptions to adapt.
- [Troubleshooting](troubleshooting.md) — common warnings and errors.
