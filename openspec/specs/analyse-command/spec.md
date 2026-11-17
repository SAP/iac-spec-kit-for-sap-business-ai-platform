# Analyse Command Capability

## Purpose

Defines the behaviour of `/sap-iac.analyse`: reading application source code or Terraform configuration to extract concrete infrastructure signals as evidenced, confidence-rated facts, and enriching `specs/scenario.md` in place without leaking secrets or overwriting user content.
## Requirements
### Requirement: select analysis input
The command SHALL ask the user whether to analyse existing application source code, Terraform configuration, or both, and SHALL NOT proceed until it has a clear answer.

#### Scenario: input type chosen
- **WHEN** the command starts
- **THEN** it asks whether to analyse application source code, Terraform configuration, or both, and waits for a clear answer before proceeding

### Requirement: read application source code
The command SHALL ask the user for the path to their application source code before proceeding, then read the files at that path.

#### Scenario: path provided
- **WHEN** the user supplies a source code path
- **THEN** the command reads the relevant files at that path

### Requirement: analyse Terraform configuration
The command SHALL support analysing Terraform configuration. It SHALL inspect all root `*.tf` and `*.tf.json` files, excluding `.terraform`, `.git`, and generated or vendored directories, and SHALL NOT stop after finding initial signals. The command SHALL follow local module `source` references up to a bounded recursion depth and SHALL explicitly report external modules that could not be inspected. It SHALL extract BTP-domain findings including subaccounts, regions, hierarchy, and labels; entitlements, subscriptions, service offerings/plans, and service instances; trust configuration, role collections and role collection assignments; provider aliases and `required_providers`; and `data`, `locals`, `outputs`, dependency edges, `for_each`, `count`, and `depends_on`. For input variables it SHALL record description, validation, `sensitive`, and whether a value has no default, and SHALL record local module interfaces and external module source/version.

#### Scenario: all Terraform files inspected
- **WHEN** the user selects Terraform analysis for a configuration directory
- **THEN** the command inspects every root `*.tf` and `*.tf.json` file, excluding `.terraform`, `.git`, and generated or vendored directories, without stopping at the first signal

#### Scenario: local modules followed, external modules reported
- **WHEN** the configuration references local and external modules
- **THEN** the command follows local module sources up to a bounded depth and explicitly reports external modules it could not inspect

### Requirement: extract infrastructure signals
The command SHALL extract concrete infrastructure signals the user cannot be expected to know, and SHALL record each finding as a discrete fact rather than an assumed requirement. For every finding the command SHALL capture: the evidence source (file and relevant location), the observed value, a confidence level of `observed`, `inferred`, or `unresolved`, and the downstream command that consumes it. The command SHALL NOT evaluate unknown or variable-driven expressions and SHALL NOT invent values where configuration is absent or indeterminate; such cases SHALL be recorded as `unresolved`. From application source the command SHALL cover service dependencies, memory requirements, role definitions, and build artifact details, and SHALL additionally extract MTA module/resource relationships, service offerings/plans, `requires`, and build parameters; CAP `cds.requires`, database kind, messaging, destinations, XSUAA, and multitenancy; CF disk, instances, routes, buildpack, stack, and health checks; and `xs-security.json` scopes, role templates, role collections, attributes, authorities, and tenant mode.

#### Scenario: findings carry evidence and confidence
- **WHEN** the command records an infrastructure signal
- **THEN** the finding includes its evidence source, observed value, a confidence level of `observed`, `inferred`, or `unresolved`, and the downstream consumer

#### Scenario: indeterminate values are not invented
- **WHEN** a value is variable-driven, absent, or otherwise indeterminate
- **THEN** the command records it as `unresolved` and does not evaluate the expression or invent a value

#### Scenario: CF binding with only an instance name
- **WHEN** a Cloud Foundry service binding references only an instance name with no offering or plan
- **THEN** the command marks the offering and plan as `unresolved` rather than inferring them

### Requirement: support MTA, CAP, and CF app formats
The command SHALL support three BTP application formats and SHALL detect all descriptors present rather than stopping at the first match. When more than one descriptor or multiple deployable applications are present (for example a monorepo containing both `mta.yaml` and CAP configuration), the command SHALL apply a defined precedence, report conflicting observations, and SHALL NOT silently discard later files.

