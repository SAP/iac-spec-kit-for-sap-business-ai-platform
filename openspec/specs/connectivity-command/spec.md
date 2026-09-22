# Connectivity Command Capability

## Purpose

Defines the behaviour of `/btp-iac.connectivity`: eliciting, structuring, and writing BTP destination and destination-certificate requirements to `specs/connectivity.md`.

## Requirements

### Requirement: read existing scenario destinations
When `specs/scenario.md` exists and contains destination information, the command SHALL reuse those destinations as a starting point and ask only for information that is missing or underspecified rather than re-collecting everything from scratch.

#### Scenario: scenario contains destinations
- **WHEN** `specs/scenario.md` exists and lists one or more destinations
- **THEN** the command presents those destinations and asks the user to confirm, extend, or modify them before collecting further detail

#### Scenario: scenario has no destinations
- **WHEN** `specs/scenario.md` exists but contains no destination information
- **THEN** the command collects destination and certificate requirements from scratch as if no scenario existed

#### Scenario: no scenario file
- **WHEN** `specs/scenario.md` does not exist
- **THEN** the command collects destination and certificate requirements from scratch

### Requirement: elicit destination configuration
The command SHALL guide the user through specifying one or more `btp_subaccount_destination_generic` resources by presenting example patterns and asking targeted questions.

The command SHALL present the following canonical example patterns as hints to help users identify their use case:

- **HTTP / NoAuthentication** — internet-accessible service, no credentials (proxy type `Internet`, auth `NoAuthentication`)
- **HTTP / BasicAuthentication via Cloud Connector** — on-premise system reached through SAP Cloud Connector (proxy type `OnPremise`, auth `BasicAuthentication`, `CloudConnectorLocationId` required)
- **HTTP / OAuth2ClientCredentials** — token-based machine-to-machine call (`clientId`, `tokenServiceURL`, auth `OAuth2ClientCredentials`)
- **RFC** — SAP ABAP system via JCo (`jco.client.*` and `jco.destination.*` properties)
- **LDAP** — directory service (`ldap.url`, `ldap.authentication`, `ldap.user`, `ldap.password`)
- **MAIL** — SMTP relay (`mail.user`, `mail.password`, auth `BasicAuthentication`, proxy `OnPremise`)
- **TCP** — raw TCP connection (`Address` as `host:port`, proxy `OnPremise`)

For each destination the command SHALL collect at minimum: name, destination type (`HTTP`, `RFC`, `LDAP`, `MAIL`, `TCP`), URL or address, authentication method, and whether it is scoped to the subaccount or a specific service instance.

#### Scenario: user selects HTTP pattern
- **WHEN** the user identifies an HTTP destination
- **THEN** the command asks for URL, proxy type, and authentication method, and for `OnPremise` proxy type additionally asks for `CloudConnectorLocationId`

#### Scenario: user selects RFC pattern
- **WHEN** the user identifies an RFC destination
- **THEN** the command asks for the ABAP host, system number, client, user, password, and proxy type

#### Scenario: destination has credentials
- **WHEN** the collected destination includes a password, client secret, or other credential
- **THEN** the command records that a credential is required but does NOT write the credential value to `specs/connectivity.md`; it instead records a placeholder and adds a warning that the value must be supplied at apply time

### Requirement: elicit certificate configuration
The command SHALL optionally guide the user through specifying one or more `btp_subaccount_destination_certificate` resources.

The command SHALL present the following canonical patterns:

- **Subaccount-level certificate** — a PEM, P12, JKS, or PFX file uploaded to the subaccount (`subaccount_id`, `certificate_name` with valid extension, `certificate_content` from file)
- **Service-instance-level certificate** — the same, additionally scoped to a specific service instance (`service_instance_id`)

For each certificate the command SHALL collect: certificate name (including extension), file path or indication that the content will be supplied at apply time, subaccount scope, and optionally a service instance ID.

#### Scenario: user adds a certificate
- **WHEN** the user specifies a certificate
- **THEN** the command records name, scope, and a placeholder for content; it does NOT write the raw certificate bytes to `specs/connectivity.md`

#### Scenario: user skips certificates
- **WHEN** the user confirms no certificates are needed
- **THEN** the command omits the certificates section from `specs/connectivity.md`

### Requirement: write connectivity file
The command SHALL write the collected requirements to `specs/connectivity.md`.

#### Scenario: connectivity file written
- **WHEN** the command completes successfully
- **THEN** `specs/connectivity.md` exists and contains a structured list of destinations and, if specified, certificates

#### Scenario: connectivity file is input to tasks
- **WHEN** `specs/connectivity.md` exists
- **THEN** `/btp-iac.tasks` reads it and includes destination and certificate tasks in `specs/tasks.md`

### Requirement: apply BTP operation safety boundary
The command SHALL apply the shared BTP platform-validation and read-only boundary: it SHALL only invoke BTP read/list operations; it SHALL never mutate BTP state. `btp target --global-account <subdomain>` is permitted only as an account-selection prelude to BTP CLI read/list commands.

#### Scenario: command does not mutate BTP state
- **WHEN** the command runs
- **THEN** it does not create, update, or delete any BTP resource, and does not invoke any BTP CLI command other than read/list commands and the permitted target prelude
