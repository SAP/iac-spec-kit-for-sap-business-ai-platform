# Analyse Command Capability

## Purpose

Defines the behaviour of `/btp-iac.analyse`: reading application source code or Terraform configuration to extract concrete infrastructure signals as evidenced, confidence-rated facts, and enriching `specs/scenario.md` in place without leaking secrets or overwriting user content.

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
The command SHALL support analysing Terraform configuration. It SHALL inspect all root `*.tf` and `*.tf.json` files, excluding `.terraform`, `.git`, and generated or vendored directories, and SHALL NOT stop after finding initial signals. The command SHALL follow local module `source` references up to a bounded recursion depth and SHALL explicitly report external modules that could not be inspected. It SHALL extract BTP-domain findings including subaccounts, regions, hierarchy, and labels; entitlements, subscriptions, service offerings/plans, and service instances; trust configuration, role collections, assignments, destinations, and connectivity; provider aliases and `required_providers`; and `data`, `locals`, `outputs`, dependency edges, `for_each`, `count`, and `depends_on`. For input variables it SHALL record description, validation, `sensitive`, and whether a value has no default, and SHALL record local module interfaces and external module source/version.

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
The command SHALL enrich `specs/scenario.md` in place with findings from the analysis, writing them only between the paired markers `<!-- btp-iac:analyse:begin -->` and `<!-- btp-iac:analyse:end -->`. On re-runs the command SHALL replace only the content enclosed by those markers and SHALL preserve all content outside them. The command SHALL NOT perform a whole-file rewrite. The generated section SHALL organise findings under "Observed services", "Observed security and connectivity", "Terraform architecture", and "Unresolved mappings and conflicts".

#### Scenario: scenario file enriched between markers
- **WHEN** the analysis is complete
- **THEN** the extracted findings are written between `<!-- btp-iac:analyse:begin -->` and `<!-- btp-iac:analyse:end -->`, organised into the observed-services, observed-security, Terraform-architecture, and unresolved subsections

#### Scenario: re-run preserves surrounding content
- **WHEN** the command runs again on a file that already contains the paired markers
- **THEN** only the content between the markers is replaced and all content outside them is preserved

### Requirement: report completion
The command SHALL produce a completion report identifying the files analysed and skipped, the findings written, the redactions performed, the unresolved mappings, and the recommended next command.

#### Scenario: completion report produced
- **WHEN** the analysis finishes
- **THEN** the command reports analysed and skipped files, findings written, redactions performed, unresolved mappings, and the recommended next command
