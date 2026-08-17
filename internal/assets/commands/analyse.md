# BTP IaC — Analyse

Reads application source code to extract concrete infrastructure signals that the user cannot be expected to know — service dependencies, memory requirements, role definitions from the security descriptor, and build artifact details.

Ask the user for the path to their application source code before proceeding. Enriches `specs/scenario.md` in place with these findings.

Supports MTA apps (mta.yaml, xs-security.json, xs-app.json), CAP apps (package.json, schema.cds, xs-security.json), and plain CF apps (manifest.yml).

## Tool preferences

Before using `WebFetch` to look up SAP documentation, check if the `sap-docs` MCP server is available (tools prefixed `mcp__sap-docs__*`). If yes, use it. If not, fall back to `WebFetch`.
