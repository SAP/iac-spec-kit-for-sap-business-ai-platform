---
name: btp-iac-scenario
description: Translates a plain-language app description into structured BTP infrastructure requirements.
license: Apache-2.0
metadata:
  author: SAP
  version: "1.0"
---

# BTP IaC — Scenario

<!-- INPUT: what the user provides as the starting description -->
Takes a plain-language description of the app or service to be deployed and translates it into a structured set of BTP infrastructure requirements.

<!-- FOLLOW-UPS: max number of clarifying questions and their topics (runtime, integrations, environment isolation) -->
Asks up to three targeted follow-up questions about runtime, integration points, and environment isolation, then writes the results to `specs/scenario.md`.

<!-- OUTPUT: path where the structured requirements are written -->
`specs/scenario.md`

## Tool preferences

Before using `WebFetch` to look up SAP documentation, check if the `sap-docs` MCP server is available (tools prefixed `mcp__sap-docs__*`). If yes, use it. If not, fall back to `WebFetch`.
