---
name: sap-iac-analyse
description: Analyses application source code, Terraform code, or a live SAP BTP landscape (read-only) to extract infrastructure signals, then enriches specs/scenario.md.
license: Apache-2.0
metadata:
  author: SAP
  version: "1.2"
---

# BTP IaC — Analyse

Analyses code — or a live SAP BTP landscape, read-only — to extract concrete infrastructure signals and enriches `specs/scenario.md` in place. There are three input sources: **application source code**, **Terraform code**, and the **live SAP BTP platform** (queried through an authenticated BTP CLI or BTP MCP integration). Findings are recorded as **observed facts with evidence, a confidence level, and a source** — not as approved desired state. Downstream skills consume reviewed requirements, so never present a raw implementation detail as intent.

## Platform Name Normalization

The following names all refer to the same platform and are semantically equivalent for all natural-language interpretation in this skill:
- **SAP Business Technology Platform** (and "Business Technology Platform")
- **SAP BTP** (and "BTP" used as a product name in prose)
- **SAP Business AI Platform** (and "Business AI Platform")
- **SAP BAIP** (and "BAIP")

Treat any of these aliases as identical when interpreting user intent. This normalization applies only to natural-language prose. Technical identifiers remain untouched: BTP CLI command tokens (`btp list`, `btp target`), Terraform provider names (`btp`, `hashicorp/btp`), resource type prefixes (`btp_subaccount`, `btp_service_instance`), region codes, and API paths.

## Preconditions

Before analysing, check these and report (do not silently proceed) if any hold:

- **`specs/scenario.md` missing** — tell the user to run `/sap-iac.scenario` first; do not create the file from scratch here.
- **Path outside the project root** — refuse a path that resolves outside the current project; do not read it.
- **No supported input at the path** — if no descriptor (application source) or `*.tf`/`*.tf.json` (Terraform) is found, report that and stop.
- **No authenticated platform route (live platform only)** — before discovering the live platform, confirm an authenticated read/list route exists (see *BTP platform validation* below). If none is available, report that authentication is required and stop the platform branch. **Never** run a login, config, or any other state-changing command to obtain one — authentication is the user's responsibility.

## Workflow

**Phase 1 — Ask what to analyse**: Ask whether to analyse (a) **application source code**, (b) **Terraform code**, (c) the **live SAP BTP platform**, or any **combination** of these. Do not proceed without a clear answer.

**Phase 2 — Locate the inputs**:
- *Application source / Terraform* — ask for the path if not provided. Confirm it resolves inside the project root before reading.
- *Live platform* — no path is required. Confirm an authenticated read/list route exists (per the precondition above) and ask whether discovery should be scoped to a specific global account and/or subaccount before issuing lookups.

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

*Live SAP BTP platform* — discover the landscape using **only read/list lookups** (see *BTP operation safety boundary*). Scope to the user's chosen global account/subaccount when given. Discover, to the extent the authenticated principal's permissions allow, and **do not stop at the first result**:

- **account structure** — global account, directories, subaccounts, and their hierarchy
- **subaccount configuration** — region, environment, labels, and other supported account-level settings
- **entitlements & quotas** — entitled service offerings, plans, and assigned quotas
- **services** — service instances, subscriptions, and related configuration where readable
- **security** — users, role collections, and role collection assignments where permissions allow
- **connectivity & trust** — trust configurations, destinations, and other supported account-level settings
- **relationships** — dependencies and references between discovered resources (e.g. a subscription that consumes an entitlement, a role assignment scoped to a subaccount)

When a read/list lookup is denied, unsupported, or returns a resource type you do not recognise, record a diagnostic (what was inaccessible and the implicated permission, or the unsupported type) and mark the affected items `unresolved` — then continue with what you can read. Never log in, target outside the chosen account, or run any non-read/list command to widen access.

**Never evaluate unknown or variable-driven expressions, and never invent a value** where configuration is indeterminate — record it as `unresolved`.

**Phase 4 — Record findings**: For every signal capture five things:

| Field | Meaning |
|---|---|
| Evidence | file and relevant location, or — for the platform — the read/list lookup and resource identifier |
| Observed value | the literal value found (or `unresolved`) |
| Confidence | `observed` (read directly), `inferred` (derived), or `unresolved` (indeterminate/variable-driven/absent) |
| Source | provenance of the finding — `platform`, `terraform`, or `app` |
| Downstream consumer | which command uses it — `sap-iac.accounts`, `sap-iac.services`, `sap-iac.security`, or `sap-iac.design` |

