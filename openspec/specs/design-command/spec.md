# Design Command Capability

## Purpose

Defines the behaviour of `/btp-iac.design`: translating the task list into a concrete Terraform folder structure and annotating each task with its target file path.

## Requirements

### Requirement: translate task list into Terraform folder structure
The command SHALL read `specs/tasks.md` and define a concrete Terraform folder structure: how resources are split across files, whether modules are introduced, and how environment-specific variable files are organised.

#### Scenario: structure defined
- **WHEN** `specs/tasks.md` is read
- **THEN** the command produces a Terraform folder structure decision for the task set

### Requirement: annotate tasks with file paths
The command SHALL annotate each task in `specs/tasks.md` with the file path it will be written to.

#### Scenario: tasks annotated
- **WHEN** the folder structure is defined
- **THEN** every task in `specs/tasks.md` has a file path annotation indicating which Terraform file will contain it
