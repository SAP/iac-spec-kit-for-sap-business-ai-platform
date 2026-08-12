# Project Structure Capability

## Purpose

Defines the directory and file layout that `btp-iac init` creates in a new project.

## Requirements

### Requirement: canonical directory structure created
The CLI SHALL create the base project directories and the agent-specific directories for each selected agent.

#### Scenario: base dirs always present
- **WHEN** `btp-iac init <name>` completes successfully
- **THEN** `specs/`, `memory/`, and `terraform/` always exist regardless of agent selection

#### Scenario: agent dirs vary by selection
- **WHEN** only `cursor` is selected
- **THEN** `.cursor/rules/` is created and `.claude/` is NOT created

#### Scenario: full structure present after init (Claude)
- **WHEN** `btp-iac init <name> --agent claude` completes successfully
- **THEN** the following paths exist under `<name>/`:
  - `specs/` (directory)
  - `memory/` (directory)
  - `terraform/` (directory)
  - `.claude/commands/btp-iac/` (directory, populated with command files)

### Requirement: .gitignore created
The CLI SHALL write a `.gitignore` file in the project root that excludes Terraform local state and provider cache files.

#### Scenario: .gitignore present after init
- **WHEN** `btp-iac init <name>` completes successfully
- **THEN** `<name>/.gitignore` exists and contains entries for `.terraform/` and `*.tfstate`

### Requirement: success message printed
The CLI SHALL print a human-readable success summary after init completes.

#### Scenario: init succeeds
- **WHEN** `btp-iac init <name>` completes without error
- **THEN** the CLI prints the project name and a list of next steps (e.g. open the project in Claude Code and run `/btp-iac:scenario`)
