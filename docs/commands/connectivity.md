# `btp-iac.connectivity`

!!! abstract "Summary"
    **Role:** Optional. · **Reads:** `specs/scenario.md` (if present), `.btp-iac/platform-validation.md` · **Writes:** `specs/connectivity.md`

Elicits and structures BTP destination and destination-certificate requirements. Run it when your scenario needs to reach external or on-premise systems — an internet API, an on-premise backend through Cloud Connector, an RFC/ABAP system, LDAP, mail, or a raw TCP host.

## When to run it

After `btp-iac.security` and before `btp-iac.tasks`. It is optional: skip it for a self-contained application that talks to no external systems.

## Inputs and outputs

| | |
|---|---|
| **Reads** | `specs/scenario.md` if present, reusing any destinations the scenario already captured. It also reads `.btp-iac/platform-validation.md` for a live BTP route when one is available. |
| **Writes** | `specs/connectivity.md` — a Destinations section (one entry per destination) and an optional Certificates section. |
| **Asks** | How many destinations are needed and, for each, which connection pattern fits; whether any destination certificates are needed; and the scope (subaccount- or service-instance-level) of each. |

## Behaviour

- Offers canonical destination patterns as hints — HTTP (`NoAuthentication`, `BasicAuthentication` via Cloud Connector, `OAuth2ClientCredentials`), RFC, LDAP, MAIL, and TCP — and collects only the required fields for the chosen pattern.
- Optionally collects `btp_subaccount_destination_certificate` requirements (`.pem`, `.p12`, `.jks`, `.pfx`), at subaccount or service-instance scope.
- Prefers the `sap-docs` and `terraform` MCP servers over web fetches.
- Any BTP operation it performs is read/list only — it never creates, updates, deletes, or otherwise mutates BTP state.

!!! warning "Credentials are never written to the spec"
    Passwords, client secrets, tokens, and certificate content are recorded as `<PLACEHOLDER>` with a note to supply them at apply time — never committed to the repository.

## Example

```
/btp-iac.connectivity
```

For an order-sync service reaching an on-premise S/4HANA backend and an external shipping API, `connectivity` records two destinations in `specs/connectivity.md`:

```markdown
## Destinations

### shipping-api
- **Type**: HTTP
- **ProxyType**: Internet
- **Authentication**: OAuth2ClientCredentials
- **URL / Address**: https://shipping.example.com
- **Credentials**: ⚠ PLACEHOLDER — supply at apply time, do not commit to the repository

### s4-orders
- **Type**: HTTP
- **ProxyType**: OnPremise
- **Authentication**: BasicAuthentication
- **CloudConnectorLocationId**: s4-onprem
- **Credentials**: ⚠ PLACEHOLDER — supply at apply time, do not commit to the repository

## Certificates

### s4-truststore.pem
- **Subaccount**: order-sync-prod
- **Service Instance** (if applicable): —
- **File**: ⚠ PLACEHOLDER — supply at apply time
```

`tasks` then appends a task per destination and certificate, and `generate` produces the `btp_subaccount_destination_generic` and `btp_subaccount_destination_certificate` resources.

## Related

- Runs after [`btp-iac.security`](security.md).
- Feeds [`btp-iac.tasks`](tasks.md) when present.
- Informs [`btp-iac.generate`](generate.md), which produces the `btp_subaccount_destination_generic` and `btp_subaccount_destination_certificate` resources.
