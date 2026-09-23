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
The command SHALL annotate each task in `specs/tasks.md` with the file path it will be written to and preserve all existing Task metadata unchanged.

#### Scenario: tasks annotated
- **WHEN** the folder structure is defined
- **THEN** every task in `specs/tasks.md` has a file path annotation indicating which Terraform file will contain it
- **AND** every task retains its existing Task metadata

### Requirement: separate provider files by resource type and location
The command SHALL place BTP-provider and Cloud Foundry-provider resources in separate Terraform files. For service instance tasks the split is driven by `location` (`btp` or `cf`). Subscription tasks (`resource_type = btp_subaccount_subscription`) and entitlement-only tasks (`resource_type = btp_subaccount_entitlement`) carry no `location` but are always BTP-provider resources and SHALL be placed in the BTP-provider file.

#### Scenario: mixed locations
- **WHEN** the task set contains both `btp` and `cf` service instances
- **THEN** the file path annotations separate BTP-provider resources from CF-provider resources

#### Scenario: subscription tasks
- **WHEN** the task set contains `btp_subaccount_subscription` tasks
- **THEN** those tasks are placed in the BTP-provider file alongside BTP service instance tasks
