# Tasks Command Capability

## Purpose

Defines the behaviour of `/btp-iac:tasks`: consolidating landscape, services, and trust specs into a single dependency-ordered task list with IDs and parallel execution markers.

## Requirements

### Requirement: consolidate specs into task list
The command SHALL read `specs/landscape.md`, `specs/services.md`, and `specs/trust.md` and consolidate them into a single dependency-ordered task list with IDs and parallel execution markers.

#### Scenario: task list produced
- **WHEN** all three input specs exist
- **THEN** the command produces a dependency-ordered task list with unique IDs and parallel execution markers

### Requirement: write tasks file
The command SHALL write the task list to `specs/tasks.md`.

#### Scenario: tasks file written
- **WHEN** the command completes
- **THEN** `specs/tasks.md` exists and is the direct input to `/btp-iac:design` and `/btp-iac:generate`
