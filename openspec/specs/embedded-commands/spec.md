# Embedded Commands Capability

## Purpose

Defines how the CLI embeds and installs agent command files into new projects.

## Requirements

### Requirement: agent command files embedded in binary
The CLI binary SHALL embed command files for all four supported agents (`claude`, `codex`, `cursor`, `copilot`) using Go's `embed.FS` so no external asset directory is required at runtime.

#### Scenario: binary ships standalone
- **WHEN** the compiled binary is moved to a new machine with no source tree
- **THEN** `btp-iac init` correctly extracts all command files without any external asset path

#### Scenario: all variants present in binary
- **WHEN** the compiled binary is run on any machine
- **THEN** it can install command files for any combination of the four supported agents without external assets

### Requirement: command files written on init
The CLI SHALL write the embedded command files for each selected agent to that agent's directory during `btp-iac init`. Five of the nine files contain substantive agent instructions rather than placeholder content.

#### Scenario: all nine files present after init (Claude)
- **WHEN** `btp-iac init <name> --agent claude` completes successfully
- **THEN** nine `.md` files exist under `<name>/.claude/commands/`:
  - `btp-iac.govern.md`
  - `btp-iac.scenario.md`
  - `btp-iac.analyse.md`
  - `btp-iac.accounts.md`
  - `btp-iac.services.md`
  - `btp-iac.security.md`
  - `btp-iac.tasks.md`
  - `btp-iac.design.md`
  - `btp-iac.generate.md`

#### Scenario: Cursor files present after init with cursor selected
- **WHEN** `btp-iac init <name> --agent cursor` completes successfully
- **THEN** nine `.mdc` files exist under `<name>/.cursor/rules/`

#### Scenario: Codex files present after init with codex selected
- **WHEN** `btp-iac init <name> --agent codex` completes successfully
- **THEN** nine `.md` files exist under `<name>/.codex/prompts/`

#### Scenario: Copilot files present after init with copilot selected
- **WHEN** `btp-iac init <name> --agent copilot` completes successfully
- **THEN** nine `.instructions.md` files exist under `<name>/.github/instructions/`

#### Scenario: govern command has full implementation
- **WHEN** `btp-iac init <name>` completes successfully
- **THEN** the installed `btp-iac.govern` command file contains the full governance collection and file-writing workflow

#### Scenario: downstream commands contain governance validation
- **WHEN** `btp-iac init <name>` completes successfully
- **THEN** the installed `btp-iac.accounts`, `btp-iac.services`, `btp-iac.security`, and `btp-iac.generate` command files each contain a governance validation step

### Requirement: command files are valid markdown
Each embedded command file SHALL be non-empty and readable as UTF-8 text.

#### Scenario: files are readable
- **WHEN** any embedded command file is written to disk
- **THEN** it can be opened and read as plain text without error

### Requirement: shared platform-validation guidance
Every embedded command file SHALL contain the local BTP platform-validation capability contract. Commands requiring live BTP checks SHALL prefer the recorded CLI route and otherwise use a configured BTP MCP route without blocking when neither is available.

#### Scenario: installed command contains capability guidance
- **WHEN** an embedded command is installed during init
- **THEN** it instructs the agent to read `.btp-iac/platform-validation.md` before a required live BTP check
