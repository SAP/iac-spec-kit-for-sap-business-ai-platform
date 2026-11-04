# Governance Format Capability

## Purpose

Defines the structured markdown format of `memory/governance.md` and all required sections.

## Requirements

### Requirement: regions section
The governance file SHALL contain a Regions section specifying the preferred infrastructure provider, allowed and forbidden BTP regions, the data residency classification, the compliance reason behind the restriction, and the preferred default region.

#### Scenario: regions section present
- **WHEN** `memory/governance.md` is read by a downstream command
- **THEN** it contains a `## Regions` section with:
  - `- Preferred infrastructure provider:` — `AWS`, `Microsoft Azure`, `Google Cloud`, `SAP Cloud Infrastructure`, `Alibaba Cloud`, or `none`
  - `- Allowed:` — comma-separated list of approved BTP region codes (e.g. `eu10, eu20`)
  - `- Forbidden:` — comma-separated list of explicitly disallowed region codes, or `none`
  - `- Data residency:` — classification such as `EU only`, `US only`, or `unrestricted`
  - `- Compliance reason:` — the regulatory or policy driver (e.g. `GDPR — customer data must not leave the EU`)
  - `- Preferred region:` — the default region to use when none is explicitly specified (e.g. `eu10`)

#### Scenario: legacy regions section
- **WHEN** an existing governance file lacks `Preferred infrastructure provider`
- **THEN** downstream commands treat it as `none`

### Requirement: account setup section
The governance file SHALL contain an Account Setup section specifying the runtime environment types permitted in the project. The global account subdomain is stored separately in `memory/global-account.md`, which initialization creates. The `Environments` question group writes its runtime answer here and its environment-tier answer to `## Naming → Environments`; it does not rename either persisted section.

The section MAY additionally contain a `Tier classifications` mapping to override the default keyword-based inference of `usage` and `beta_enabled` in the accounts command. The mapping is optional; when absent, inference applies. Each entry explicitly records both independent attributes for one tier.

The section MAY additionally contain an optional directory policy that pre-defines how directories are used in the global account. Each directory-policy field is independent and optional; when a field is absent, `/sap-iac.accounts` resolves that aspect through an interactive discussion instead.

#### Scenario: account setup section present
- **WHEN** `memory/governance.md` is read by a downstream command
- **THEN** it contains a `## Account Setup` section with:
  - `- Allowed environments:` — `Cloud Foundry`, `Kyma`, or both
  - `- Tier classifications:` _(optional)_ — exact, case-insensitive tier-name mapping. Each nested entry specifies both `usage` (`USED_FOR_PRODUCTION` or `NOT_USED_FOR_PRODUCTION`) and `beta_enabled` (`true` or `false`), using this nested form:

    ```markdown
    - Tier classifications:
      - integration: usage = NOT_USED_FOR_PRODUCTION, beta_enabled = false
    ```
  - `- Directory management:` _(optional)_ — `managed` or `unmanaged`; the default feature posture for directories. `unmanaged` means plain organisational folders (`btp_directory` DEFAULT feature only); `managed` means directories carry the features named in `Directory features`.
  - `- Directory features:` _(optional)_ — comma-separated subset of `ENTITLEMENTS`, `AUTHORIZATIONS` applied to managed directories, or `N/A` when unmanaged. `AUTHORIZATIONS` implies `ENTITLEMENTS` and requires a directory subdomain.
  - `- Subaccount directory assignment:` _(optional)_ — free-text policy for assigning subaccounts to directories (e.g. `one directory per environment tier`).
  - `- Shared subaccounts:` _(optional)_ — subaccounts shared across teams, stages, or applications with their owners (e.g. `shared-connectivity → platform team`), or `none`.

### Requirement: optional directory policy
The `## Account Setup` section MAY contain the optional directory-policy fields `Directory management`, `Directory features`, `Subaccount directory assignment`, and `Shared subaccounts`, each independently governing one aspect of directory setup. When a directory-policy field is present, `/sap-iac.accounts` SHALL apply it without discussion; when a field is absent, `/sap-iac.accounts` SHALL resolve that aspect interactively.

