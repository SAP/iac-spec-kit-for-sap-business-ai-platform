# Tasks Command Capability

## Purpose

Defines the behaviour of `/btp-iac.tasks`: consolidating landscape, services, and trust specs into a single dependency-ordered task list with IDs and parallel execution markers.

## Requirements

### Requirement: consolidate specs into task list
The command SHALL read `specs/landscape.md`, `specs/services.md`, and `specs/trust.md` and consolidate them into a single dependency-ordered task list with IDs and parallel execution markers.

#### Scenario: task list produced
- **WHEN** all three input specs exist
- **THEN** the command produces a dependency-ordered task list with unique IDs and parallel execution markers

### Requirement: preserve account environments as tasks
The command SHALL create a task for every Cloud Foundry environment, Kyma environment, and Cloud Foundry space defined in `specs/landscape.md`. Each task SHALL retain its subaccount, resource type, and name. Environment tasks SHALL depend on their subaccount, and Cloud Foundry space tasks SHALL depend on their Cloud Foundry environment.

#### Scenario: landscape contains account environments
- **WHEN** `specs/landscape.md` defines Cloud Foundry or Kyma environments and Cloud Foundry spaces
- **THEN** `specs/tasks.md` contains the corresponding dependency-ordered tasks with their subaccount, type, and name

### Requirement: write tasks file
The command SHALL write the task list to `specs/tasks.md`.

#### Scenario: tasks file written
- **WHEN** the command completes
- **THEN** `specs/tasks.md` exists and is the direct input to `/btp-iac.design` and `/btp-iac.generate`

### Requirement: summarise output in terminal
When the task list is extensive the command SHALL display a short summary in the terminal and instruct the user to open `specs/tasks.md` for the full list.

#### Scenario: large task list summary
- **WHEN** the generated task list is too long to read comfortably inline
- **THEN** the terminal output shows only a count-by-group summary and a reference to `specs/tasks.md`

#### Scenario: small task list inline
- **WHEN** the generated task list is short enough to read comfortably inline
- **THEN** the full list is printed in the terminal

### Requirement: track completion state
The command SHALL write all tasks with an unchecked checkbox (`- [ ]`). The `/btp-iac.generate` command is responsible for marking tasks complete (`- [x]`) as it generates each one.

#### Scenario: tasks written with unchecked checkboxes
- **WHEN** `specs/tasks.md` is written
- **THEN** every task entry uses `- [ ]` markdown checkbox syntax

### Requirement: stage annotation
Each task SHALL be annotated with the stage(s) it belongs to (e.g. `dev`, `test`, `prod`), derived from the input specs. All tasks are always written to `specs/tasks.md` regardless of stage; the stage annotation is consumed by `/btp-iac.generate` to filter which tasks are transferred into code.

#### Scenario: tasks annotated with stages
- **WHEN** `specs/tasks.md` is written
- **THEN** every task entry includes a `stage` annotation
