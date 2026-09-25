# `sap-iac.scenario`

!!! abstract "Summary"
    **Role:** Required — the starting point. · **Reads:** your description · **Writes:** `specs/scenario.md`

Translates a plain-language description of your application into a structured set of BTP infrastructure requirements. This is the entry point of the workflow: everything downstream builds on the scenario it produces.

## When to run it

First (after the optional [`sap-iac.govern`](govern.md)). You need only a short description of what you are building.

## Inputs and outputs

| | |
|---|---|
| **Reads** | Your plain-language description. It also reads `.sap-iac/platform-validation.md` and `memory/global-account.md`, and `memory/governance.md` when it exists, to ground and constrain its follow-up questions — but no `specs/` files are required. |
| **Writes** | `specs/scenario.md`. |
| **Asks** | Up to three clarifying questions — typically about unresolved runtime and Cloud Foundry sizing, destinations, and setup structure. |

## Behaviour

- Takes your description of the application — runtime, data, users, environments — and structures it into requirements.
- Asks up to three focused follow-up questions to fill obvious gaps; when governance already determines a choice (for example, a single permitted runtime), it does not ask.
- Prefers the `sap-docs` MCP server over web fetches when it needs to consult SAP documentation.
- Validates any regions, service offerings, subscriptions, or plan pairs named in the description; if governance sets a preferred infrastructure provider other than `none`, it emits a non-blocking provider-mismatch warning (naming the region and both providers) while retaining the region.

## Example

Describe the application in one paragraph:

```
/sap-iac.scenario An internal HR leave-request app built with CAP (Node.js) on
Cloud Foundry. It stores leave requests in SAP HANA Cloud and authenticates
employees via XSUAA. Single environment, one subaccount, region eu10.
```

After answering any follow-up questions, `specs/scenario.md` captures the structured requirements — the runtime (CAP on Cloud Foundry), the persistence (SAP HANA Cloud), the authentication (XSUAA), and the single-environment, single-subaccount topology in region eu10.

!!! tip "Need a starting description?"
    The [scenario examples](../scenarios.md) page has ready-made descriptions at every complexity level — copy one and adapt it.

## Related

- Optional enrichment next: [`sap-iac.analyse`](analyse.md).
- Otherwise, continue to [`sap-iac.accounts`](accounts.md).
