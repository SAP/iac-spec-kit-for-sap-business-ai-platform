# BTP IaC — Security

Identifies the authentication, authorisation, and connectivity requirements — role collection definitions, user assignments, IdP trust configurations, and destination configurations for external system access.

Reads `specs/scenario.md` and `specs/landscape.md`. Produces `specs/trust.md`.

## Tool preferences

Before using `WebFetch` to look up SAP documentation, check if the `sap-docs` MCP server is available (tools prefixed `mcp__sap-docs__*`). If yes, use it. If not, fall back to `WebFetch`.

---

## Governance Check

**Before defining any security configuration**, locate the btp-iac project root by walking up from the current working directory until a directory containing `specs/`, `memory/`, and `terraform/` is found. Then check whether `<project-root>/memory/governance.md` exists.

**If it does not exist:** proceed without constraints and note: "No governance rules found — proceeding without enforcement."

**If it exists**, load all rules and validate every security decision against the `## Security` section.

### IdP validation
- If `Custom IdP: required` and no custom IdP is being configured: **STOP**
  > "GOVERNANCE VIOLATION: Governance requires a custom Identity Provider. Configure a custom IdP trust or add `- Override: true` to memory/governance.md."

### Role collection validation
- If `Default role collections` specifies assignments and the configuration does not include them: **STOP**
  > "GOVERNANCE VIOLATION: Governance requires default role collection assignments (`<required>`). Include these assignments or add `- Override: true` to memory/governance.md."

### Destination validation
- If `Destinations` restricts access (e.g. "internal systems only") and an external destination is being configured: **STOP**
  > "GOVERNANCE VIOLATION: Governance restricts destinations to `<restriction>`. Remove the external destination or add `- Override: true` to memory/governance.md."

**If `- Override: true` is set in `<project-root>/memory/governance.md`:** log a warning for each violation and continue instead of stopping.

---

## Workflow

Read `specs/scenario.md` and `specs/landscape.md` to understand authentication flows, user roles, and external system integrations.

Define for each subaccount:
- IdP trust configurations (platform and application)
- Role collections with role template assignments
- User and group assignments to role collections
- Destination configurations for external system access
- Connectivity service setup if required

Write `specs/trust.md` with the complete security and connectivity configuration. This file is the direct input to `/btp-iac.tasks`.
