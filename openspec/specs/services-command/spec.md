# Services Command Capability

## Purpose

Defines the behaviour of `/btp-iac.services`: classifying each service by consumption type, resolving BTP service dependencies with governance validation, and producing `specs/services.md`.

## Requirements

### Requirement: read inputs
The command SHALL read `specs/scenario.md`, `specs/landscape.md`, and if present `memory/governance.md` before resolving any services.

#### Scenario: inputs read
- **WHEN** the command starts
- **THEN** it reads both spec files and optionally governance

### Requirement: classify service consumption type
Before resolving any service dependencies, the command SHALL determine each service's `consumption_type`, one of `instance`, `subscription`, or `entitlement-only`. The instance-vs-subscription distinction SHALL be **derived from the entitlement plan `category`** of the matched service/plan, not guessed from whether the service is SaaS or technical:

- `SERVICE`, `ELASTIC_SERVICE`, `ELASTIC_LIMITED` → service instance (default `consumption_type: instance`)
- `APPLICATION`, `QUOTA_BASED_APPLICATION` → subscription (default `consumption_type: subscription`)

The command SHALL obtain the `category` from the global account's entitlement data. When the BTP CLI is available it SHALL parse the JSON from `btp --format json list accounts/entitlement` (preceded by the permitted `btp target --global-account <subdomain>` prelude when a subdomain is recorded), matching the service by its `name` or `displayName` and the plan by its `name` or `displayName` under `servicePlans`, with no dependency on external tooling such as `jq`. When the CLI is unavailable and the agent is recorded as having BTP MCP support, the command SHALL ask the BTP MCP tool for the service category of the service-name/plan-name combination and map it the same way. When neither the CLI nor MCP is available, or when the required service/plan is not found in the entitlement data, the command SHALL fall back to asking the user.

For every service whose type is not fixed by governance, the command SHALL let the user confirm or override the type so that any service can be marked `entitlement-only`, tailoring the question to the derived type.

#### Scenario: governance pre-states consumption type
- **WHEN** `memory/governance.md` records the consumption type for a service
- **THEN** the command uses that decision without asking the user or deriving from category

#### Scenario: category derives service instance
- **WHEN** the matched plan's `category` is `SERVICE`, `ELASTIC_SERVICE`, or `ELASTIC_LIMITED`
- **THEN** the command derives the default `consumption_type: instance` and records the summary type `service instance`

#### Scenario: category derives subscription
- **WHEN** the matched plan's `category` is `APPLICATION` or `QUOTA_BASED_APPLICATION`
- **THEN** the command derives the default `consumption_type: subscription` and records the summary type `subscription`

#### Scenario: CLI entitlement lookup and matching
- **WHEN** the BTP CLI is available
- **THEN** the command parses `btp --format json list accounts/entitlement`, searches `entitledServices` for a service matching by `name` or `displayName` (case-insensitive, first match wins), then matches the plan by `name` or `displayName` under `servicePlans`, and reads that plan's `category` without relying on external tooling

#### Scenario: MCP category lookup
- **WHEN** the BTP CLI is unavailable and the agent is recorded as having BTP MCP support
- **THEN** the command asks the BTP MCP tool for the service category of the service-name/plan-name combination and maps the returned category to instance or subscription

#### Scenario: user confirms or overrides type
- **WHEN** a service's consumption type is not fixed by governance
- **THEN** the command asks the user for that service (one question per service), offering the default and always allowing `entitlement-only`

#### Scenario: user confirms instance-derived type
- **WHEN** a service's derived type is `instance` and it is not fixed by governance
- **THEN** the command asks that service (one question per service): "For `<service-name>`: (1) service instance or (2) entitlement only? Default is (`<default>`)."

#### Scenario: user confirms subscription-derived type
- **WHEN** a service's derived type is `subscription` and it is not fixed by governance
- **THEN** the command asks that service (one question per service): "For `<service-name>`: (1) app subscription or (2) entitlement only? Default is (`<default>`)."

#### Scenario: no tooling or no match — three-option fallback
- **WHEN** neither the BTP CLI nor BTP MCP is available, or the required service/plan is not found in the entitlement data, or the plan's `category` is not one of the five mapped values
- **THEN** the command informs the user that the type cannot be derived automatically and asks, per service/plan, to choose among three options: `service instance`, `subscription`, or `entitlement-only`

#### Scenario: entitlement-only service
- **WHEN** a service is classified as `entitlement-only`
- **THEN** the command records an entitlement assignment for that service and creates no instance or subscription

### Requirement: determine service instance location
For each service with `consumption_type: instance`, the command SHALL determine whether the instance is created on BTP (using the BTP Terraform provider) or inside a specific Cloud Foundry space (using the CF Terraform provider), recording `location` and, when CF, the target `cf_space`.

#### Scenario: no Cloud Foundry environment in subaccount
- **WHEN** `specs/landscape.md` defines no Cloud Foundry environment for the relevant subaccount
- **THEN** the command records `location: btp` without asking

#### Scenario: Cloud Foundry environment present — ask user
- **WHEN** `specs/landscape.md` defines a Cloud Foundry environment for the relevant subaccount
- **THEN** the command asks whether the instance should be created on BTP or inside a CF space; default is `btp`

#### Scenario: cf location with a single space
- **WHEN** the user chooses CF and the subaccount has exactly one Cloud Foundry space
- **THEN** the command records `location: cf` and `cf_space` set to that space without asking

