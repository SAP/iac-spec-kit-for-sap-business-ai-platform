# Agent Adapters Capability

## Purpose

Defines the per-agent directory targets and file extensions used when installing command files.

## Requirements

### Requirement: Claude Code adapter
The CLI SHALL install command files for Claude Code into `.claude/commands/btp-iac/` as `.md` files.

#### Scenario: Claude files written
- **WHEN** `claude` is a selected agent
- **THEN** `.claude/commands/btp-iac/*.md` files exist in the project

### Requirement: Cursor adapter
The CLI SHALL install command files for Cursor into `.cursor/rules/` as `.mdc` files.

#### Scenario: Cursor files written
- **WHEN** `cursor` is a selected agent
- **THEN** `.cursor/rules/*.mdc` files exist in the project

### Requirement: GitHub Copilot adapter
The CLI SHALL install command files for GitHub Copilot into `.github/instructions/` as `.instructions.md` files.

#### Scenario: Copilot files written
- **WHEN** `copilot` is a selected agent
- **THEN** `.github/instructions/*.instructions.md` files exist in the project

### Requirement: adapters are independent
Each adapter SHALL create only its own directories and files without affecting other adapters.

#### Scenario: single adapter does not create other dirs
- **WHEN** only `cursor` is selected
- **THEN** `.claude/` and `.github/instructions/` directories are NOT created
