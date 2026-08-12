# Analyse Command Capability

## Purpose

Defines the behaviour of `/btp-iac:analyse`: reading application source code to extract concrete infrastructure signals and enriching `specs/scenario.md` in place.

## Requirements

### Requirement: read application source code
The command SHALL ask the user for the path to their application source code before proceeding, then read the files at that path.

#### Scenario: path provided
- **WHEN** the user supplies a source code path
- **THEN** the command reads the relevant files at that path

### Requirement: extract infrastructure signals
The command SHALL extract concrete infrastructure signals the user cannot be expected to know: service dependencies, memory requirements, role definitions, and build artifact details.

#### Scenario: signals extracted
- **WHEN** source code is read
- **THEN** the command identifies service dependencies, memory requirements, role definitions from the security descriptor, and build artifact details

### Requirement: support MTA, CAP, and CF app formats
The command SHALL support three BTP application formats.

#### Scenario: MTA app
- **WHEN** the source contains `mta.yaml`, `xs-security.json`, or `xs-app.json`
- **THEN** the command extracts signals from those files

#### Scenario: CAP app
- **WHEN** the source contains `package.json`, `schema.cds`, or `xs-security.json`
- **THEN** the command extracts signals from those files

#### Scenario: CF app
- **WHEN** the source contains `manifest.yml`
- **THEN** the command extracts signals from that file

### Requirement: enrich scenario file in place
The command SHALL enrich `specs/scenario.md` in place with findings from the source code analysis.

#### Scenario: scenario file enriched
- **WHEN** the analysis is complete
- **THEN** `specs/scenario.md` is updated with the extracted infrastructure signals