#### Scenario: cf location with multiple spaces
- **WHEN** the user chooses CF and the subaccount has more than one Cloud Foundry space
- **THEN** the command asks which space and records `location: cf` with the chosen `cf_space`

#### Scenario: cf location with no space defined
- **WHEN** the user chooses CF but the subaccount's Cloud Foundry environment has no space defined in `specs/landscape.md`
- **THEN** the command stops and instructs the user to add a Cloud Foundry space via `/btp-iac.accounts`, without inventing a space name

#### Scenario: location recorded in services file
- **WHEN** a service instance is written to `specs/services.md`
- **THEN** its entry includes `location` (`btp` or `cf`) and, when `location: cf`, `cf_space`

### Requirement: resolve service dependencies
The command SHALL produce a dependency-ordered list of BTP entitlements, subscriptions, and service instances per subaccount.

#### Scenario: services resolved
- **WHEN** inputs are read and services are classified
- **THEN** the command resolves required entitlements, subscriptions, service instances with configuration, location, and CF space, and inter-service dependencies in order

### Requirement: validate service plans against governance
The command SHALL validate each service instance plan against governance rules if `memory/governance.md` exists.

#### Scenario: permitted plan
- **WHEN** the plan is in `## Service Plans → <tier> → Permitted` and not in `Forbidden`
- **THEN** the command continues

#### Scenario: forbidden plan — hard block
- **WHEN** the plan is not permitted or is explicitly forbidden for the environment tier
- **THEN** the command stops with the plan, tier, permitted plans, and fix instructions
- **UNLESS** `- Override: true` is set, in which case a warning is logged and the command continues

### Requirement: Write specs/services.md with full service list
The command SHALL write `specs/services.md` with the full dependency-ordered service list. Every service entry SHALL record `consumption_type` (`instance`, `subscription`, or `entitlement-only`); every `instance` entry SHALL also record `location` (`btp` or `cf`) and, when `location: cf`, `cf_space`. For each service/plan combination, the summary SHALL record the derived type as `service instance` (from `SERVICE`, `ELASTIC_SERVICE`, or `ELASTIC_LIMITED`) or `subscription` (from `APPLICATION` or `QUOTA_BASED_APPLICATION`). When a `parameters:` block has been collected from the catalogue for a service instance task, the entry SHALL include that block with the user-supplied values. When no parameters were collected (either no catalogue entry matched or the user supplied none), the entry SHALL omit the `parameters:` block entirely.

#### Scenario: services file written
- **WHEN** the command completes successfully
- **THEN** `specs/services.md` exists, each entry carries `consumption_type`, each `instance` entry carries `location` (and `cf_space` when `cf`), and the file is the direct input to `/btp-iac.tasks`

#### Scenario: derived type recorded per service/plan
- **WHEN** a service/plan combination is written to `specs/services.md`
- **THEN** the summary records its derived type as `service instance` or `subscription`

#### Scenario: Service instance with collected parameters written to services.md
- **WHEN** a service instance task matched a catalogue entry and the user supplied parameter values
- **THEN** `specs/services.md` contains a `parameters:` block on that entry with the collected values

#### Scenario: Service instance without catalogue match written without parameters block
- **WHEN** a service instance task did not match any catalogue entry
- **THEN** `specs/services.md` contains no `parameters:` block on that entry

### Requirement: record quota_required flag in services output
When the resolved entitlement plan `category` for a service is `SERVICE` or `QUOTA_BASED_APPLICATION`, the command SHALL record `quota_required: true` on that service's entry in `specs/services.md`. For all other resolved categories, and when the category cannot be resolved at all (fallback path or plan not found), the flag SHALL be omitted. The flag is derived from the same category lookup used to determine `consumption_type` and requires no additional lookup.

#### Scenario: SERVICE category sets flag
- **WHEN** the matched plan's `category` is `SERVICE`
- **THEN** the service entry in `specs/services.md` contains `quota_required: true`

#### Scenario: QUOTA_BASED_APPLICATION category sets flag
- **WHEN** the matched plan's `category` is `QUOTA_BASED_APPLICATION`
- **THEN** the service entry in `specs/services.md` contains `quota_required: true`

#### Scenario: other categories omit flag
- **WHEN** the matched plan's `category` is any value other than `SERVICE` or `QUOTA_BASED_APPLICATION`
- **THEN** the service entry in `specs/services.md` does not contain a `quota_required` field

#### Scenario: flag absent when category fallback used
- **WHEN** neither the BTP CLI nor BTP MCP is available and the user manually selects a consumption type
- **THEN** no `quota_required` flag is recorded (the category is unknown and the flag cannot be derived)

#### Scenario: flag absent when plan not found in entitlement data
- **WHEN** the plan cannot be matched in the entitlement data and no category is resolved
- **THEN** no `quota_required` flag is recorded (an unresolved category is not an "other" category and SHALL NOT produce the flag)

### Requirement: validate resolved entitlements
Before writing output, the command SHALL validate every resolved entitlement, subscription, service offering, and plan with the shared platform-validation capability scoped to the landscape global account.

#### Scenario: resolved entitlement unavailable
- **WHEN** the scoped platform lookup does not include a resolved entitlement or subscription
- **THEN** the command asks for a valid replacement before writing `specs/services.md`
