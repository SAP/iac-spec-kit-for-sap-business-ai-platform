# `sap-iac.security`

!!! abstract "Summary"
    **Role:** Required. · **Reads:** `specs/scenario.md`, `specs/landscape.md` (+ governance) · **Writes:** `specs/trust.md`

Identifies the authentication and authorisation requirements — identity-provider trust, role collections, and user/group assignments.

## When to run it

Normally after `sap-iac.services`, following the site-wide `services` → `security` order. It has no hard dependency on `specs/services.md`, though — `security` reads only `specs/scenario.md` and `specs/landscape.md`.

## Inputs and outputs

| | |
|---|---|
| **Reads** | `specs/scenario.md` and `specs/landscape.md`, and `memory/governance.md` if present. It also reads `.sap-iac/platform-validation.md` for an optional live BTP availability check, making only read/list BTP calls. |
| **Writes** | `specs/trust.md` — per subaccount: platform and application IdP trust, role collections with their structured roles list (may be empty), and user/group assignments. |

## Behaviour

- Derives trust and authorisation requirements from the scenario and landscape.
- For a custom IdP, records the IdP URL as the trust configuration's `identity_provider`. Its derived origin remains available for assignments, but `btp_subaccount_trust_configuration.origin` is recorded only when the user explicitly supplies a separate trust-configuration origin.
- Prefers the `sap-docs` MCP server over web fetches for SAP documentation.

!!! note "Governance enforcement"
    `security` stops when the guardrails require a custom IdP that is not configured, or when required default role-collection assignments are missing — unless `- Override: true` is set in `memory/governance.md` (see [`sap-iac.govern`](govern.md)).

## Example

```
/sap-iac.security
```

For the HR leave-request app, `security` asks whether role collections are needed (yes) and whether individual roles should be assigned, then records the result in `specs/trust.md`:

```markdown
## hr-leave-prod
### IdP trust
- Application IdP: default — no custom IdP required
### Role collections
- **Role collection**: HR_Leave_Employee
  - Description: Employees using the leave-request app
  - Roles:
    - role_name: Employee, role_template_name: Employee, role_template_app_id: hr-leave-xsuaa!b1
### User assignments
- (assign employees to HR_Leave_Employee)
```

For a custom IdP, the trust section keeps derived and explicit values distinct:

```markdown
### Custom IdP
- IdP URL: https://corp.accounts.ondemand.com
- Origin: corp-platform
- Trust configuration origin: corp-custom
- Trust configuration origin explicit: true
```

The final two lines are optional and must appear together. If either is absent, generated `btp_subaccount_trust_configuration` resources include `identity_provider` but omit `origin`.

## Related

- Requires [`sap-iac.accounts`](accounts.md).
- Optional next step for external systems: [`sap-iac.connectivity`](connectivity.md).
- Feeds [`sap-iac.tasks`](tasks.md).
