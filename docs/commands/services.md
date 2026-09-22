# `btp-iac.services`

!!! abstract "Summary"
    **Role:** Required. · **Reads:** `specs/scenario.md`, `specs/landscape.md` (+ governance) · **Writes:** `specs/services.md`

Classifies each requested service, then resolves a dependency-ordered list of BTP entitlements, subscriptions, and service instances for each subaccount.

## When to run it

After `btp-iac.accounts`, once the account topology exists.

## Inputs and outputs

| | |
|---|---|
| **Reads** | `specs/scenario.md` and `specs/landscape.md`, and `memory/global-account.md` and `memory/governance.md` if present. |
| **Writes** | `specs/services.md` — per subaccount: required entitlements (service + plan), subscriptions, and service instances with configuration, consumption type, and location metadata. |

## Behaviour

- Maps the scenario's needs onto concrete BTP services and plans for each subaccount in the landscape.
- For every service not already decided by governance, asks whether it is a service instance, an application subscription, or entitlement-only. SaaS applications default to subscriptions; technical services default to instances.
- For each service instance in a subaccount with Cloud Foundry enabled, asks whether it belongs on BTP or in a CF space. CF instances record their selected `cf_space`; if the landscape contains no CF space, the command asks you to add one with `btp-iac.accounts` before continuing.
- Records `consumption_type` for every service and records `location` (`btp` or `cf`) and `cf_space` for CF instances. Entitlement-only services receive only their entitlement assignment.
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

For the HR leave-request subaccount, `services` records the SAP HANA Cloud and XSUAA entitlements and their instances in `specs/services.md`. This example selects BTP for XSUAA and a CF space for HANA Cloud:

```markdown
## hr-leave-prod
### Entitlements
- hana-cloud / hana (SAP HANA Cloud)
- xsuaa / application
### Service instances (in order)
1. xsuaa (application) — consumption_type: instance, location: btp
2. hana-cloud (hana) — consumption_type: instance, location: cf, cf_space: hr-leave-apps
```

## Related

- Requires [`btp-iac.accounts`](accounts.md).
- Feeds [`btp-iac.security`](security.md) and [`btp-iac.tasks`](tasks.md).

Next: [`btp-iac.security`](security.md).