#### Scenario: MTA app
- **WHEN** the source contains `mta.yaml`, `xs-security.json`, or `xs-app.json`
- **THEN** the command extracts signals from those files

#### Scenario: CAP app
- **WHEN** the source contains `package.json`, `schema.cds`, or `xs-security.json`
- **THEN** the command extracts signals from those files

#### Scenario: CF app
- **WHEN** the source contains `manifest.yml`
- **THEN** the command extracts signals from that file

#### Scenario: multiple descriptors present
- **WHEN** more than one supported descriptor or multiple deployable applications are found
- **THEN** the command detects all of them, reports conflicting observations, and does not discard later files

### Requirement: redact secrets
The command SHALL NOT write secrets into `specs/scenario.md`. When it encounters backend settings, provider credentials, destination credentials, certificates, tokens, or `sensitive` variable defaults, it SHALL redact the value and report only the existence and type of the configuration.

#### Scenario: sensitive value encountered
- **WHEN** the command encounters a credential, certificate, token, or `sensitive` value
- **THEN** it redacts the value and records only that the configuration exists and its type

### Requirement: validate preconditions
The command SHALL define its behaviour for boundary conditions before and during analysis. It SHALL handle a missing `specs/scenario.md`, a supplied path outside the project root, the absence of any supported descriptor or configuration, disagreement between source code and Terraform, and local modules that are unavailable or external modules that cannot be inspected — reporting each condition rather than proceeding on an assumption.

#### Scenario: path outside project root
- **WHEN** the supplied path is outside the project root
- **THEN** the command reports the condition and does not proceed

#### Scenario: no supported input found
- **WHEN** no supported descriptor or Terraform configuration is found at the path
- **THEN** the command reports that no supported input was found and stops

#### Scenario: source and Terraform disagree
- **WHEN** application source and Terraform configuration provide conflicting observations
- **THEN** the command records the conflict as an unresolved item rather than choosing one silently

### Requirement: enrich scenario file in place
The command SHALL enrich `specs/scenario.md` in place with findings from the analysis, writing them only between the paired markers `<!-- sap-iac:analyse:begin -->` and `<!-- sap-iac:analyse:end -->`. On re-runs the command SHALL replace only the content enclosed by those markers and SHALL preserve all content outside them. The command SHALL NOT perform a whole-file rewrite. The generated section SHALL organise findings under "Observed services", "Observed security", "Terraform architecture", "Platform landscape", and "Unresolved mappings and conflicts", and every findings table SHALL carry an explicit source column distinguishing `platform`, `terraform`, and `app` findings. A subsection whose table has no rows SHALL be omitted.

#### Scenario: scenario file enriched between markers
- **WHEN** the analysis is complete
- **THEN** the extracted findings are written between `<!-- sap-iac:analyse:begin -->` and `<!-- sap-iac:analyse:end -->`, organised into the observed-services, observed-security, Terraform-architecture, platform-landscape, and unresolved subsections, each findings table carrying an explicit source column

#### Scenario: re-run preserves surrounding content
- **WHEN** the command runs again on a file that already contains the paired markers
- **THEN** only the content between the markers is replaced and all content outside them is preserved

#### Scenario: platform-only landscape recorded
- **WHEN** live-platform discovery produces structural topology that is neither a service nor a security finding (for example the account hierarchy, regions, or destinations)
- **THEN** the command records it under the "Platform landscape" subsection with its source marked `platform`

### Requirement: report completion
The command SHALL produce a completion report identifying the files analysed and skipped, the findings written, the redactions performed, the unresolved mappings, and the recommended next command.

#### Scenario: completion report produced
- **WHEN** the analysis finishes
- **THEN** the command reports analysed and skipped files, findings written, redactions performed, unresolved mappings, and the recommended next command

