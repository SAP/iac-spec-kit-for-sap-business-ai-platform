# `sap-iac.analyse`

!!! abstract "Summary"
    **Role:** Optional — enrichment. · **Reads:** `specs/scenario.md` + your code and/or live SAP BTP landscape · **Writes:** `specs/scenario.md`

Scans existing application source, Terraform code, or a live SAP BTP landscape (read-only) and enriches `specs/scenario.md` in place with the infrastructure facts it observes. Use it when you already have code — or a running subaccount — and want the scenario grounded in what actually exists, rather than in your description alone.

## When to run it

After `sap-iac.scenario` and before `sap-iac.accounts`, when application source, existing Terraform, or a reachable SAP BTP landscape is available. Skip it for a greenfield project described entirely by hand.

## Inputs and outputs

| | |
|---|---|
| **Reads** | `specs/scenario.md` (must exist) plus one or more of: your application code — descriptors such as `mta.yaml`, `package.json` (`@sap/cds`), `schema.cds`, `manifest.yml`, `xs-security.json`, `xs-app.json`; Terraform (`*.tf`, `*.tf.json`); and the **live SAP BTP platform** via an authenticated BTP CLI or BTP MCP integration (read/list only). |
| **Writes** | `specs/scenario.md`, under an `## Infrastructure Signals` heading. It rewrites only the content between its own `<!-- sap-iac:analyse:begin -->` and `<!-- sap-iac:analyse:end -->` markers (the heading sits just above the begin marker); everything else in the file is preserved. |
| **Asks** | Whether to analyse application source, Terraform, the live platform, or any combination; the path to the code; and, for the platform, an optional global-account/subaccount scope. |

## Behaviour

- Detects all descriptors it finds (it does not stop at the first) and applies a precedence order — MTA › CAP › Cloud Foundry — recording disagreements as conflicts.
- Treats Terraform as a co-equal input: it inspects every root `*.tf`/`*.tf.json` file (excluding `.terraform`, `.git`, and generated or vendored directories) without stopping at the first signal, follows local module `source` references up to a depth of 3, and records external (registry/git) module source and version while reporting that they were not inspected.
- Discovers a **live SAP BTP landscape** read-only when selected: the global account / directory / subaccount hierarchy, regions and subaccount configuration, entitlements and quotas, service plans, service instances and subscriptions, users / role collections / role assignments, trust configuration and destinations, and the relationships between them — all within what the authenticated principal's permissions allow. It uses only documented read/list commands and read/list BTP MCP tools; it never logs in or changes anything.
- Records each signal with five fields: evidence, observed value, confidence (`observed`, `inferred`, or `unresolved`), **source** (`platform`, `terraform`, or `app`), and downstream consumer.
- Organises findings into five subsections — **Observed services**, **Observed security**, **Terraform architecture**, **Platform landscape**, and **Unresolved mappings and conflicts** — and omits any subsection whose table has no rows.
- Reports incomplete, inaccessible, or ambiguous information as actionable diagnostics — missing permissions, unsupported resource types, unresolved Terraform inputs, inaccessible remote/platform state — rather than inventing values.
- Refuses a path that resolves outside the project root, stops when it finds no supported descriptor or `*.tf`/`*.tf.json` at the path, and — for the live platform — stops the platform branch (reporting that authentication is required) when no authenticated read/list route is available.
- Produces a completion report listing analysed and skipped files, live-platform coverage and inaccessible scopes, findings, redactions, and conflicts.

!!! warning "Secrets are never copied into the spec"
    `analyse` redacts credentials, certificates, tokens, client secrets, service-binding credentials, and trust/destination secrets — from code **and** from live-platform read output — recording only that a piece of configuration exists and its type. It never pastes raw command output into the spec, and never invents values: anything it cannot determine is marked `unresolved`.

!!! note "Out of scope"
    `analyse` never imports discovered resources into Terraform state or emits `terraform import` commands, never applies/modifies/deletes any BTP resource, and never auto-reconciles differences between the live platform and your Terraform — conflicts are recorded for you to resolve.

## Example

With the HR leave-request project's CAP source checked out alongside the specs:

```
/sap-iac.analyse
```

Choose to analyse the application source and give its path. `analyse` reads `package.json` and finds `@sap/cds`, reads `xs-security.json` for the XSUAA scopes, and writes its findings as tables between the markers in `specs/scenario.md`:

```markdown
## Infrastructure Signals
<!-- sap-iac:analyse:begin -->

### Observed services
| Signal | Evidence | Confidence | Source | Downstream consumer |
|---|---|---|---|---|
| SAP HANA Cloud | `@sap/cds` + `hana` in package.json | observed | app | sap-iac.services |

<!-- sap-iac:analyse:end -->
```

Choosing Terraform instead points `analyse` at your `*.tf`/`*.tf.json` files, where it treats the configuration as a graph across files rather than reading a single descriptor. From a subaccount configuration it records the observed account topology, the pinned provider versions, sensitive variables (whose values are redacted), and how modules are wired — noting local modules it followed against external modules it recorded but did not inspect:

```markdown
### Terraform architecture
| Signal | Evidence | Confidence | Source | Downstream consumer |
|---|---|---|---|---|
| Subaccount `dev` in region `eu10` | `btp_subaccount.dev` in `main.tf` | observed | terraform | sap-iac.accounts |
| Provider `SAP/btp` pinned to `~> 1.5` | `required_providers` in `providers.tf` | observed | terraform | sap-iac.design |
| Sensitive variable `idp_secret` (value redacted) | `variable "idp_secret"` (`sensitive = true`) in `variables.tf` | observed | terraform | sap-iac.security |
| Local module `./modules/entitlements` | `module "entitlements" source` in `main.tf` | observed | terraform | sap-iac.services |
| External module `terraform-sap/subaccount` v0.3.0 (not inspected) | `module "sa" source`/`version` in `main.tf` | unresolved | terraform | sap-iac.accounts |
```

Choosing the live platform points `analyse` at your authenticated BTP CLI / BTP MCP route (read-only). It discovers the account topology, entitlements, and role assignments its permissions allow, records structural topology under **Platform landscape**, and reports anything it could not read as a diagnostic rather than guessing:

```markdown
### Platform landscape
| Signal | Evidence | Confidence | Source | Downstream consumer |
|---|---|---|---|---|
| Subaccount `dev` (region `eu10`) under directory `engineering` | `btp list accounts/subaccount` | observed | platform | sap-iac.accounts |
| Trust configuration to IdP `acme-corp` (secret redacted) | `btp list security/trust` | observed | platform | sap-iac.security |
| Role collections not accessible (missing permission) | `btp list security/role-collection` → denied | unresolved | platform | sap-iac.security |
```

## Related
- Requires [`sap-iac.scenario`](scenario.md) to have run first.
- Continue to [`sap-iac.accounts`](accounts.md).
