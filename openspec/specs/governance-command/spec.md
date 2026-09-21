# Governance Command Capability

## Purpose

Defines the behaviour of the `/btp-iac.govern` agent command: project root detection, existing file handling, category completeness evaluation, and writing the governance file.

## Requirements

### Requirement: locate project root
The command SHALL identify the btp-iac project root by walking up from the current working directory until it finds a directory containing `specs/`, `memory/`, and `terraform/` subdirectories.

#### Scenario: project root found
- **WHEN** a directory in the ancestor chain contains `specs/`, `memory/`, and `terraform/`
- **THEN** the command uses that directory as the project root for all file reads and writes

#### Scenario: project root not found
- **WHEN** no ancestor directory contains the required subdirectories
- **THEN** the command stops with: "Could not locate a btp-iac project root from the current directory. Run this command from within a project created by `btp-iac init`."
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

### Requirement: evaluate prompt completeness
The command SHALL assess whether the user's invocation prompt covers all six governance categories: regions, account setup, naming, service plans, security, and cost controls.

#### Scenario: all categories covered in prompt
- **WHEN** the user's prompt contains rules for all six categories
- **THEN** the command writes governance.md directly without asking follow-up questions

#### Scenario: some categories missing from prompt
- **WHEN** the user's prompt covers only some categories
- **THEN** the command asks one targeted question per missing category and no more
- **AND** does not re-ask about categories already addressed in the prompt

#### Scenario: no governance detail in prompt
- **WHEN** the user invokes the command with no governance detail
- **THEN** the command asks one targeted question for each of the six categories in sequence

### Requirement: write governance file
The command SHALL write the collected rules to `<project-root>/memory/governance.md` in the structured markdown format defined by the governance-format spec.

#### Scenario: successful write
- **WHEN** all required information has been collected
- **THEN** `<project-root>/memory/governance.md` is created or overwritten with the structured rules
- **AND** the command confirms the file was saved and lists the rules that will be enforced

### Requirement: validate explicitly specified platform values
When the input identifies a global-account subdomain and named service offerings, subscriptions, service-plan pairs, or regions, the command SHALL validate them against the platform using the shared platform-validation capability before writing governance. CLI checks SHALL target that global account; region checks SHALL exclude NEO entries.

#### Scenario: no global-account target supplied
- **WHEN** governance input names platform values but does not provide a global-account subdomain
- **THEN** the command retains those values without a targeted live CLI check