Record provenance explicitly via **Source**; never leave it to be inferred from context. When the same resource is seen from more than one source, record both findings with their distinct sources rather than merging them. When sources **disagree** about a resource (e.g. the platform and Terraform report different regions), record the discrepancy under **Unresolved mappings and conflicts** with both observed values and their sources — do not silently prefer one. Reconciling platform against Terraform is out of scope for this command.

**Phase 5 — Redact secrets**: Never write secrets into `specs/scenario.md` or the completion report. For backend settings, provider credentials, destination credentials, certificates, tokens, `sensitive` variable defaults, or — from live-platform discovery — service-binding credentials, client secrets, and trust/identity-provider secrets, redact the value and report only that the configuration exists and its type. Never paste raw read/list command output into the spec or the report.

**Phase 6 — Enrich `specs/scenario.md`**: Write findings **only** between the paired markers below, under the `## Infrastructure Signals` heading. On re-runs, replace only the content between the markers and preserve everything outside them. **Never rewrite the whole file.** Render each subsection as a markdown table carrying an explicit **Source** column; omit a table with no rows. Platform-discovered services and roles go under **Observed services** / **Observed security** (distinguished by Source); structural topology that is neither a service nor a security finding — the account hierarchy, regions, subaccount config, trust, destinations — goes under **Platform landscape**.

```md
## Infrastructure Signals
<!-- sap-iac:analyse:begin -->

### Observed services
| Signal | Evidence | Confidence | Source | Downstream consumer |
|---|---|---|---|---|

### Observed security
| Signal | Evidence | Confidence | Source | Downstream consumer |
|---|---|---|---|---|

### Terraform architecture
| Signal | Evidence | Confidence | Source | Downstream consumer |
|---|---|---|---|---|

### Platform landscape
| Signal | Evidence | Confidence | Source | Downstream consumer |
|---|---|---|---|---|

### Unresolved mappings and conflicts
| Item | Sources | Evidence | Required decision |
|---|---|---|---|

<!-- sap-iac:analyse:end -->
```

If the file already contains an older single `<!-- auto-generated by /sap-iac.analyse -->` marker from a prior version, replace that whole legacy section with the paired-marker block.

**Phase 7 — Completion report**: Report analysed files, skipped files (with reason — excluded dir, external module, unreadable), live-platform coverage (which accounts/subaccounts and resource categories were discovered) and resources or scopes that could not be accessed (with the implicated permission), findings written, redactions performed, unresolved mappings and conflicts, and the recommended next command.

## Tool preferences

Before using `WebFetch` to look up SAP documentation, check if the `sap-docs` MCP server is available (tools prefixed `mcp__sap-docs__*`). If yes, use it. If not, fall back to `WebFetch`.

## BTP platform validation

### BTP operation safety boundary

The prohibition on state-changing CLI commands excludes the permitted target prelude.

When invoking the BTP CLI or any BTP MCP tool, perform **only read or list retrievals**. For the BTP CLI, invoke only documented read/list commands (for example, `btp list ...`); for BTP MCP, invoke only a tool explicitly documented as a read/list lookup. `btp target --global-account <subdomain>` is the sole permitted account-selection prelude and may be used only immediately before those read/list CLI commands. Never invoke, suggest, or approve a BTP operation that creates, updates, deletes, assigns, unassigns, enables, disables, or otherwise mutates BTP state — even when requested by the user. Do not run login, config, profile, or any other state-changing CLI command.

This boundary applies equally to **live-platform discovery** (Phase 3) and to validating an observed value: discovery is performed with the same read/list-only lookups and never widens its own access.

When this workflow needs a platform route — to discover the live landscape, or to validate an observed BTP value — read `<project-root>/.sap-iac/platform-validation.md`. Prefer its recorded CLI route, then this agent's recorded BTP MCP route. If no route is recorded, report that an authenticated read/list route is required (for discovery, stop the platform branch per the precondition; for validation, retain user input without blocking); if a recorded route fails, ask the user to resolve the setup before relying on it. Do not turn observed findings into desired state merely because a platform route is available.
