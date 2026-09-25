# Design Command Capability

## Purpose

Defines the behaviour of `/sap-iac.design`: translating the task list into a concrete Terraform folder structure and annotating each task with its target directory and file path.

## Requirements

### Requirement: translate task list into Terraform folder structure
The command SHALL read `specs/tasks.md` and define a concrete Terraform folder structure using a standard file layout per configuration unit: `main.tf` (resources), `variables.tf` (input variables), `outputs.tf` (output values), `providers.tf` (provider configuration and `required_providers`), and `backend.tf` (defaulting to a local backend). It SHALL decide how resources are split across these files and how the units are organised across directories according to the selected stage-modelling mode.

#### Scenario: structure defined
- **WHEN** `specs/tasks.md` is read
- **THEN** the command produces a Terraform folder structure decision for the task set using the standard `main.tf` / `variables.tf` / `outputs.tf` / `providers.tf` / `backend.tf` layout per configuration unit

#### Scenario: default backend
- **WHEN** the folder structure is defined
- **THEN** each configuration unit includes a `backend.tf` whose default is a local backend

### Requirement: separate provider resources into directories by type and location
The command SHALL place BTP-provider and Cloud Foundry-provider resources in separate directories when a configuration unit contains a Cloud Foundry or Kyma environment: all BTP-provider resources in a `btp/` subdirectory and all Cloud Foundry- or Kyma-provider resources in a sibling `cf/` or `kyma/` subdirectory. For service instance tasks the split is driven by `location` (`btp` or `cf`). Subscription tasks (`resource_type = btp_subaccount_subscription`) and entitlement-only tasks (`resource_type = btp_subaccount_entitlement`) carry no `location` but are always BTP-provider resources and SHALL be placed in the BTP directory. When no CF or Kyma environment is present, the unit is not split and all resources use the standard layout in a single directory.

#### Scenario: mixed locations
- **WHEN** the task set contains both `btp` and `cf` service instances
- **THEN** the file path annotations place BTP-provider resources under a `btp/` subdirectory and CF-provider resources under a sibling `cf/` subdirectory

#### Scenario: kyma environment
- **WHEN** the task set contains a Kyma environment
- **THEN** BTP-provider resources are placed under a `btp/` subdirectory and Kyma-provider resources under a sibling `kyma/` subdirectory

#### Scenario: subscription tasks
- **WHEN** the task set contains `btp_subaccount_subscription` tasks
- **THEN** those tasks are placed in the BTP directory alongside BTP service instance tasks

#### Scenario: btp-only scenario
- **WHEN** the task set contains no Cloud Foundry or Kyma environment
- **THEN** the unit is not split into `btp/` and `cf/`|`kyma/` subdirectories

### Requirement: annotate tasks with directory and file paths
The command SHALL annotate each task in `specs/tasks.md` with the directory and file path it will be written to under the selected structure and preserve all existing Task metadata and dependencies unchanged.

#### Scenario: tasks annotated
- **WHEN** the folder structure is defined
- **THEN** every task in `specs/tasks.md` has a path annotation indicating which directory and Terraform file will contain it
- **AND** every task retains its existing Task metadata

### Requirement: preserve CF space role barriers in design annotations
When `specs/tasks.md` contains a CF space role barrier, `/sap-iac.design` SHALL preserve every existing dependency and annotate the space task, its `cloudfoundry_space_role` tasks, and each subsequent task scoped to that space into the same Cloud Foundry configuration unit and `main.tf`. It SHALL not reorder, remove, or weaken the role-task dependencies while assigning paths.

#### Scenario: CF dependency chain is co-located
- **WHEN** a CF space task, its role tasks, and a service-instance task scoped to that space are present in `specs/tasks.md`
- **THEN** their annotations target the same CF configuration unit and `main.tf`, and the service-instance task retains dependencies on all of that space's role tasks

#### Scenario: other spaces remain isolated
- **WHEN** tasks target two different CF spaces
- **THEN** design preserves each task's own space-role dependencies and does not add dependencies between the two spaces

### Requirement: select stage-modelling mode from governance or user
The command SHALL determine how stages are modelled by reading `memory/governance.md`. When governance records a stage-modelling choice, the command SHALL use it. When governance is absent or records no choice, the command SHALL prompt the user to choose. The two modes are: (a) **per-stage directories** — one directory per stage, named after the stage, each containing the standard layout; and (b) **single configuration** — one directory whose configuration handles all stages via stage-specific variables in `variables.tf`, with values supplied through per-stage tfvars files named after the stage (e.g. `<stage>.tfvars`).

#### Scenario: mode from governance
- **WHEN** `memory/governance.md` records a stage-modelling choice
- **THEN** the command uses that mode without prompting

#### Scenario: mode prompted
- **WHEN** no stage-modelling choice is available from governance
- **THEN** the command prompts the user to choose between per-stage directories and a single configuration

#### Scenario: per-stage directories
- **WHEN** per-stage-directory mode is selected
- **THEN** the structure places one directory per stage, named after the stage, each containing the standard layout (and, when a CF/Kyma environment is present, the `btp/` + `cf/`|`kyma/` split inside that stage directory)

#### Scenario: single configuration
- **WHEN** single-configuration mode is selected
- **THEN** the structure uses one directory whose `variables.tf` declares stage-specific variables and whose stage values are supplied via `<stage>.tfvars` files

### Requirement: define BTP directory-per-stage layer when directories are used
The command SHALL define a dedicated directory-per-stage configuration (using the standard layout) only when governance or tasks indicate BTP **directories** are used for stages. That configuration's `outputs.tf` SHALL expose the directory ID, which is handed to the BTP configuration's `parent_id`. When BTP directories are not used for stages, no directory-per-stage layer is defined and subaccounts sit directly under the global account.

#### Scenario: directories used for stages
- **WHEN** governance or tasks indicate BTP directories are used for stages
- **THEN** the command defines a directory-per-stage configuration whose output is the directory ID feeding the BTP configuration's `parent_id`

#### Scenario: directories not used
- **WHEN** BTP directories are not used for stages
- **THEN** no directory-per-stage layer is defined
