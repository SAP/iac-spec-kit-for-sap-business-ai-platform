# Agent Selection Capability

## Purpose

Defines how the CLI resolves which AI agents to configure during `btp-iac init`, via flag or interactive prompt.

## Requirements

### Requirement: agent flag accepted by init
The CLI SHALL accept an `--agent` flag on the `init` subcommand taking a comma-separated list of agent IDs.

#### Scenario: single agent via flag
- **WHEN** user runs `btp-iac init my-project --agent claude`
- **THEN** only Claude Code command files are installed

#### Scenario: multiple agents via flag
- **WHEN** user runs `btp-iac init my-project --agent claude,cursor`
- **THEN** command files for both Claude Code and Cursor are installed

#### Scenario: invalid agent ID
- **WHEN** user provides an unrecognised agent ID via `--agent`
- **THEN** the CLI exits with a non-zero status and lists the supported agent IDs

### Requirement: interactive prompt when flag is omitted
The CLI SHALL display an interactive multi-select prompt when `--agent` is not provided and stdin is a TTY.

#### Scenario: user selects agents interactively
- **WHEN** `btp-iac init my-project` is run without `--agent` in a TTY
- **THEN** a multi-select prompt lists `claude`, `codex`, `cursor`, and `copilot`
- **AND** the user may select one or more agents with space and confirm with enter
- **AND** init proceeds with the selected agents

#### Scenario: no agent selected
- **WHEN** the user confirms the prompt without selecting any agent
- **THEN** the CLI exits with a non-zero status and informs the user that at least one agent must be selected

#### Scenario: no TTY detected
- **WHEN** `btp-iac init my-project` is run without `--agent` and stdin is not a TTY
- **THEN** the CLI exits with a non-zero status directing the user to use `--agent`

### Requirement: supported agent IDs
The CLI SHALL recognise exactly four agent IDs: `claude`, `codex`, `cursor`, `copilot`.

#### Scenario: complete supported set
- **WHEN** `btp-iac init --agent claude,codex,cursor,copilot` is run
- **THEN** all four agents are configured in the new project
