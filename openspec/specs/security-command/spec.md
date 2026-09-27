# Security Command Capability

## Purpose

Defines the behaviour of `/sap-iac.security`: defining authentication and authorisation configuration with governance validation, producing `specs/trust.md`.

## Requirements

### Requirement: read inputs
The command SHALL read `specs/scenario.md`, `specs/landscape.md`, and if present `memory/governance.md` before defining any security configuration.

#### Scenario: inputs read
- **WHEN** the command starts
- **THEN** it reads both spec files and optionally governance

### Requirement: define security configuration
The command SHALL confirm for each subaccount whether role collections are required. When role collections are required, the command SHALL optionally ask which individual roles should be assigned to each collection (each role identified by `role_name`, `role_template_name`, and `role_template_app_id`). Role collections with no roles specified are valid. The command SHALL write each role collection to `specs/trust.md` with its name, optional description, and a structured list of individually named roles (which may be empty). The command SHALL NOT write unstructured "role template assignments" to the trust file.

When role collections are created, the command SHALL ask how assignments should be made: by **user** or by **group**. For each assignment the command SHALL collect the full assignment detail (user name or group name, and the role collection name). When a custom IdP is in use, each assignment entry MUST include the derived origin (see `collect custom IdP URL` requirement).

The command SHALL also define for each subaccount: IdP trust configurations and user and group assignments to role collections.

#### Scenario: role collections required, roles specified
- **WHEN** the user confirms role collections are required and provides individual role entries (role_name, role_template_name, role_template_app_id) for a collection
- **THEN** `specs/trust.md` records that collection with a structured roles list containing each named role

#### Scenario: role collections required, no roles specified
- **WHEN** the user confirms role collections are required but does not specify individual roles
- **THEN** `specs/trust.md` records the collection with an empty roles list and no role assignments

#### Scenario: role collections not required
- **WHEN** the user confirms role collections are not required for a subaccount
- **THEN** no role collection entries appear in `specs/trust.md` for that subaccount

#### Scenario: role collection assignment by user, no custom IdP
- **WHEN** role collections are created and no custom IdP is in use
- **THEN** the command asks whether assignments are by user or group and collects the assignment detail (name, role collection) without an origin field

#### Scenario: role collection assignment by user, custom IdP
- **WHEN** role collections are created and a custom IdP is in use
- **THEN** the command asks whether assignments are by user or group and collects the assignment detail (name, origin, role collection), where origin is derived from the IdP URL

#### Scenario: role collection assignment by group
- **WHEN** the user selects group assignment
- **THEN** the command collects group name, role collection, and (when a custom IdP is in use) the derived origin

### Requirement: validate IdP against governance
The command SHALL validate the custom IdP configuration against governance rules if `memory/governance.md` exists.

#### Scenario: custom IdP required and configured
- **WHEN** `Custom IdP: required` and a custom IdP trust is being configured
- **THEN** the command continues

#### Scenario: custom IdP required but missing — hard block
- **WHEN** `Custom IdP: required` and no custom IdP is configured
- **THEN** the command stops with the rule and fix instructions
- **UNLESS** `- Override: true` is set

### Requirement: validate role collections against governance
The command SHALL validate default role collection assignments against governance rules.

#### Scenario: required assignments missing — hard block
- **WHEN** `Default role collections` specifies assignments that are not included
- **THEN** the command stops with the required assignments and fix instructions
- **UNLESS** `- Override: true` is set

### Requirement: write trust file
The command SHALL write `specs/trust.md` with the complete security configuration.

#### Scenario: trust file written
- **WHEN** the command completes successfully
- **THEN** `specs/trust.md` exists and is the direct input to `/sap-iac.tasks`

### Requirement: collect custom IdP URL
When the command detects a custom IdP signal in `specs/scenario.md` or `specs/landscape.md`, it SHALL ask the user for the IdP URL. If no signal is present in those files, the command SHALL also check `memory/governance.md`: if it exists and contains `Custom IdP: required`, that SHALL be treated as a custom IdP signal and the URL question MUST be asked. If neither source provides a signal, the command SHALL proceed without asking. This question MUST NOT be skipped or omitted when a signal is present. From the URL, the command SHALL derive the origin as `<first-subdomain>-platform` (the portion of the hostname before the first `.`, followed by `-platform`). The derived origin and the raw URL SHALL be recorded in `specs/trust.md`.

