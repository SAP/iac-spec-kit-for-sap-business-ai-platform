# Tasks Command Capability

## Purpose

Defines the behaviour of `/btp-iac.tasks`: consolidating landscape, services, and trust specs into a single dependency-ordered task list with IDs and parallel execution markers.

## Requirements

### Requirement: consolidate specs into task list
The command SHALL read `specs/landscape.md`, `specs/services.md`, and `specs/trust.md` and, when `specs/connectivity.md` exists, also read `specs/connectivity.md`. It SHALL consolidate all inputs into a single dependency-ordered task list with IDs and parallel execution markers.

#### Scenario: task list produced
- **WHEN** `specs/landscape.md`, `specs/services.md`, and `specs/trust.md` exist
- **THEN** the command produces a dependency-ordered task list with unique IDs and parallel execution markers

#### Scenario: task list produced — without connectivity
- **WHEN** `specs/landscape.md`, `specs/services.md`, and `specs/trust.md` exist and `specs/connectivity.md` does not exist
- **THEN** the command produces a dependency-ordered task list from the three mandatory inputs, with unique IDs and parallel execution markers

#### Scenario: task list produced — with connectivity
- **WHEN** `specs/landscape.md`, `specs/services.md`, `specs/trust.md`, and `specs/connectivity.md` all exist
- **THEN** the command produces a dependency-ordered task list that also includes destination and certificate tasks derived from `specs/connectivity.md`

### Requirement: preserve account environments as tasks
The command SHALL create a task for every Cloud Foundry environment, Kyma environment, and Cloud Foundry space defined in `specs/landscape.md`. Each task SHALL retain its subaccount, resource type, and name. Environment tasks SHALL depend on their subaccount, and Cloud Foundry space tasks SHALL depend on their Cloud Foundry environment.

#### Scenario: landscape contains account environments
- **WHEN** `specs/landscape.md` defines Cloud Foundry or Kyma environments and Cloud Foundry spaces
- **THEN** `specs/tasks.md` contains the corresponding dependency-ordered tasks with their subaccount, type, and name

### Requirement: preserve confirmed subaccount classification metadata
For every subaccount task, the command SHALL copy any confirmed `usage` and `beta_enabled` values from `specs/landscape.md` into structured task metadata. The task metadata is generation input and SHALL NOT alter values downstream. When a legacy landscape omits either value, the command SHALL still create the task without inferring, requesting, or inventing the unavailable metadata.

#### Scenario: classified subaccount task
- **WHEN** `specs/landscape.md` defines a subaccount with confirmed `usage` and `beta_enabled` values
- **THEN** its `btp_subaccount` task contains those values in its task metadata

#### Scenario: legacy subaccount classification metadata absent
- **WHEN** a subaccount in `specs/landscape.md` omits `usage`, `beta_enabled`, or both
- **THEN** the command creates its `btp_subaccount` task without the unavailable metadata and continues

### Requirement: write tasks file
The command SHALL write the task list to `specs/tasks.md`.

#### Scenario: tasks file written
- **WHEN** the command completes
- **THEN** `specs/tasks.md` exists and is the direct input to `/btp-iac.design` and `/btp-iac.generate`

### Requirement: preserve service instance location
The command SHALL preserve each service instance's `location` (`btp` or `cf`), and its `cf_space` when CF-located, from `specs/services.md` on the corresponding task so that `/btp-iac.design` and `/btp-iac.generate` select the correct Terraform provider and CF space.

#### Scenario: btp-located service instance
- **WHEN** a service instance in `specs/services.md` has `location: btp`
- **THEN** its task retains the `btp` location

#### Scenario: cf-located service instance
- **WHEN** a service instance in `specs/services.md` has `location: cf` with a `cf_space`
- **THEN** its task retains the `cf` location and `cf_space`, and depends on that specific Cloud Foundry space task

### Requirement: include connectivity tasks when connectivity spec exists
When `specs/connectivity.md` exists, the command SHALL append a dependency-ordered set of destination and certificate tasks to `specs/tasks.md`. Each destination task SHALL depend on the subaccount it is scoped to. Each certificate task SHALL depend on its subaccount and, when service-instance-scoped, on the relevant service instance task.

#### Scenario: subaccount-level destination task
- **WHEN** `specs/connectivity.md` defines a destination scoped to a subaccount
- **THEN** `specs/tasks.md` contains a task for that destination that depends on the corresponding subaccount task

#### Scenario: service-instance-scoped destination task
- **WHEN** `specs/connectivity.md` defines a destination scoped to a service instance
- **THEN** `specs/tasks.md` contains a task for that destination that depends on both the subaccount task and the service instance task

#### Scenario: certificate task ordering
- **WHEN** `specs/connectivity.md` defines one or more certificates
- **THEN** each certificate has a corresponding task in `specs/tasks.md` ordered after its subaccount task (and service instance task when applicable)

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
