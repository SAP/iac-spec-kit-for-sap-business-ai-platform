# `btp-iac.security`

!!! abstract "Summary"
    **Role:** Required. · **Reads:** `specs/scenario.md`, `specs/landscape.md` (+ governance) · **Writes:** `specs/trust.md`

Identifies the authentication and authorisation requirements — identity-provider trust, role collections, role-template assignments, and user/group assignments.

## When to run it

Normally after `btp-iac.services`, following the site-wide `services` → `security` order. It has no hard dependency on `specs/services.md`, though — `security` reads only `specs/scenario.md` and `specs/landscape.md`.

## Inputs and outputs

| | |
|---|---|
| **Reads** | `specs/scenario.md` and `specs/landscape.md`, and `memory/governance.md` if present. It also reads `.btp-iac/platform-validation.md` for an optional live BTP availability check, making only read/list BTP calls. |
| **Writes** | `specs/trust.md` — per subaccount: platform and application IdP trust, role collections and their role-template assignments, and user/group assignments. |

## Behaviour

- Derives trust and authorisation requirements from the scenario and landscape.
- Prefers the `sap-docs` MCP server over web fetches for SAP documentation.

!!! note "Governance enforcement"
    `security` stops when the guardrails require a custom IdP that is not configured, or when required default role-collection assignments are missing — unless `- Override: true` is set in `memory/governance.md` (see [`btp-iac.govern`](govern.md)).

## Example

```
/btp-iac.security
```

For the HR leave-request app, `security` records XSUAA-based authentication and a role collection in `specs/trust.md`:

```markdown
## hr-leave-prod
### IdP trust
- Application IdP: default — no custom IdP required
### Role collections
- HR_Leave_Employee -> role template: Employee
### User assignments
- (assign employees to HR_Leave_Employee)
```

## Related

- Requires [`btp-iac.accounts`](accounts.md).
- Optional next step for external systems: [`btp-iac.connectivity`](connectivity.md).
- Feeds [`btp-iac.tasks`](tasks.md).