#### Scenario: custom IdP detected, URL collected
- **WHEN** a custom IdP signal is present in `specs/scenario.md` or `specs/landscape.md`
- **THEN** the command asks for the IdP URL and derives origin as `<first-subdomain>-platform`

#### Scenario: custom IdP required by governance, not in input specs
- **WHEN** no custom IdP signal is present in the input specs AND `memory/governance.md` contains `Custom IdP: required`
- **THEN** the command asks for the IdP URL and derives origin as `<first-subdomain>-platform`

#### Scenario: origin derivation example
- **WHEN** the IdP URL is `ainfvn15r.accounts.ondemand.com`
- **THEN** the derived origin is `ainfvn15r-platform`

#### Scenario: no custom IdP signal
- **WHEN** no custom IdP signal is present in the input specs and governance does not require one
- **THEN** the command does not ask for an IdP URL

### Requirement: collect CF space user assignments
When the command detects that a Cloud Foundry space is to be created (from `specs/landscape.md` or `specs/scenario.md`), it SHALL ask which users should be assigned to which space roles. For each assignment entry the command SHALL collect: `space_name` (matching a CF space defined in `specs/landscape.md`), `username`, and exactly one role from the fixed list `space_auditor`, `space_developer`, `space_manager`, `space_supporter`. The `origin` field is always written to `specs/trust.md`: when a custom IdP is in use, the command SHALL automatically use the derived origin from the `collect custom IdP URL` requirement and MUST NOT ask the user to provide or override it; when no custom IdP is in use, the command SHALL use `sap.ids` as the origin. Each `(space_name, username, role)` tuple is a separate entry in `specs/trust.md` and maps to one `cloudfoundry_space_role` resource. To assign multiple roles to the same user, the user provides multiple entries with the same username. These assignments SHALL be written to `specs/trust.md`.

#### Scenario: CF space in scope, assignments collected — with custom IdP
- **WHEN** a Cloud Foundry space creation is detected and a custom IdP is in use
- **THEN** the command asks for CF space user assignments and collects space_name, username, and one role per entry; origin is set to the derived custom-IdP origin

#### Scenario: CF space in scope, assignments collected — without custom IdP
- **WHEN** a Cloud Foundry space creation is detected and no custom IdP is in use
- **THEN** the command asks for CF space user assignments and collects space_name, username, and one role per entry; origin is set to `sap.ids`

#### Scenario: multiple roles for one user
- **WHEN** a user needs multiple space roles
- **THEN** each role is recorded as a separate entry with the same username, producing one cloudfoundry_space_role resource per entry

#### Scenario: space_name must match landscape
- **WHEN** collecting CF space assignments
- **THEN** the space_name field MUST match a CF space name defined in specs/landscape.md

#### Scenario: no CF space in scope
- **WHEN** no Cloud Foundry space creation is detected
- **THEN** the command does not ask for CF space user assignments

#### Scenario: role values constrained
- **WHEN** collecting CF space roles
- **THEN** only `space_auditor`, `space_developer`, `space_manager`, `space_supporter` are accepted

### Requirement: distinguish explicit trust-configuration origin
For each custom IdP trust configuration, the command SHALL preserve the raw IdP URL and derived origin used by other resources. It SHALL separately capture an optional origin for `btp_subaccount_trust_configuration` and SHALL record `Trust configuration origin explicit: true` with that value. The command SHALL NOT treat the URL-derived origin as an explicit trust-configuration origin. When the user does not explicitly supply a trust-configuration origin, the trust configuration data in `specs/trust.md` SHALL contain the IdP URL but no trust-configuration origin value or explicit-origin marker.

#### Scenario: custom IdP URL without an explicit trust-configuration origin
- **WHEN** a custom IdP URL is collected and the user does not explicitly provide an origin for its trust configuration
- **THEN** `specs/trust.md` retains the IdP URL and the separately derived origin for resources that use it, but records no trust-configuration origin or explicit-origin marker

#### Scenario: explicit trust-configuration origin supplied
- **WHEN** the user explicitly provides an origin for a custom IdP trust configuration
- **THEN** `specs/trust.md` records that value as the trust-configuration origin with `Trust configuration origin explicit: true`

#### Scenario: derived origin remains available to other resources
- **WHEN** a custom IdP URL produces a derived origin and no trust-configuration origin is explicitly supplied
- **THEN** role collection assignments and Cloud Foundry space-role assignments continue to use the derived origin according to their existing requirements
