# `btp-iac.services`

!!! abstract "Summary"
    **Role:** Required. · **Reads:** `specs/scenario.md`, `specs/landscape.md` (+ governance) · **Writes:** `specs/services.md`

Resolves your requirements into a dependency-ordered list of BTP entitlements, subscriptions, and service instances for each subaccount.

## When to run it

After `btp-iac.accounts`, once the account topology exists.

## Inputs and outputs

| | |
|---|---|
| **Reads** | `specs/scenario.md` and `specs/landscape.md`, and `memory/global-account.md` and `memory/governance.md` if present. |
| **Writes** | `specs/services.md` — per subaccount: required entitlements (service + plan), subscriptions, service instances with their configuration parameters, and the order in which they must be created. |

## Behaviour

- Maps the scenario's needs onto concrete BTP services and plans for each subaccount in the landscape.
- Resolves the dependency ordering between services so they can be created in a valid sequence.
- Prefers the `sap-docs` MCP server over web fetches for SAP documentation.

!!! note "Governance enforcement"
    For each service instance, `services` determines the environment tier from `specs/landscape.md` and validates the chosen plan against the Service Plans rules, **stopping on a violation** unless `- Override: true` is set. Without a governance file it proceeds without enforcement and says so.

!!! note "BTP platform validation"
    Before writing `specs/services.md`, `services` validates every resolved entitlement, subscription, service offering, and plan against the configured global-account subdomain from `memory/global-account.md`, using **read-only** BTP CLI/MCP lookups (no mutations). If a resolved item is unavailable in that global account, it asks for a valid replacement before writing.

## Example

```
/btp-iac.services
```

For the HR leave-request subaccount, `services` records the SAP HANA Cloud and XSUAA entitlements and their instances in `specs/services.md`:

```markdown
## hr-leave-prod
### Entitlements
- hana-cloud / hana (SAP HANA Cloud)
- xsuaa / application
### Service instances (in order)
1. xsuaa (application)
2. hana-cloud (hana) — depends on: entitlement
```

## Related

- Requires [`btp-iac.accounts`](accounts.md).
- Feeds [`btp-iac.security`](security.md) and [`btp-iac.tasks`](tasks.md).

Next: [`btp-iac.security`](security.md).