#### Scenario: directory policy field present
- **WHEN** `memory/governance.md` contains a `Directory management`, `Directory features`, `Subaccount directory assignment`, or `Shared subaccounts` field under `## Account Setup`
- **THEN** `/sap-iac.accounts` applies that field directly and does not open an interactive discussion for that aspect

#### Scenario: directory policy field absent
- **WHEN** `memory/governance.md` omits a directory-policy field (including a legacy file with no directory policy at all)
- **THEN** `/sap-iac.accounts` resolves that aspect through an interactive discussion before writing `specs/landscape.md`

#### Scenario: optional tier override fields absent
- **WHEN** `memory/governance.md` does not contain `- Tier classifications:`
- **THEN** the accounts command applies keyword-based inference for `usage` and `beta_enabled`

#### Scenario: tier override fields present
- **WHEN** `memory/governance.md` contains a matching `Tier classifications` entry
- **THEN** the accounts command uses its explicit values instead of keyword inference

#### Scenario: duplicate tier classification
- **WHEN** `Tier classifications` contains the same normalized tier name more than once
- **THEN** the accounts command stops and asks the user to correct the conflicting governance entries before classifying accounts
- **AND** `- Override: true` does not bypass this governance-data error

### Requirement: naming section
The governance file SHALL contain a Naming section specifying the subaccount, Cloud Foundry organization, Kyma environment, and Cloud Foundry space naming patterns, token definitions, environment tiers, directory structure preference, and maximum subaccount name length.

#### Scenario: naming section present
- **WHEN** `memory/governance.md` is read by a downstream command
- **THEN** it contains a `## Naming` section with:
  - `- Subaccount pattern:` — the naming template (e.g. `{org}-{env}-{app}`)
  - `- Pattern tokens:` — definition of each token in the pattern (e.g. `{org} = organisation short code, {env} = environment tier, {app} = application identifier`)
  - `- Cloud Foundry org pattern:` — naming template for Cloud Foundry organizations
  - `- Kyma environment pattern:` — naming template for Kyma environments
  - `- Cloud Foundry space pattern:` — naming template for Cloud Foundry spaces
  - `- Environments:` — comma-separated ordered list of environment tiers (e.g. `dev, test, prod`)
  - `- Directory structure:` — `flat` or `hierarchical`
  - `- Max name length:` — maximum character length for subaccount names (BTP platform limit is 255)

### Requirement: service plans section
The governance file SHALL contain a Service Plans section with permitted and forbidden plans listed per environment tier, including the reason for any forbidden plan and the default plan for each tier.

#### Scenario: service plans section present
- **WHEN** `memory/governance.md` is read by a downstream command
- **THEN** it contains a `## Service Plans` section with a subsection per environment tier, each containing:
  - `- Permitted:` — comma-separated list of allowed plan names for that tier
  - `- Forbidden:` — comma-separated list of disallowed plan names, or `none`
  - `- Forbidden reason:` — explanation for each forbidden plan (e.g. `free plan has no SLA — not permitted in prod`)
  - `- Default plan:` — the plan to use when multiple are permitted and none is specified

### Requirement: security section
The governance file SHALL contain a Security section specifying the custom IdP requirement and type, role collection scope and default assignments.

#### Scenario: security section present
- **WHEN** `memory/governance.md` is read by a downstream command
- **THEN** it contains a `## Security` section with:
  - `- Custom IdP:` — `required`, `optional`, or `not required`
  - `- IdP type:` — `SAML 2.0`, `OIDC`, or `N/A` if no custom IdP
  - `- Default role collections:` — named role collections and which user groups receive them (e.g. `BTP_OPERATOR → platform admins, BTP_VIEWER → developers`)
  - `- Role collection scope:` — `platform`, `application`, or `both`

### Requirement: cost controls section
The governance file SHALL contain a Cost Controls section specifying which metered services trigger warnings, the cost centre tag key name, and whether the tag is mandatory.

