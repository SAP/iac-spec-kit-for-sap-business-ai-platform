---
name: btp-iac-analyse
description: Analyses application source code or Terraform code to extract infrastructure signals, then enriches specs/scenario.md.
license: Apache-2.0
metadata:
  author: SAP
  version: "1.1"
---

# BTP IaC — Analyse

Analyses code to extract concrete infrastructure signals and enriches `specs/scenario.md` in place. Findings are recorded as **observed facts with evidence and a confidence level** — not as approved desired state. Downstream skills consume reviewed requirements, so never present a raw implementation detail as intent.

## Preconditions

Before analysing, check these and report (do not silently proceed) if any hold:

- **`specs/scenario.md` missing** — tell the user to run `/btp-iac.scenario` first; do not create the file from scratch here.
- **Path outside the project root** — refuse a path that resolves outside the current project; do not read it.
- **No supported input at the path** — if no descriptor (application source) or `*.tf`/`*.tf.json` (Terraform) is found, report that and stop.

## Workflow

**Phase 1 — Ask what to analyse**: Ask whether to analyse (a) **application source code**, (b) **Terraform code**, or (c) **both**. Do not proceed without a clear answer.

**Phase 2 — Locate the code**: Ask for the path if not provided. Confirm it resolves inside the project root before reading.

**Phase 3 — Detect and read** (existence checks before full reads):

*Application source code* — detect **all** descriptors present; do not stop at the first match. If more than one descriptor or multiple deployable apps are found (e.g. a monorepo with both `mta.yaml` and CAP config), analyse each, apply this precedence for overlapping facts — **MTA › CAP › CF** — and record disagreements as conflicts rather than discarding later files.

- **MTA** (`mta.yaml` present): module/resource relationships, service offerings/plans, `requires`, memory/disk/instances, build parameters. Also read `xs-security.json` and `xs-app.json` if present.
- **CAP** (`package.json` contains `@sap/cds`): `cds.requires`, database kind, messaging, destinations, XSUAA, multitenancy, HANA. Read `schema.cds` (entity names only) and `xs-security.json` if present.
- **CF** (`manifest.yml` present): disk, instances, routes, buildpack, stack, health checks, and any unresolved variable placeholders. For a service binding that gives only an instance name, mark its offering/plan `unresolved` — do not infer them.
- **`xs-security.json`** (any format): scopes, role templates, role collections, attribute references, authorities, tenant mode.

*Terraform code* — inspect **all** root `*.tf` and `*.tf.json` files; exclude `.terraform`, `.git`, and generated/vendored directories. Terraform is a graph across files (providers, variables, locals, resources, modules, instances are commonly split), so **do not stop after initial signals**. Follow local module `source` references up to a depth of **3**; for external module sources (registry/git), record source and version and report that they were not inspected. Extract:

- **subaccounts, directories** — regions, hierarchy, labels
- **entitlements & subscriptions** — service offerings/plans, service instances
- **security** — trust configurations, role collections, assignments
- **providers** — `required_providers` (versions) and provider aliases
- **graph** — `data`, `locals`, `outputs`, dependency edges, `for_each`, `count`, `depends_on`
- **variables** — type, `description`, `validation`, `sensitive`, and whether there is no default
- **modules** — local module interfaces (inputs/outputs); external module source/version

**Never evaluate unknown or variable-driven expressions, and never invent a value** where configuration is indeterminate — record it as `unresolved`.

**Phase 4 — Record findings**: For every signal capture four things:

| Field | Meaning |
|---|---|
| Evidence | file and relevant location |
| Observed value | the literal value found (or `unresolved`) |
| Confidence | `observed` (read directly), `inferred` (derived), or `unresolved` (indeterminate/variable-driven/absent) |
| Downstream consumer | which command uses it — `btp-iac.accounts`, `btp-iac.services`, `btp-iac.security`, or `btp-iac.design` |

**Phase 5 — Redact secrets**: Never write secrets into `specs/scenario.md`. For backend settings, provider credentials, destination credentials, certificates, tokens, or `sensitive` variable defaults, redact the value and report only that the configuration exists and its type.

**Phase 6 — Enrich `specs/scenario.md`**: Write findings **only** between the paired markers below, under the `## Infrastructure Signals` heading. On re-runs, replace only the content between the markers and preserve everything outside them. **Never rewrite the whole file.** Render each subsection as a markdown table; omit a table with no rows.

```md
## Infrastructure Signals
<!-- btp-iac:analyse:begin -->

### Observed services
| Signal | Evidence | Confidence | Downstream consumer |
|---|---|---|---|

### Observed security and connectivity
| Signal | Evidence | Confidence | Downstream consumer |
|---|---|---|---|

### Terraform architecture
| Signal | Evidence | Confidence | Downstream consumer |
|---|---|---|---|

### Unresolved mappings and conflicts
| Item | Evidence | Required decision |
|---|---|---|

<!-- btp-iac:analyse:end -->
```

If the file already contains an older single `<!-- auto-generated by /btp-iac.analyse -->` marker from a prior version, replace that whole legacy section with the paired-marker block.

**Phase 7 — Completion report**: Report analysed files, skipped files (with reason — excluded dir, external module, unreadable), findings written, redactions performed, unresolved mappings and conflicts, and the recommended next command.

## Tool preferences

Before using `WebFetch` to look up SAP documentation, check if the `sap-docs` MCP server is available (tools prefixed `mcp__sap-docs__*`). If yes, use it. If not, fall back to `WebFetch`.
