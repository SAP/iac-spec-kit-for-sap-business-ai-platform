# `sap-iac.accounts`

!!! abstract "Summary"
    **Role:** Required. · **Reads:** `specs/scenario.md` (+ governance) · **Writes:** `specs/landscape.md`

Defines the BTP account topology — the subaccounts, regions, and directory groupings your scenario needs. Its output is the authoritative account structure that services, security, and generation all build on.

## When to run it

After `sap-iac.scenario` (and optional `sap-iac.analyse`), once the requirements are captured.

## Inputs and outputs

| | |
|---|---|
| **Reads** | `specs/scenario.md`, and `memory/global-account.md` and `memory/governance.md` if present. |
| **Writes** | `specs/landscape.md` — the global account reference, a `## Directories` section (each directory's managed/unmanaged posture, features, and subdomain when applicable), and one entry per subaccount (name, region, description, subdomain, directory assignment, shared metadata, `usage`, and `beta_enabled`) across your environment tiers, plus each subaccount's runtime environments (Cloud Foundry org and optional space names, Kyma environment names). |

## Behaviour

- Derives the account topology from the scenario's environment and deployment requirements.
- Produces the authoritative structure referenced by `services`, `security`, and `generate`.
- Determines `usage` and `beta_enabled` from exact governance tier classifications when provided. Otherwise it infers them from `prod`/`production`, `dev`/`development`, and `test`/`staging`/`qa` tier-name tokens, asks once for any unmatched tiers, and confirms the complete table before writing.
- Prefers the `sap-docs` MCP server over web fetches for SAP documentation.

The command is interactive: when the scenario and governance leave an input underspecified, it asks targeted questions before writing `specs/landscape.md` rather than guessing.

### Directory setup

When the governance `## Naming → Directory structure` is `hierarchical`, `accounts` resolves three directory aspects — directory management, subaccount-to-directory assignment, and shared subaccounts. If `Directory structure` is `none`/`flat`, no directories are created and every subaccount is parented to the global account; if it is **absent**, `accounts` first asks whether directories should be used. For each aspect it applies the matching `## Account Setup` governance field (`Directory management`, `Directory features`, `Subaccount directory assignment`, `Shared subaccounts`) when present, and otherwise asks you:

- **Managed vs. unmanaged** — unmanaged directories are plain folders (`btp_directory` DEFAULT feature). Managed directories are offered as two fixed feature sets: **Entitlement** (`ENTITLEMENTS`) or **Entitlement and authorization** (`ENTITLEMENTS` + `AUTHORIZATIONS`, which needs a directory subdomain); `DEFAULT` is always included.
- **Assignment** — which directory each subaccount sits under (or directly under the global account).
- **Shared subaccounts** — subaccounts not tied to a single stage, used as a common home for services shared across multiple subaccounts; recorded with a descriptive `Shared` flag, scope, and owner.

`accounts` records in `specs/landscape.md` only the decisions governance did not already define — governed facts stay authoritative in `memory/governance.md` and are not duplicated.

### Runtime environments

For each subaccount, `accounts` determines whether to create Cloud Foundry, Kyma, both, or no runtime environment, respecting any decision already implied by the scenario or governance and asking the user when it is not. For each Cloud Foundry environment it collects the organization name and asks whether spaces are needed, collecting their concrete names if so; for each Kyma environment it collects the environment name. These are recorded in `specs/landscape.md`.

### BTP platform validation

Using only read/list BTP CLI commands and BTP MCP lookups, `accounts` validates topology regions and any explicitly named service offering, subscription, or service-plan pair against the resolved global-account subdomain (targeted with `btp target --global-account <subdomain>` from `memory/global-account.md`). It uses `btp list accounts/available-region` for regions, ignoring all `NEO` region entries. This stays within the read/list-only BTP safety boundary — the command never creates, updates, or otherwise mutates BTP state.

!!! note "Governance enforcement"
    When a governance file is present, `accounts` validates every decision and **stops on a violation** unless `- Override: true` is set. It enforces: subaccount regions against the Allowed/Forbidden lists, subaccount names against the naming pattern, environment tiers against the defined tiers, selected runtime environment types against the allowed environment types, and Cloud Foundry org/space and Kyma environment names against their naming patterns. A preferred-provider mismatch (configured provider vs. the region's returned provider metadata) is a non-blocking warning. Without a governance file it proceeds without enforcement and says so.

## Example

```
/sap-iac.accounts
```

For the single-environment HR leave-request scenario, `accounts` writes one subaccount into `specs/landscape.md`:

```markdown
## Subaccounts
- Name: hr-leave-prod
  Region: eu10
  Subdomain: hr-leave-prod
  Description: HR leave-request application (single environment)
  Runtime environments:
    - Cloud Foundry:
        Org: hr-leave-prod
        Spaces: [prod]
```

Had governance restricted regions to `eu10`/`eu20`, a scenario asking for `us10` would have stopped here with a region violation.

## Related

- Requires [`sap-iac.scenario`](scenario.md).
- Feeds [`sap-iac.services`](services.md), [`sap-iac.security`](security.md), and [`sap-iac.generate`](generate.md).
