# Scenario Command Capability

## Purpose

Defines the behaviour of `/btp-iac.scenario`: translating a plain-language application description into structured BTP infrastructure requirements.

## Requirements

### Requirement: translate description into structured requirements
The command SHALL accept a plain-language description of the app or service to be deployed and produce a structured set of BTP infrastructure requirements.

#### Scenario: description provided
- **WHEN** the user describes their application in natural language
- **THEN** the command translates it into structured infrastructure requirements

### Requirement: ask targeted follow-up questions
The command SHALL ask at most three targeted follow-up questions covering runtime, integration points, and environment isolation before writing output.

#### Scenario: follow-ups asked
- **WHEN** the description is ambiguous or incomplete
- **THEN** the command asks up to three questions and no more

### Requirement: write scenario file
The command SHALL write results to `specs/scenario.md`.

#### Scenario: output written
- **WHEN** the command completes
- **THEN** `specs/scenario.md` exists and contains the structured requirements

### Requirement: validate supplied platform values
The command SHALL apply the shared platform-validation capability to supplied regions, named service offerings, subscriptions, and plan pairs. When governance specifies a preferred infrastructure provider other than `none`, it SHALL compare each supplied region's unambiguous returned provider metadata to that preference and warn, without blocking, on a mismatch. It SHALL not infer a provider from a region code. When a live check needs a global-account target that is absent from governance and input, it SHALL collect the subdomain within its targeted follow-up limit.

#### Scenario: target needed for supplied region
- **WHEN** scenario input supplies a region that needs live validation but no global-account subdomain
- **THEN** the command obtains the subdomain in a targeted follow-up before checking the region

#### Scenario: preferred provider mismatch
- **WHEN** governance selects a preferred infrastructure provider and a supplied region returns different unambiguous provider metadata
- **THEN** the command warns with the region and both providers
- **AND** retains the region for further scenario processing
