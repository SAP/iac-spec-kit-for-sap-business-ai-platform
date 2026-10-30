# Embedded Commands Capability

## Purpose

Defines how the CLI embeds and installs agent command files into new projects.

## Requirements

### Requirement: agent command files embedded in binary
The CLI binary SHALL embed the workflow instruction assets needed for all five supported agents (`claude`, `codex`, `cursor`, `copilot`, `gemini`) using Go's `embed.FS` so no external asset directory is required at runtime. Gemini commands and native skill packages SHALL be installable from these embedded assets.

#### Scenario: binary ships standalone
- **WHEN** the compiled binary is moved to a new machine with no source tree
- **THEN** `sap-iac init` correctly extracts all command files without any external asset path

#### Scenario: all variants present in binary
- **WHEN** the compiled binary is run on any machine
- **THEN** it can install command files for any combination of the five supported agents without external assets
- **AND** selecting Gemini also installs native skills and their embedded runtime resources

### Requirement: command files written on init
The CLI SHALL write the embedded command files for each selected agent to that agent's directory during `sap-iac init`. Every command file contains substantive agent instructions.

#### Scenario: all eleven files present after init (Claude)
- **WHEN** `sap-iac init <name> --agent claude` completes successfully
- **THEN** eleven `.md` files exist under `<name>/.claude/commands/`:
  - `sap-iac.govern.md`
  - `sap-iac.scenario.md`
  - `sap-iac.analyse.md`
  - `sap-iac.accounts.md`
  - `sap-iac.services.md`
  - `sap-iac.security.md`
  - `sap-iac.connectivity.md`
  - `sap-iac.next.md`
  - `sap-iac.tasks.md`
  - `sap-iac.design.md`
  - `sap-iac.generate.md`

#### Scenario: Cursor files present after init with cursor selected
- **WHEN** `sap-iac init <name> --agent cursor` completes successfully
- **THEN** eleven `.mdc` files exist under `<name>/.cursor/rules/`

#### Scenario: Codex files present after init with codex selected
- **WHEN** `sap-iac init <name> --agent codex` completes successfully
- **THEN** eleven `.md` files exist under `<name>/.codex/prompts/`

#### Scenario: Copilot files present after init with copilot selected
- **WHEN** `sap-iac init <name> --agent copilot` completes successfully
- **THEN** eleven `.instructions.md` files exist under `<name>/.github/instructions/`

#### Scenario: govern command has full implementation
- **WHEN** `sap-iac init <name>` completes successfully
- **THEN** the installed `sap-iac.govern` command file contains the full governance collection and file-writing workflow

#### Scenario: downstream commands contain governance validation
- **WHEN** `sap-iac init <name>` completes successfully
- **THEN** the installed `sap-iac.accounts`, `sap-iac.services`, `sap-iac.security`, and `sap-iac.generate` command files each contain a governance validation step

### Requirement: command files are valid markdown
Embedded skill entry points SHALL remain non-empty UTF-8 Markdown. Installed files SHALL be non-empty UTF-8 text valid for the selected agent's format: Gemini command definitions SHALL be valid TOML, Gemini native skill entry points SHALL be Markdown with skill metadata, and existing agents SHALL retain their established formats.

#### Scenario: files are readable
- **WHEN** any embedded command file is written to disk
- **THEN** it can be opened and read as plain text without error

#### Scenario: Gemini files use their required formats
- **WHEN** Gemini workflow assets are installed
- **THEN** command definitions parse as TOML with non-empty prompts
- **AND** native skill entry points contain Markdown and valid name/description frontmatter

### Requirement: shared platform-validation guidance
Every embedded command file that invokes BTP operations SHALL contain the local BTP platform-validation capability contract and a mandatory BTP read-only boundary. Commands that do not invoke BTP operations are exempt. Commands requiring live BTP checks SHALL prefer the recorded CLI route and otherwise use a configured BTP MCP read/list route without blocking when neither is available. The commands SHALL never invoke BTP mutations; `btp target --global-account <subdomain>` is permitted only as an account-selection prelude to BTP CLI read/list commands.

#### Scenario: installed BTP-invoking command contains capability guidance
- **WHEN** an embedded command that performs BTP operations is installed during init
- **THEN** it instructs the agent to read `.sap-iac/platform-validation.md` before a required live BTP check

#### Scenario: sap-iac-next exempt from BTP boundary
- **WHEN** `sap-iac.next` is installed during init
- **THEN** it does not contain BTP platform-validation boilerplate, as it performs no BTP operations

### Requirement: all Gemini workflows installed
Gemini installation SHALL provide command definitions and native skills for `govern`, `scenario`, `analyse`, `accounts`, `services`, `security`, `connectivity`, `tasks`, `design`, `generate`, and `next`, from the shared embedded workflow instructions. Both forms SHALL retain all substantive workflow instructions after Gemini compatibility adaptation.

#### Scenario: complete Gemini output
- **WHEN** initialization with Gemini selected completes successfully
- **THEN** eleven `.gemini/commands/sap-iac.<command>.toml` files exist
- **AND** eleven `.gemini/skills/sap-iac-<command>/SKILL.md` files exist
- **AND** no workflow is omitted or replaced with placeholder instructions

#### Scenario: governed BTP workflows retain guidance
- **WHEN** Gemini accounts, services, security, and generate workflows are installed
- **THEN** both command prompts and native skills retain their governance validation and applicable platform-validation instructions
- **AND** the next utility retains its existing exemption from BTP boilerplate
