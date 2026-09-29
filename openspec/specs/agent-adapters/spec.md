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
