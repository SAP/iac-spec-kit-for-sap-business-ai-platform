# `sap-iac.analyse`

!!! abstract "Summary"
    **Role:** Optional — enrichment. · **Reads:** `specs/scenario.md` + your code · **Writes:** `specs/scenario.md`

Scans existing application source or Terraform code and enriches `specs/scenario.md` in place with the infrastructure facts it observes. Use it when you already have code and want the scenario grounded in what the code actually declares, rather than in your description alone.

## When to run it

After `sap-iac.scenario` and before `sap-iac.accounts`, when application source or existing Terraform is available. Skip it for a greenfield project described entirely by hand.

## Inputs and outputs

| | |
|---|---|
| **Reads** | `specs/scenario.md` (must exist) plus your code — descriptors such as `mta.yaml`, `package.json` (`@sap/cds`), `schema.cds`, `manifest.yml`, `xs-security.json`, `xs-app.json`, and/or Terraform (`*.tf`, `*.tf.json`). |
| **Writes** | `specs/scenario.md`, under an `## Infrastructure Signals` heading. It rewrites only the content between its own `<!-- sap-iac:analyse:begin -->` and `<!-- sap-iac:analyse:end -->` markers (the heading sits just above the begin marker); everything else in the file is preserved. |
| **Asks** | Whether to analyse application source, Terraform, or both, and the path to the code. |

## Behaviour

- Detects all descriptors it finds (it does not stop at the first) and applies a precedence order — MTA › CAP › Cloud Foundry — recording disagreements as conflicts.
- Treats Terraform as a co-equal input: it inspects every root `*.tf`/`*.tf.json` file (excluding `.terraform`, `.git`, and generated or vendored directories) without stopping at the first signal, follows local module `source` references up to a depth of 3, and records external (registry/git) module source and version while reporting that they were not inspected.
- Records each signal with four fields: evidence, observed value, confidence (`observed`, `inferred`, or `unresolved`), and downstream consumer.
- Organises findings into four subsections — **Observed services**, **Observed security**, **Terraform architecture**, and **Unresolved mappings and conflicts** — and omits any subsection whose table has no rows.
- Refuses a path that resolves outside the project root, and stops when it finds no supported descriptor or `*.tf`/`*.tf.json` at the path.
- Produces a completion report listing analysed and skipped files, findings, redactions, and conflicts.

!!! warning "Secrets are never copied into the spec"
    `analyse` redacts credentials, certificates, tokens, and sensitive values — it records only that a piece of configuration exists and its type. It never invents values: anything it cannot determine is marked `unresolved`.

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
| Signal | Evidence | Confidence | Downstream consumer |
|---|---|---|---|
| SAP HANA Cloud | `@sap/cds` + `hana` in package.json | observed | sap-iac.services |

<!-- sap-iac:analyse:end -->
```

Choosing Terraform instead points `analyse` at your `*.tf`/`*.tf.json` files, where it treats the configuration as a graph across files rather than reading a single descriptor. From a subaccount configuration it records the observed account topology, the pinned provider versions, sensitive variables (whose values are redacted), and how modules are wired — noting local modules it followed against external modules it recorded but did not inspect:

```markdown
### Terraform architecture
| Signal | Evidence | Confidence | Downstream consumer |
|---|---|---|---|
| Subaccount `dev` in region `eu10` | `btp_subaccount.dev` in `main.tf` | observed | sap-iac.accounts |
| Provider `SAP/btp` pinned to `~> 1.5` | `required_providers` in `providers.tf` | observed | sap-iac.design |
| Sensitive variable `idp_secret` (value redacted) | `variable "idp_secret"` (`sensitive = true`) in `variables.tf` | observed | sap-iac.security |
| Local module `./modules/entitlements` | `module "entitlements" source` in `main.tf` | observed | sap-iac.services |
| External module `terraform-sap/subaccount` v0.3.0 (not inspected) | `module "sa" source`/`version` in `main.tf` | unresolved | sap-iac.accounts |
```

## Related
- Requires [`sap-iac.scenario`](scenario.md) to have run first.
- Continue to [`sap-iac.accounts`](accounts.md).
