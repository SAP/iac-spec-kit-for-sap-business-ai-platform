# Next Command Capability

## Purpose

Defines the behaviour of `/btp-iac.next`: inspecting local project state to recommend the next command in the greenfield BTP IaC flow, both as a standalone slash command and as a footer reference from greenfield-flow skills.

## Requirements

### Requirement: locate project root
Before inspecting project state, the command SHALL walk upward from the current working directory until it finds the first ancestor containing `specs/`, `memory/`, and `terraform/` subdirectories. It SHALL resolve every subsequent state check relative to that directory.

#### Scenario: project root located from a descendant directory
- **WHEN** the command is invoked from a descendant of a directory containing `specs/`, `memory/`, and `terraform/`
- **THEN** it uses that ancestor as the project root for every state check

#### Scenario: project root not found
- **WHEN** neither the current working directory nor any ancestor contains all of `specs/`, `memory/`, and `terraform/`
- **THEN** it reports: ``Could not locate a btp-iac project root. Run this command from within a project created by `btp-iac init`.``
- **AND** it stops without printing a state summary or recommendation

### Requirement: inspect project state
The command SHALL determine the current position in the greenfield flow by checking the existence of the following files in order: `memory/governance.md`, `specs/scenario.md`, `specs/landscape.md`, `specs/services.md`, `specs/trust.md`, `specs/connectivity.md`, `specs/tasks.md`, and at least one `terraform/**/*.tf` file. The generated-code check SHALL match `.tf` files in any subdirectory of `terraform/` (recursive glob), because the design/generate layout places files under provider and stage subdirectories (e.g. `terraform/btp/…`, `terraform/<stage>/btp/…`). The generated-code check SHALL match `.tf` files in any subdirectory of `terraform/` (recursive glob), because the design/generate layout places files under provider and stage subdirectories (e.g. `terraform/btp/…`, `terraform/<stage>/btp/…`).

#### Scenario: no files exist
- **WHEN** none of the inspected files exist
- **THEN** the command recommends `/btp-iac.govern` as the first step

#### Scenario: governance present, scenario missing
- **WHEN** `memory/governance.md` exists and `specs/scenario.md` does not
- **THEN** the command recommends `/btp-iac.scenario`

#### Scenario: scenario present, landscape missing
- **WHEN** `specs/scenario.md` exists and `specs/landscape.md` does not
- **THEN** the command recommends `/btp-iac.accounts`

#### Scenario: landscape present, services or trust missing
- **WHEN** `specs/landscape.md` exists and either `specs/services.md` or `specs/trust.md` is absent
- **THEN** the command recommends the missing step(s) in order: `/btp-iac.services` before `/btp-iac.security`

#### Scenario: services and trust present, connectivity and tasks absent
- **WHEN** `specs/services.md` and `specs/trust.md` exist, and neither `specs/connectivity.md` nor `specs/tasks.md` exists
- **THEN** the command recommends `/btp-iac.connectivity` (optional) or `/btp-iac.tasks` as a two-option line

#### Scenario: connectivity present, tasks absent
- **WHEN** `specs/connectivity.md` exists and `specs/tasks.md` does not
- **THEN** the command recommends `/btp-iac.tasks`

#### Scenario: tasks present, no terraform files
- **WHEN** `specs/tasks.md` exists and no `terraform/**/*.tf` files exist
- **THEN** the command recommends `/btp-iac.design`

#### Scenario: terraform files present
- **WHEN** at least one `terraform/**/*.tf` file exists
- **THEN** the command outputs a completion message indicating Terraform code is ready to review and apply

### Requirement: standalone output format
When invoked directly as `/btp-iac.next`, the command SHALL print the current state (which files were found) followed by the recommendation.

#### Scenario: standalone invocation
- **WHEN** the user runs `/btp-iac.next` directly
- **THEN** the output lists which spec files exist, states the current position in the flow, and gives the next command with a one-line description of what it does

### Requirement: footer reference from greenfield-flow skills
Each skill in the linear greenfield flow (`btp-iac.govern`, `btp-iac.scenario`, `btp-iac.accounts`, `btp-iac.services`, `btp-iac.security`, `btp-iac.connectivity`, `btp-iac.tasks`, `btp-iac.design`, and `btp-iac.generate`) SHALL end with a `## Next step` section containing a hardcoded one-liner naming its direct successor. The connectivity branch at the `security` step SHALL present both options on one line. The optional `btp-iac.analyse` utility and standalone `btp-iac.next` command are excluded.

#### Scenario: linear successor footer
- **WHEN** a non-branching, non-terminal skill in that linear greenfield flow (`btp-iac.govern`, `btp-iac.scenario`, `btp-iac.accounts`, `btp-iac.services`, `btp-iac.connectivity`, `btp-iac.tasks`, or `btp-iac.design`) completes
- **THEN** its output ends with a single line of the form: `Next: /<successor> — <one-line description>`

#### Scenario: branching footer at security
- **WHEN** `btp-iac.security` completes
- **THEN** its output ends with: `Next: /btp-iac.connectivity (optional) — define destinations and certificates, or /btp-iac.tasks to skip`

#### Scenario: terminal footer at generate
- **WHEN** `btp-iac.generate` completes
- **THEN** its output ends with a completion message indicating the Terraform code is in `terraform/` and ready to review, commit, and apply

### Requirement: no BTP operations
The command SHALL NOT invoke any BTP CLI command or BTP MCP tool. It inspects only the local filesystem.

#### Scenario: offline execution
- **WHEN** the command runs with no network access and no BTP credentials
- **THEN** it completes successfully and produces a recommendation based solely on local file state
