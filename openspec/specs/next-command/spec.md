# Next Command Capability

## Purpose

Defines the behaviour of `/sap-iac.next`: inspecting local project state to recommend the next command in the greenfield BTP IaC flow, both as a standalone slash command and as a footer reference from greenfield-flow skills.

## Requirements

### Requirement: locate project root
Before inspecting project state, the command SHALL walk upward from the current working directory until it finds the first ancestor containing `specs/`, `memory/`, and `terraform/` subdirectories. It SHALL resolve every subsequent state check relative to that directory.

#### Scenario: project root located from a descendant directory
- **WHEN** the command is invoked from a descendant of a directory containing `specs/`, `memory/`, and `terraform/`
- **THEN** it uses that ancestor as the project root for every state check

#### Scenario: project root not found
- **WHEN** neither the current working directory nor any ancestor contains all of `specs/`, `memory/`, and `terraform/`
- **THEN** it reports: ``Could not locate a sap-iac project root. Run this command from within a project created by `sap-iac init`.``
- **AND** it stops without printing a state summary or recommendation

### Requirement: inspect project state
The command SHALL determine the current position in the greenfield flow by checking the existence of the following files in order: `memory/governance.md`, `specs/scenario.md`, `specs/landscape.md`, `specs/services.md`, `specs/trust.md`, `specs/connectivity.md`, `specs/tasks.md`, and at least one `terraform/**/*.tf` file. The generated-code check SHALL match `.tf` files in any subdirectory of `terraform/` (recursive glob), because the design/generate layout places files under provider and stage subdirectories (e.g. `terraform/btp/…`, `terraform/<stage>/btp/…`). The generated-code check SHALL match `.tf` files in any subdirectory of `terraform/` (recursive glob), because the design/generate layout places files under provider and stage subdirectories (e.g. `terraform/btp/…`, `terraform/<stage>/btp/…`).

#### Scenario: no files exist
- **WHEN** none of the inspected files exist
- **THEN** the command recommends `/sap-iac.govern` as the first step

#### Scenario: governance present, scenario missing
- **WHEN** `memory/governance.md` exists and `specs/scenario.md` does not
- **THEN** the command recommends `/sap-iac.scenario`

#### Scenario: scenario present, landscape missing
- **WHEN** `specs/scenario.md` exists and `specs/landscape.md` does not
- **THEN** the command recommends `/sap-iac.accounts`

#### Scenario: landscape present, services or trust missing
- **WHEN** `specs/landscape.md` exists and either `specs/services.md` or `specs/trust.md` is absent
- **THEN** the command recommends the missing step(s) in order: `/sap-iac.services` before `/sap-iac.security`

#### Scenario: services and trust present, connectivity and tasks absent
- **WHEN** `specs/services.md` and `specs/trust.md` exist, and neither `specs/connectivity.md` nor `specs/tasks.md` exists
- **THEN** the command recommends `/sap-iac.connectivity` as the next step

#### Scenario: connectivity present, tasks absent
- **WHEN** `specs/connectivity.md` exists and `specs/tasks.md` does not
- **THEN** the command recommends `/sap-iac.tasks`

#### Scenario: tasks present, no terraform files
- **WHEN** `specs/tasks.md` exists and no `terraform/**/*.tf` files exist
- **THEN** the command recommends `/sap-iac.design`

#### Scenario: terraform files present
- **WHEN** at least one `terraform/**/*.tf` file exists
- **THEN** the command outputs a completion message indicating Terraform code is ready to review and apply

### Requirement: standalone output format
When invoked directly as `/sap-iac.next`, the command SHALL print the current state (which files were found) followed by the recommendation.

#### Scenario: standalone invocation
- **WHEN** the user runs `/sap-iac.next` directly
- **THEN** the output lists which spec files exist, states the current position in the flow, and gives the next command with a one-line description of what it does

### Requirement: footer reference from greenfield-flow skills
Each skill in the linear greenfield flow (`sap-iac.govern`, `sap-iac.scenario`, `sap-iac.accounts`, `sap-iac.services`, `sap-iac.security`, `sap-iac.connectivity`, `sap-iac.tasks`, `sap-iac.design`, and `sap-iac.generate`) SHALL end with a `## Next step` section containing a hardcoded one-liner naming its direct successor. The `sap-iac.analyse` utility and standalone `sap-iac.next` command are excluded.

#### Scenario: linear successor footer
- **WHEN** a non-terminal skill in that linear greenfield flow (`sap-iac.govern`, `sap-iac.scenario`, `sap-iac.accounts`, `sap-iac.services`, `sap-iac.security`, `sap-iac.connectivity`, `sap-iac.tasks`, or `sap-iac.design`) completes
- **THEN** its output ends with a single line of the form: `Next: /<successor> — <one-line description>`

#### Scenario: branching footer at security
- **WHEN** `sap-iac.security` completes
- **THEN** its output ends with: `Next: /sap-iac.connectivity — define destinations and certificates`

#### Scenario: terminal footer at generate
- **WHEN** `sap-iac.generate` completes
- **THEN** its output ends with a completion message indicating the Terraform code is in `terraform/` and ready to review, commit, and apply

### Requirement: no BTP operations
The command SHALL NOT invoke any BTP CLI command or BTP MCP tool. It inspects only the local filesystem.

#### Scenario: offline execution
- **WHEN** the command runs with no network access and no BTP credentials
- **THEN** it completes successfully and produces a recommendation based solely on local file state
