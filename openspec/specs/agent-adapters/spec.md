# Agent Adapters Capability

## Purpose

Defines the per-agent directory targets, file name prefix, and file extensions used when installing command files.

## Requirements

### Requirement: Claude Code adapter
The CLI SHALL install command files for Claude Code into `.claude/commands/` as `sap-iac.<command>.md` files.

#### Scenario: Claude files written
- **WHEN** `claude` is a selected agent
- **THEN** `.claude/commands/sap-iac.<command>.md` files exist in the project
- **THEN** commands are invocable as `/sap-iac.<command>` in Claude Code

### Requirement: Codex adapter
The CLI SHALL install command files for Codex into `.codex/prompts/` as `sap-iac.<command>.md` files.

#### Scenario: Codex files written
- **WHEN** `codex` is a selected agent
- **THEN** `.codex/prompts/sap-iac.<command>.md` files exist in the project
- **THEN** prompts are available to Codex from the project-local prompt directory

### Requirement: Cursor adapter
The CLI SHALL install command files for Cursor into `.cursor/rules/` as `sap-iac.<command>.mdc` files.

#### Scenario: Cursor files written
- **WHEN** `cursor` is a selected agent
- **THEN** `.cursor/rules/sap-iac.<command>.mdc` files exist in the project
- **THEN** rules are referenceable as `@sap-iac.<command>.mdc` in Cursor

### Requirement: GitHub Copilot adapter
The CLI SHALL install command files for GitHub Copilot into `.github/instructions/` as `sap-iac.<command>.instructions.md` files.

#### Scenario: Copilot files written
- **WHEN** `copilot` is a selected agent
- **THEN** `.github/instructions/sap-iac.<command>.instructions.md` files exist in the project
- **THEN** instructions are referenceable as `@sap-iac.<command>.instructions.md` in Copilot

### Requirement: consistent sap-iac. prefix across all agents
All agents SHALL use the `sap-iac.` prefix on command filenames so the invocation pattern is consistent regardless of agent.

#### Scenario: prefix applied uniformly
- **WHEN** any agent is selected
- **THEN** every installed command file is named `sap-iac.<command>.<ext>`

### Requirement: adapters are independent
Each adapter SHALL create only its own directories and files without affecting other adapters.

#### Scenario: single adapter does not create other dirs
- **WHEN** only `cursor` is selected
- **THEN** `.claude/`, `.codex/`, and `.github/instructions/` directories are NOT created

### Requirement: Gemini CLI command adapter
When `gemini` is selected, the CLI SHALL install one project-local command per workflow command at `.gemini/commands/sap-iac.<command>.toml`. Each file SHALL be valid TOML with a non-empty string `prompt` containing the substantive workflow instructions. Commands SHALL be invocable in Gemini CLI as `/sap-iac.<command>`.

#### Scenario: Gemini command files written
- **WHEN** initialization with Gemini selected completes successfully
- **THEN** all eleven workflow command TOML files exist under `.gemini/commands/`
- **AND** Gemini CLI can discover names such as `/sap-iac.scenario` and `/sap-iac.generate`

#### Scenario: command content survives serialization
- **WHEN** embedded instructions contain multiline Markdown, quotes, backslashes, or Unicode
- **THEN** the generated file parses as TOML and its decoded prompt retains the adapted instructions

### Requirement: Gemini CLI native skill adapter
When `gemini` is selected, the CLI SHALL also install all eleven native skills at `.gemini/skills/sap-iac-<command>/SKILL.md`. Each skill SHALL contain valid YAML frontmatter with a name matching its directory and a non-empty description, its substantive instructions, and its bundled embedded runtime resources in the corresponding skill directory.

#### Scenario: native skills discoverable
- **WHEN** initialization with Gemini selected completes successfully
- **THEN** `.gemini/skills/sap-iac-scenario/SKILL.md` and the other ten skill entry points exist
- **AND** Gemini CLI can discover their names and descriptions as workspace skills

#### Scenario: resources retain their locations
- **WHEN** an embedded skill has a bundled runtime resource
- **THEN** the resource is installed under the corresponding Gemini skill directory with its relative location preserved

### Requirement: Gemini instruction compatibility
Gemini slash commands and native skills SHALL provide equivalent workflow instructions adapted to Gemini tool capabilities. Documentation retrieval SHALL preserve the existing MCP-first preference and use Gemini web retrieval as its fallback. MCP lookup guidance SHALL identify configured servers without requiring another agent's tool naming convention.

#### Scenario: Gemini web fallback
- **WHEN** a Gemini workflow requires documentation and its preferred documentation MCP server is unavailable
- **THEN** its instructions direct Gemini to use its `web_fetch` capability

#### Scenario: shared behavior retained
- **WHEN** a workflow is installed as a Gemini command and native skill
- **THEN** both retain the same substantive collection, validation, and output behavior
- **AND** both retain applicable governance and BTP platform-validation guidance

### Requirement: Gemini adapter independence
The Gemini adapter SHALL write only its project-local command and skill outputs. It SHALL NOT install files under the user's home directory or modify unrelated agent files or Gemini settings. Existing agent adapters SHALL retain their established output formats and locations.

#### Scenario: Gemini-only installation
- **WHEN** only Gemini is selected in a fresh project
- **THEN** the agent integration directories are `.gemini/commands/` and `.gemini/skills/`
- **AND** other agent integration directories and global Gemini files are not created

#### Scenario: mixed-agent installation
- **WHEN** Gemini and Claude Code are selected
- **THEN** Gemini TOML files and native skills coexist with Claude Markdown commands
- **AND** Claude command content and format retain their established behavior
