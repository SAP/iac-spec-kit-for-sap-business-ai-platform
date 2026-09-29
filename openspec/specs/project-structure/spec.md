# Project Structure Capability

## Purpose

Defines the directory and file layout that `sap-iac init` creates and maintains.

## Requirements

### Requirement: canonical directory structure created
The CLI SHALL create `specs/`, `memory/`, and `terraform/`, the selected agent directories, and `.sap-iac/platform-validation.md`. Fresh and adopted projects SHALL contain Git-ignored `memory/global-account.md` and user-editable `memory/service-params-catalogue.yaml`. Adoption SHALL preserve an existing catalogue, and agent-only initialization SHALL not create or modify it.

#### Scenario: base dirs always present
- **WHEN** fresh or adopt initialization completes successfully
- **THEN** `specs/`, `memory/`, and `terraform/` exist regardless of agent selection
- **AND** `.sap-iac/` exists with the local platform-validation record
- **AND** `memory/global-account.md` and `memory/service-params-catalogue.yaml` exist

#### Scenario: agent dirs vary by selection
- **WHEN** only `cursor` is selected
- **THEN** `.cursor/rules/` is created and `.claude/` and `.codex/` are NOT created

#### Scenario: full structure present after init (Claude)
- **WHEN** `sap-iac init <name> --agent claude` completes successfully
- **THEN** `specs/`, `memory/`, `terraform/`, `.sap-iac/`, and `.claude/commands/` exist under `<name>/`
- **AND** the Claude command directory is populated

#### Scenario: full structure present after init (Codex)
- **WHEN** `sap-iac init <name> --agent codex` completes successfully
- **THEN** `specs/`, `memory/`, `terraform/`, `.sap-iac/`, and `.codex/prompts/` exist under `<name>/`
- **AND** the Codex prompt directory is populated

#### Scenario: adoption preserves catalogue
- **WHEN** adopt initialization finds an existing `memory/service-params-catalogue.yaml`
- **THEN** the existing catalogue is left unchanged

#### Scenario: agent-only update leaves catalogue unchanged
- **WHEN** initialization uses agent-only mode
- **THEN** `memory/service-params-catalogue.yaml` is not created or modified

### Requirement: .gitignore created
The CLI SHALL ensure the project `.gitignore` contains Terraform local-state and provider-cache entries, `.sap-iac/platform-validation.md`, and `memory/global-account.md`. It SHALL preserve unrelated existing entries.

#### Scenario: .gitignore present after fresh init
- **WHEN** fresh initialization completes
- **THEN** `.gitignore` contains `.terraform/`, `*.tfstate`, `*.tfstate.backup`, `.terraform.lock.hcl`, `.sap-iac/platform-validation.md`, and `memory/global-account.md`

#### Scenario: existing .gitignore updated
- **WHEN** adopt or agent-only initialization finds an existing `.gitignore`
- **THEN** missing managed entries are appended without removing unrelated entries

### Requirement: success message printed
The CLI SHALL print a human-readable success summary after initialization completes.

#### Scenario: init succeeds
- **WHEN** initialization completes without error
- **THEN** the CLI reports whether the project was created or updated, lists the configured agents, and prints the workflow guide
