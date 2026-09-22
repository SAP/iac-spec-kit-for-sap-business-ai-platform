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
Before resolving any service dependencies, the command SHALL determine each service's `consumption_type`, one of `instance`, `subscription`, or `entitlement-only`. The default is: SaaS applications → `subscription`, technical services → `instance`. The command SHALL let the user confirm or override the type for every service so that any service can be marked `entitlement-only`.

#### Scenario: governance pre-states consumption type
- **WHEN** `memory/governance.md` records the consumption type for a service
- **THEN** the command uses that decision without asking the user

#### Scenario: user confirms or overrides type
- **WHEN** a service's consumption type is not fixed by governance
- **THEN** the command asks the user for that service (one question per service), offering the default and the three choices `instance`, `subscription`, `entitlement-only`

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
- **THEN** the command resolves required entitlements, SaaS subscriptions, service instances with configuration, location, and CF space, and inter-service dependencies in order

### Requirement: validate service plans against governance
The command SHALL validate each service instance plan against governance rules if `memory/governance.md` exists.

#### Scenario: permitted plan
- **WHEN** the plan is in `## Service Plans → <tier> → Permitted` and not in `Forbidden`
- **THEN** the command continues

#### Scenario: forbidden plan — hard block
- **WHEN** the plan is not permitted or is explicitly forbidden for the environment tier
- **THEN** the command stops with the plan, tier, permitted plans, and fix instructions
- **UNLESS** `- Override: true` is set, in which case a warning is logged and the command continues

### Requirement: write services file
The command SHALL write `specs/services.md` with the full dependency-ordered service list. Every service entry SHALL record `consumption_type` (`instance`, `subscription`, or `entitlement-only`); every `instance` entry SHALL also record `location` (`btp` or `cf`) and, when `location: cf`, `cf_space`.

#### Scenario: services file written
- **WHEN** the command completes successfully
- **THEN** `specs/services.md` exists, each entry carries `consumption_type`, each `instance` entry carries `location` (and `cf_space` when `cf`), and the file is the direct input to `/btp-iac.tasks`

### Requirement: validate resolved entitlements
Before writing output, the command SHALL validate every resolved entitlement, subscription, service offering, and plan with the shared platform-validation capability scoped to the landscape global account.

#### Scenario: resolved entitlement unavailable
- **WHEN** the scoped platform lookup does not include a resolved entitlement or subscription
- **THEN** the command asks for a valid replacement before writing `specs/services.md`
