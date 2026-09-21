# Scenario Command Capability

## Purpose

Defines the behaviour of `/btp-iac.scenario`: translating a plain-language application description into structured BTP infrastructure requirements.

## Requirements

### Requirement: translate description into structured requirements
The command SHALL accept a plain-language description of the app or service to be deployed and produce a structured set of BTP infrastructure requirements.

#### Scenario: description provided
- **WHEN** the user describes their application in natural language
- **THEN** the command translates it into structured infrastructure requirements

### Requirement: governance-aware targeted follow-up questions
The command SHALL read `memory/governance.md` when it exists before asking follow-up questions. It SHALL ask at most three targeted follow-up questions covering unresolved runtime and Cloud Foundry sizing, destinations, and setup structure before writing output.

#### Scenario: follow-ups asked
- **WHEN** the description is ambiguous or incomplete
- **THEN** the command asks up to three questions and no more

#### Scenario: environment already governed
- **WHEN** scenario input and governance together sufficiently determine the required runtime and setup structure
- **THEN** the command does not ask an environment follow-up question

#### Scenario: environment decision remains unresolved
- **WHEN** governance permits multiple runtimes and neither governance nor scenario input selects the runtime required by the scenario
- **THEN** the command asks one narrowly scoped environment follow-up question

#### Scenario: Cloud Foundry sizing required
- **WHEN** Cloud Foundry is selected or permitted and scenario input does not specify its application memory allocation or sizing
- **THEN** the runtime follow-up asks for Cloud Foundry memory allocation or sizing

#### Scenario: destinations unresolved
- **WHEN** scenario input does not establish whether integrations require destinations
- **THEN** the integration follow-up asks whether destinations are required and, if so, which destinations and purposes are needed

#### Scenario: setup structure unresolved
- **WHEN** scenario input does not establish how stages should be structured
- **THEN** the structure follow-up asks, for example, whether each stage needs its own subaccount

### Requirement: write scenario file
The command SHALL write results to `specs/scenario.md`.

#### Scenario: output written
- **WHEN** the command completes
- **THEN** `specs/scenario.md` exists and contains the structured requirements

### Requirement: validate supplied platform values
The command SHALL apply the shared platform-validation capability to supplied regions, named service offerings, subscriptions, and plan pairs. When governance specifies a preferred infrastructure provider other than `none`, it SHALL compare each supplied region's unambiguous returned provider metadata to that preference and warn, without blocking, on a mismatch. It SHALL not infer a provider from a region code. It SHALL use the optional target configured by initialization in `memory/global-account.md`; when that target is absent, it SHALL retain user input without a targeted live check and SHALL not request a subdomain.

#### Scenario: target needed for supplied region
- **WHEN** scenario input supplies a region that needs live validation but `memory/global-account.md` has no configured subdomain
- **THEN** the command retains the input without a targeted live check and does not request a subdomain

#### Scenario: preferred provider mismatch
- **WHEN** governance selects a preferred infrastructure provider and a supplied region returns different unambiguous provider metadata
- **THEN** the command warns with the region and both providers
- **AND** retains the region for further scenario processing
