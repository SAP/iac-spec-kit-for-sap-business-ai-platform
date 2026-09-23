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

### Requirement: optional global-account subdomain
For fresh and adopted infrastructure projects, the CLI SHALL offer an optional global-account subdomain input and write the selected value to `memory/global-account.md`. GUID-shaped values SHALL be accepted because they can be valid global-account subdomains. Leaving the input blank SHALL be accepted and SHALL create an empty subdomain record for a new project. The record SHALL be ignored by Git. Agent-only updates SHALL not prompt for or change this record.

#### Scenario: subdomain supplied
- **WHEN** a user enters `acme-global` during initialization
- **THEN** `memory/global-account.md` contains `- Subdomain: acme-global`

#### Scenario: subdomain skipped
- **WHEN** a user leaves the optional input blank
- **THEN** initialization continues without asking for the subdomain again

### Requirement: terraform pre-flight check
The CLI SHALL verify that `terraform` is available on `$PATH` before creating any files or directories.

#### Scenario: terraform found
- **WHEN** `terraform` is present on `$PATH`
- **THEN** init proceeds normally

#### Scenario: terraform not found
- **WHEN** `terraform` is not present on `$PATH`
- **THEN** the CLI exits with a non-zero status and prints a message directing the user to install Terraform
- **THEN** no directories or files are created

### Requirement: optional BTP platform-validation pre-flight
The CLI SHALL detect whether `btp` is on PATH and whether each selected agent has a BTP MCP server, then persist the non-sensitive result in `.btp-iac/platform-validation.md`. The record SHALL be ignored by Git and refreshed on fresh, adopt, and agent-only initialization.

#### Scenario: no BTP validation route
- **WHEN** neither the CLI nor a selected agent's BTP MCP server is available
- **THEN** init warns for that agent and completes normally

### Requirement: git repository initialised
The CLI SHALL run `git init` inside the newly created project directory.

#### Scenario: git init succeeds
- **WHEN** `git` is available and init completes
- **THEN** a `.git` directory exists inside the project directory

#### Scenario: git not found
- **WHEN** `git` is not present on `$PATH`
- **THEN** the CLI exits with a non-zero status and prints a message directing the user to install Git
- **THEN** no directories or files are created

### Requirement: success output shows workflow guide
On successful init the CLI SHALL print a structured workflow guide showing the project name, configured agents, and the ordered sequence of slash commands to run inside the AI agent.

#### Scenario: success output after init
- **WHEN** `btp-iac init my-project --agent claude` completes successfully
- **THEN** the CLI prints confirmation of what was created (project, agent, git repository)
- **THEN** the CLI prints a boxed workflow guide with two steps: open the project in the terminal, then run the slash commands inside the AI agent in order
- **THEN** the guide groups commands into three sections: Before you start, Define your infrastructure, and Generate Terraform
- **THEN** optional commands (`/btp-iac.govern`, `/btp-iac.analyse`) are marked as optional

### Requirement: idempotent on existing directory
The CLI SHALL refuse to overwrite an existing directory with the given name.

#### Scenario: directory already exists
- **WHEN** a directory with the given name already exists in the current working directory
- **THEN** the CLI exits with a non-zero status and prints an error
- **THEN** the existing directory is not modified