### Requirement: discover the live SAP BTP platform
The command SHALL support discovering an existing SAP BTP landscape through an authenticated BTP CLI or BTP MCP integration, using only read/list lookups. Where the authenticated principal's permissions allow, it SHALL discover and record: the global account, directory, and subaccount structure; regions, environments, and subaccount configuration; entitlements, quotas, and service plans; service instances, subscriptions, and related configuration; users, role collections, and role collection assignments; trust configuration, destinations, and other supported account-level settings; and the relationships and dependencies between these resources. The command SHALL record each discovered item in the same findings model used for static sources — evidence, observed value, confidence (`observed`, `inferred`, or `unresolved`), and downstream consumer — and SHALL NOT invent values for resources it cannot read.

#### Scenario: landscape discovered within permissions
- **WHEN** the user selects live-platform analysis and an authenticated read/list route is available
- **THEN** the command discovers the account hierarchy, subaccount configuration, entitlements and plans, service instances and subscriptions, role collections and assignments, trust configuration, and destinations that its permissions allow, and records each as a finding with evidence, observed value, confidence, and downstream consumer

#### Scenario: relationships captured
- **WHEN** discovered resources depend on or reference one another (for example a subscription that consumes an entitlement, or a role assignment scoped to a subaccount)
- **THEN** the command records the relationship between the discovered resources

#### Scenario: discovery uses read/list only
- **WHEN** the command performs live-platform discovery
- **THEN** it invokes only documented read/list CLI commands and read/list BTP MCP tools, using `btp target --global-account <subdomain>` only as the account-selection prelude, and never a command that creates, updates, deletes, assigns, enables, disables, logs in, or otherwise mutates BTP state

### Requirement: tag finding provenance
The command SHALL tag every recorded finding with its provenance — whether it was observed from the live SAP BTP platform, from Terraform configuration, or from application source — so that downstream commands can distinguish the origin of each item. Provenance SHALL be represented explicitly (a `Source` value on each finding) rather than inferred from context.

#### Scenario: provenance recorded per finding
- **WHEN** the command records any finding
- **THEN** the finding carries an explicit source of `platform`, `terraform`, or `app`

#### Scenario: same resource seen from two sources
- **WHEN** the same resource is observed from both the live platform and Terraform
- **THEN** the command records both findings with their distinct sources rather than merging them into one unsourced finding

### Requirement: report platform discovery diagnostics
The command SHALL report incomplete, inaccessible, or ambiguous live-platform information as actionable diagnostics rather than silently omitting or inventing it. Diagnostics SHALL cover at least missing permissions, unsupported or unrecognised resource types, and inaccessible remote or platform state. For Terraform analysis the command SHALL likewise report unresolved inputs and inaccessible remote state. A resource that cannot be read SHALL be recorded as `unresolved` with a diagnostic, never as a guessed value.

#### Scenario: permission denied during discovery
- **WHEN** a read/list lookup is denied because the principal lacks permission
- **THEN** the command reports a diagnostic naming the inaccessible resource and the implicated permission, records the affected items as `unresolved`, and continues discovering what it can access

#### Scenario: unsupported resource type encountered
- **WHEN** discovery encounters a resource type the command does not support
- **THEN** the command reports it as an unsupported type diagnostic rather than omitting it silently

#### Scenario: inaccessible state reported
- **WHEN** Terraform remote state or a platform resource cannot be accessed
- **THEN** the command reports the inaccessibility as a diagnostic and does not fabricate the missing values

### Requirement: preserve BTP read-only safety boundary
The command SHALL operate within a read-only safety boundary for all SAP BTP interactions. It SHALL invoke only documented read/list BTP CLI commands and read/list BTP MCP tools, SHALL use `btp target --global-account <subdomain>` only as the account-selection prelude to such read/list calls, and SHALL NOT invoke, suggest, or approve any command that creates, updates, deletes, assigns, unassigns, enables, disables, logs in, configures, or otherwise mutates BTP state — even when the user requests it.

#### Scenario: mutation requested
- **WHEN** the user asks the command to change, create, delete, or assign a BTP resource during analysis
- **THEN** the command declines, explains the read-only boundary, and does not perform or suggest the mutation

#### Scenario: account targeting prelude
- **WHEN** the command must select a global account before a read/list lookup
- **THEN** it may run `btp target --global-account <subdomain>` solely as the prelude to read/list commands and runs no other state-changing command

