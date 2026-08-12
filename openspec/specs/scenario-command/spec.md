# Scenario Command Capability

## Purpose

Defines the behaviour of `/btp-iac:scenario`: translating a plain-language application description into structured BTP infrastructure requirements.

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
