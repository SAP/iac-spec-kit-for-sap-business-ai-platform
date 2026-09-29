# CLI Init Capability

## Purpose

Defines the behavior of `sap-iac init`: argument handling, preflight checks, and fresh, adopt, and agent-only initialization lifecycles.

## Requirements

### Requirement: init command accepts a project name
The CLI SHALL accept at most one optional positional project name and an optional `--agent` flag. When no name is supplied in a TTY, it SHALL offer fresh-project, adopt-current-directory, and agent-only-update choices. Non-interactive execution SHALL require a directory name and `--agent`; fresh and adopt initialization still require interactive input for their remaining prompts.

#### Scenario: name provided with agent flag
- **WHEN** the user runs `sap-iac init my-project --agent claude` and `my-project` does not exist
- **THEN** the CLI selects fresh mode and installs Claude Code files after collecting the initialization input

#### Scenario: name provided without agent flag
- **WHEN** the user runs `sap-iac init my-project` in a TTY
- **THEN** an interactive multi-select prompt is shown to select agents

#### Scenario: no name in a terminal
- **WHEN** the user runs `sap-iac init` in a TTY
- **THEN** the CLI presents the initialization-mode menu

#### Scenario: no name in a non-interactive environment
- **WHEN** the user runs `sap-iac init` without a usable TTY
- **THEN** initialization exits non-zero because its mode cannot be selected interactively

### Requirement: optional global-account subdomain
For fresh and adopted infrastructure projects, the CLI SHALL offer an optional global-account subdomain input and write the selected value to `memory/global-account.md`. GUID-shaped values SHALL be accepted because they can be valid global-account subdomains. Leaving the input blank SHALL be accepted and SHALL create an empty subdomain record for a new project. The record SHALL be ignored by Git. Agent-only updates SHALL not prompt for or change this record.

#### Scenario: subdomain supplied
- **WHEN** a user enters `acme-global` during initialization
- **THEN** `memory/global-account.md` contains `- Subdomain: acme-global`

#### Scenario: subdomain skipped
- **WHEN** a user leaves the optional input blank
- **THEN** initialization continues without asking for the subdomain again

### Requirement: write initialization memory files
For fresh and adopted infrastructure projects, the CLI SHALL write `memory/global-account.md` and copy the embedded `memory/service-params-catalogue.yaml`. It SHALL NOT write or overwrite either file during agent-only initialization. If `memory/service-params-catalogue.yaml` already exists during adoption, it SHALL be left unchanged.

#### Scenario: fresh init writes both memory files
- **WHEN** initialization completes in fresh mode
- **THEN** both `memory/global-account.md` and `memory/service-params-catalogue.yaml` exist in the new project directory

#### Scenario: adopt init writes catalogue when absent
- **WHEN** initialization uses adopt mode and `memory/service-params-catalogue.yaml` does not exist
- **THEN** the embedded catalogue is written to `memory/`

#### Scenario: adopt init preserves existing catalogue
- **WHEN** initialization uses adopt mode and `memory/service-params-catalogue.yaml` already exists
- **THEN** the existing catalogue file is left unchanged

#### Scenario: agent-only init does not write memory files
- **WHEN** initialization uses agent-only mode
- **THEN** neither initialization memory file is created or modified

### Requirement: terraform pre-flight check
The CLI SHALL check whether `terraform` is available on `$PATH` before creating files. When it is unavailable, initialization SHALL emit an advisory warning and continue.

#### Scenario: terraform found
- **WHEN** `terraform` is present on `$PATH`
- **THEN** initialization proceeds without the missing-Terraform warning

#### Scenario: terraform not found
- **WHEN** `terraform` is not present on `$PATH`
- **THEN** initialization warns that Terraform is needed for Terraform commands and continues

### Requirement: optional BTP platform-validation pre-flight
The CLI SHALL detect whether `btp` is on PATH and whether each selected agent has a BTP MCP server, then persist the non-sensitive result in `.sap-iac/platform-validation.md`. The record SHALL be ignored by Git and refreshed on fresh, adopt, and agent-only initialization.

#### Scenario: no BTP validation route
- **WHEN** neither the CLI nor a selected agent's BTP MCP server is available
- **THEN** init warns for that agent and completes normally

### Requirement: git repository initialised
For a fresh project, the CLI SHALL run `git init` when Git is available. When Git is unavailable, initialization SHALL emit an advisory warning and continue. Adopt and agent-only modes SHALL not initialize a Git repository.

#### Scenario: git init succeeds
- **WHEN** Git is available and fresh initialization completes
- **THEN** a `.git` directory exists inside the project directory

#### Scenario: git not found
- **WHEN** Git is not present on `$PATH` during fresh initialization
- **THEN** initialization completes and warns the user to run `git init` manually

#### Scenario: existing directory
- **WHEN** initialization uses adopt or agent-only mode
- **THEN** the CLI does not run `git init`

### Requirement: success output shows workflow guide
On successful initialization the CLI SHALL print a structured workflow guide showing the project result, configured agents, and ordered commands. The guide SHALL present governance as optional, list infrastructure-definition and Terraform-generation commands, and include `/sap-iac.next` as a utility.

#### Scenario: success output after fresh init
- **WHEN** `sap-iac init my-project --agent claude` completes in fresh mode
- **THEN** the CLI confirms the project and agent configuration and prints the Getting started, Define your guardrails, Define your infrastructure, Generate Terraform, and Utilities sections
- **AND** `/sap-iac.govern` and `/sap-iac.analyse` are marked optional

#### Scenario: success output after update
- **WHEN** initialization completes in adopt or agent-only mode
- **THEN** the CLI reports that the project was updated and omits the instruction to change into a newly created directory

### Requirement: idempotent on existing directory
When a named directory already exists, the CLI SHALL update it instead of refusing it. A directory containing `specs/` or `memory/` SHALL use agent-only mode. Another existing directory SHALL be offered for adoption. Agent-only mode SHALL refresh selected agent command files and `.sap-iac/platform-validation.md`, ensure managed `.gitignore` entries, and leave `specs/`, `memory/`, and `terraform/` content unchanged.

#### Scenario: existing sap-iac project
- **WHEN** the supplied directory contains `specs/` or `memory/`
- **THEN** the CLI refreshes selected agent files and `.sap-iac/platform-validation.md`, and ensures managed `.gitignore` entries
- **AND** it does not modify `specs/`, `memory/`, or `terraform/` content

#### Scenario: adopt another existing directory
- **WHEN** the supplied directory exists but contains neither `specs/` nor `memory/`
- **THEN** the CLI asks for confirmation before creating the sap-iac project structure in that directory
