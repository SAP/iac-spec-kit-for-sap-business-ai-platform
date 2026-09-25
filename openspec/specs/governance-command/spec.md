# Governance Command Capability

## Purpose

Defines the behaviour of the `/sap-iac.govern` agent command: project root detection, existing file handling, category completeness evaluation, and writing the governance file.

## Requirements

### Requirement: locate project root
The command SHALL identify the sap-iac project root by walking up from the current working directory until it finds a directory containing `specs/`, `memory/`, and `terraform/` subdirectories.

#### Scenario: project root found
- **WHEN** a directory in the ancestor chain contains `specs/`, `memory/`, and `terraform/`
- **THEN** the command uses that directory as the project root for all file reads and writes

#### Scenario: project root not found
- **WHEN** no ancestor directory contains the required subdirectories
- **THEN** the command stops with: "Could not locate a sap-iac project root from the current directory. Run this command from within a project created by `sap-iac init`."
- **THEN** no files are created or modified

### Requirement: detect existing governance file
The command SHALL check whether `<project-root>/memory/governance.md` exists before collecting any rules.

#### Scenario: file exists
- **WHEN** `<project-root>/memory/governance.md` is present
- **THEN** the command reads and summarises the existing rules
- **AND** asks the user what they want to change rather than starting from scratch

#### Scenario: file does not exist
- **WHEN** `<project-root>/memory/governance.md` is absent
- **THEN** the command proceeds to collect governance rules from scratch

### Requirement: collect preferred infrastructure provider
Before collecting any other missing governance details, the command SHALL determine the preferred infrastructure provider. Valid values are `AWS`, `Microsoft Azure`, `Google Cloud`, `SAP Cloud Infrastructure`, `Alibaba Cloud`, and `none`.

#### Scenario: provider preference absent
- **WHEN** a new governance invocation does not state a provider preference
- **THEN** the command asks for it as its first follow-up question, before any category question

### Requirement: fixed governance-question order
The command SHALL assess whether the user's invocation prompt covers all required governance details. It SHALL ask one question at a time and, after skipping details already supplied, use this fixed order: Preferred Infrastructure, Region, Naming, Environments, Service Plans, Security, Cost Control. Environments covers permitted runtime environments and environment-tier names. The command SHALL read the global-account subdomain only from `memory/global-account.md` and SHALL not request it.

#### Scenario: all required details covered in prompt
- **WHEN** the user's prompt contains every required governance detail and states a valid preferred infrastructure provider
- **THEN** the command writes governance.md directly without asking follow-up questions

#### Scenario: some required details missing from prompt
- **WHEN** the user's prompt covers only some required details
- **THEN** the command asks only missing details in this order: Preferred Infrastructure, Region, Naming, Environments, Service Plans, Security, Cost Control
- **AND** does not re-ask about details already addressed in the prompt

#### Scenario: no governance detail in prompt
- **WHEN** the user invokes the command with no governance detail
- **THEN** the command asks the seven questions in this sequence: Preferred Infrastructure, Region, Naming, Environments, Service Plans, Security, Cost Control

### Requirement: write governance file
The command SHALL write the collected rules to `<project-root>/memory/governance.md` in the structured markdown format defined by the governance-format spec.

#### Scenario: successful write
- **WHEN** all required information has been collected
- **THEN** `<project-root>/memory/governance.md` is created or overwritten with the structured rules
- **AND** the command confirms the file was saved and lists the rules that will be enforced

### Requirement: validate explicitly specified platform values
When the input identifies named service offerings, subscriptions, service-plan pairs, or regions, the command SHALL validate them against the platform using the shared platform-validation capability before writing governance. CLI checks SHALL target the global account configured in `memory/global-account.md` when one is present; region checks SHALL exclude NEO entries.

#### Scenario: no global-account target supplied
- **WHEN** `memory/global-account.md` has no configured subdomain
- **THEN** the command retains those values without a targeted live CLI check and does not request a subdomain

#### Scenario: provider mismatch
- **WHEN** a requested region has unambiguous provider metadata that differs from the preferred infrastructure provider
- **THEN** the command warns with the region and both providers
- **AND** continues writing governance
