# CLI Init Capability

## Purpose

Defines the behaviour of the `btp-iac init <name>` command: argument handling, pre-flight checks, and lifecycle guards.

## Requirements

### Requirement: init command accepts a project name
The CLI SHALL accept a single positional argument `<name>` and use it as the new project directory name. The command also accepts an optional `--agent` flag.

#### Scenario: name provided with agent flag
- **WHEN** user runs `btp-iac init my-project --agent claude`
- **THEN** a directory named `my-project` is created and Claude Code files are installed

#### Scenario: name provided without agent flag
- **WHEN** user runs `btp-iac init my-project` in a TTY
- **THEN** an interactive multi-select prompt is shown to select agents

#### Scenario: no name provided
- **WHEN** user runs `btp-iac init` with no argument
- **THEN** the CLI exits with a non-zero status and prints a usage message

### Requirement: terraform pre-flight check
The CLI SHALL verify that `terraform` is available on `$PATH` before creating any files or directories.

#### Scenario: terraform found
- **WHEN** `terraform` is present on `$PATH`
- **THEN** init proceeds normally

#### Scenario: terraform not found
- **WHEN** `terraform` is not present on `$PATH`
- **THEN** the CLI exits with a non-zero status and prints a message directing the user to install Terraform
- **THEN** no directories or files are created

### Requirement: git repository initialised
The CLI SHALL run `git init` inside the newly created project directory.

#### Scenario: git init succeeds
- **WHEN** `git` is available and init completes
- **THEN** a `.git` directory exists inside the project directory

#### Scenario: git not found
- **WHEN** `git` is not present on `$PATH`
- **THEN** the CLI exits with a non-zero status and prints a message directing the user to install Git
- **THEN** no directories or files are created

### Requirement: idempotent on existing directory
The CLI SHALL refuse to overwrite an existing directory with the given name.

#### Scenario: directory already exists
- **WHEN** a directory with the given name already exists in the current working directory
- **THEN** the CLI exits with a non-zero status and prints an error
- **THEN** the existing directory is not modified