#### Scenario: cost controls section present
- **WHEN** `memory/governance.md` is read by a downstream command
- **THEN** it contains a `## Cost Controls` section with:
  - `- Metered service warning:` — `enabled` or `disabled`
  - `- Metered services:` — comma-separated list of specific BTP service names that trigger the warning (e.g. `hana-cloud, ai-core, build-workzone`), or `all` to warn on any metered service
  - `- Cost centre tag:` — `required`, `optional`, or `not required`
  - `- Tag key name:` — the Terraform resource attribute name for the cost centre tag (e.g. `cost_center`, `labels`)

### Requirement: service restrictions section
The governance file MAY contain an optional `## Service Restrictions` section. Its presence indicates org-wide service and service-plan restrictions were declared during `govern`. Its absence indicates no org-wide restrictions were specified.

#### Scenario: service restrictions section absent
- **WHEN** `memory/governance.md` does not contain `## Service Restrictions`
- **THEN** downstream commands apply no org-wide service or plan filtering beyond the per-tier `## Service Plans` rules

#### Scenario: service restrictions section present with link
- **WHEN** `memory/governance.md` contains `## Service Restrictions` with `Policy source: link`
- **THEN** it contains:
  - `- Policy source: link`
  - `- Policy document:` — a URL to the external Markdown restrictions document
  - `- Allowed services:` — comma-separated BTP service names, or `all`
  - `- Allowed service-plan pairs:` — service/plan pairs (e.g. `hana-cloud/hana`), or `all`
  - `- Denied services:` — comma-separated BTP service names, or `none`
  - `- Denied service-plan pairs:` — service/plan pairs, or `none`
- **AND** downstream commands fetch the linked document when they need the full policy content

#### Scenario: service restrictions section present with inline policy
- **WHEN** `memory/governance.md` contains `## Service Restrictions` with `Policy source: inline`
- **THEN** it contains:
  - `- Policy source: inline`
  - `- Policy document: N/A`
  - `- Allowed services:` — comma-separated BTP service names, or `all`
  - `- Allowed service-plan pairs:` — service/plan pairs, or `all`
  - `- Denied services:` — comma-separated BTP service names, or `none`
  - `- Denied service-plan pairs:` — service/plan pairs, or `none`
  - `- Inline policy:` — block scalar containing the free-text policy content

#### Scenario: service restrictions section present with both link and inline
- **WHEN** `memory/governance.md` contains `## Service Restrictions` with `Policy source: both`
- **THEN** it contains all fields from both the link and inline scenarios
- **AND** `Policy document` holds the URL and `Inline policy` holds the free-text content

### Requirement: override flag
The governance file MAY contain a top-level `- Override: true` entry that downstream commands treat as explicit user permission to bypass hard blocks.

#### Scenario: override present
- **WHEN** `memory/governance.md` contains `- Override: true`
- **THEN** downstream commands warn about violations but do not block

### Requirement: deprecated services section
The governance file SHALL contain a Deprecated Services section listing the service/plan combinations explicitly approved for use despite being deprecated. Each approved combination suppresses the deprecation warning for that exact combination across the whole workflow. The global `- Override: true` flag does not cover deprecation; only this list does.

#### Scenario: deprecated services section present
- **WHEN** `memory/governance.md` is read by a downstream command
- **THEN** it contains a `## Deprecated Services` section with:
  - `- Approved:` — comma-separated list of `<service> / <plan>` combinations approved despite deprecation, or `none`

#### Scenario: approved combination suppresses warning
- **WHEN** a `<service> / <plan>` combination appears in `## Deprecated Services → Approved`
- **THEN** downstream commands use that combination without any deprecation warning and without recording a `deprecated:` marker

#### Scenario: override flag does not cover deprecation
- **WHEN** `memory/governance.md` contains `- Override: true` but the deprecated combination is not in `## Deprecated Services → Approved`
- **THEN** the services command still warns and requires explicit user confirmation for that combination
