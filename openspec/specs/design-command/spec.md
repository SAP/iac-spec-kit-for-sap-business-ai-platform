# Design Command Capability

## Purpose

Defines the behaviour of `/sap-iac.design`: translating the task list into a concrete Terraform folder structure and annotating each task with its target directory and file path.

## Requirements

### Requirement: guide state-chunking boundaries with trade-offs
Before defining the folder structure, the command SHALL guide the user through the Terraform state-chunking architecture. It SHALL present the BTP / Cloud Foundry / Kyma provider split as a fixed, non-negotiable boundary, and SHALL recommend optional finer-grained state boundaries — global account and directory configuration; subaccount provisioning and configuration; entitlements, service instances, and subscriptions; Cloud Foundry organization and space setup; and environment-, team-, or application-specific resources. For each recommended boundary the command SHALL state its trade-offs across deployment independence, ownership boundaries, blast radius, lifecycle differences, and operational overhead.

#### Scenario: boundaries presented with trade-offs
- **WHEN** the command runs on a task set spanning more than one candidate boundary
- **THEN** it presents the fixed BTP/CF/Kyma split plus the recommended finer-grained boundaries, each with its trade-offs (deployment independence, ownership, blast radius, lifecycle, operational overhead)

### Requirement: accept, refine, or override boundaries via free text
The command SHALL let the user accept, refine, or override the recommended state architecture through free-text input, and SHALL respect the resulting structure for the rest of the run and in the persisted decision. It SHALL NOT silently discard or alter a user-chosen boundary. When an override conflicts with the fixed BTP/CF/Kyma split, the command SHALL explain the conflict and ask the user to resolve it rather than overriding their intent.

#### Scenario: free-text override respected
- **WHEN** the user overrides a recommended boundary in free text (e.g. "keep subaccount and services in one state")
- **THEN** the command adopts that structure for the run and records it in `specs/state-architecture.md`

#### Scenario: override conflicts with fixed split
- **WHEN** a user override would merge BTP-provider and Cloud Foundry/Kyma-provider resources into one root
- **THEN** the command explains the conflict and asks the user to resolve it instead of applying the override

### Requirement: persist the approved state architecture
The command SHALL persist the approved state architecture to `specs/state-architecture.md`, recording the stage-modelling mode, one entry per state chunk (its boundary, Terraform root, backend/state type, ownership, provided outputs, and dependencies), the stage ordering (producers before consumers), the rationale and trade-offs, and any assumptions or user overrides. On a re-run the command SHALL read the existing file, summarize it, allow revision, and rewrite it.

#### Scenario: architecture persisted
- **WHEN** the user approves a state architecture
- **THEN** the command writes `specs/state-architecture.md` containing the stage-modelling mode, the state chunks with their boundaries/roots/ownership/dependencies, the stage ordering, and the rationale

#### Scenario: architecture revised on re-run
- **WHEN** `specs/state-architecture.md` already exists and the command is re-run
- **THEN** the command summarizes the existing architecture, lets the user revise it, and rewrites the file

### Requirement: reorganize task annotations around the approved chunks
When the state architecture is created or revised, the command SHALL annotate each task in `specs/tasks.md` with the directory and file path of its chunk/root under the approved architecture, preserving all task metadata, dependencies, and existing Cloud Foundry space-role barriers.

#### Scenario: tasks annotated to approved chunks
- **WHEN** the state architecture is approved or revised
- **THEN** every task's path annotation targets its chunk's Terraform root, and all task metadata and dependencies are preserved

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

### Requirement: place Cloud Foundry runtime-memory entitlements in BTP configuration
The command SHALL annotate a derived `APPLICATION_RUNTIME` / `MEMORY` entitlement task as a BTP-provider resource and preserve its calculated amount metadata. When the subaccount has a Cloud Foundry environment, the annotation SHALL target that subaccount's `btp/` configuration unit rather than its `cf/` configuration unit.

#### Scenario: CF runtime entitlement is annotated for BTP
- **WHEN** a CF-enabled subaccount has an `APPLICATION_RUNTIME` / `MEMORY` entitlement task with amount `1`
- **THEN** the task is annotated to the subaccount's BTP configuration unit and retains amount `1`

### Requirement: surface deprecation warning
For every task whose metadata carries `deprecated = true`, the command SHALL preserve the flag unchanged and print a deprecation warning identifying the service/plan while annotating the task.

#### Scenario: deprecated task annotated
- **WHEN** a task in `specs/tasks.md` has `deprecated = true` in its metadata
- **THEN** the command keeps the flag unchanged and prints a deprecation warning identifying the service/plan
