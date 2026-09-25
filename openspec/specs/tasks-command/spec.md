# Tasks Command Capability

## Purpose

Defines the behaviour of `/sap-iac.tasks`: consolidating landscape, services, and trust specs into a single dependency-ordered task list with IDs and parallel execution markers.

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
The command SHALL create a task for every Cloud Foundry environment, Kyma environment, and Cloud Foundry space defined in `specs/landscape.md`. Each task SHALL retain its subaccount, resource type, and name. Cloud Foundry space tasks SHALL use `resource_type = cloudfoundry_space` and depend on their Cloud Foundry environment task. Environment tasks SHALL depend on their subaccount. Every `btp_subaccount_environment_instance` task SHALL carry `environment_type` in its task metadata, set to `cloudfoundry` for Cloud Foundry environments and `kyma` for Kyma environments.

#### Scenario: landscape contains account environments
- **WHEN** `specs/landscape.md` defines Cloud Foundry or Kyma environments and Cloud Foundry spaces
- **THEN** `specs/tasks.md` contains the corresponding dependency-ordered tasks with their subaccount, type, and name

#### Scenario: CF environment task carries environment_type
- **WHEN** `specs/landscape.md` defines a Cloud Foundry environment for a subaccount
- **THEN** its `btp_subaccount_environment_instance` task metadata contains `environment_type = cloudfoundry`

#### Scenario: Kyma environment task carries environment_type
- **WHEN** `specs/landscape.md` defines a Kyma environment for a subaccount
- **THEN** its `btp_subaccount_environment_instance` task metadata contains `environment_type = kyma`

#### Scenario: CF space task has a concrete resource type
- **WHEN** `specs/landscape.md` defines a Cloud Foundry space
- **THEN** its task metadata contains `resource_type = cloudfoundry_space`, its name and subaccount, and it depends on its Cloud Foundry environment task

### Requirement: preserve confirmed subaccount classification metadata
For every subaccount task, the command SHALL copy `subdomain`, `region`, and any confirmed `usage` and `beta_enabled` values from `specs/landscape.md` into structured task metadata. The task metadata is generation input and SHALL NOT alter values downstream. When a legacy landscape omits `usage` or `beta_enabled`, the command SHALL still create the task without inferring, requesting, or inventing the unavailable metadata. `subdomain` and `region` are always present for a valid subaccount entry and SHALL always be copied.

#### Scenario: classified subaccount task
- **WHEN** `specs/landscape.md` defines a subaccount with confirmed `usage` and `beta_enabled` values
- **THEN** its `btp_subaccount` task contains `subdomain`, `region`, `usage`, and `beta_enabled` in its task metadata

#### Scenario: legacy subaccount classification metadata absent
- **WHEN** a subaccount in `specs/landscape.md` omits `usage`, `beta_enabled`, or both
- **THEN** the command creates its `btp_subaccount` task with `subdomain` and `region` in metadata, omits the unavailable classification fields, and continues

### Requirement: write tasks file
The command SHALL write the task list to `specs/tasks.md`.

#### Scenario: tasks file written
- **WHEN** the command completes
- **THEN** `specs/tasks.md` exists and is the direct input to `/sap-iac.design` and `/sap-iac.generate`

### Requirement: map service consumption type to resource type and preserve provider metadata
The command SHALL read each service entry's `consumption_type` from `specs/services.md` and record the corresponding `resource_type` in that entry's task metadata:

- `instance` → `btp_subaccount_service_instance` (or `cloudfoundry_service_instance` when `location: cf`). The task SHALL carry `location` (`btp` or `cf`) and, when CF-located, `cf_space`. A CF-located instance task SHALL depend on its specific Cloud Foundry space task.
- `subscription` → `btp_subaccount_subscription`. No `location` or `cf_space` applies; subscriptions are always BTP-provider-managed.
- `entitlement-only` → `btp_subaccount_entitlement`. Only the entitlement assignment task is created — no service instance or subscription task.

#### Scenario: btp-located service instance
- **WHEN** a service entry in `specs/services.md` has `consumption_type: instance` and `location: btp`
- **THEN** its task carries `resource_type = btp_subaccount_service_instance` and `location: btp`

#### Scenario: cf-located service instance
- **WHEN** a service entry in `specs/services.md` has `consumption_type: instance` and `location: cf` with a `cf_space`
- **THEN** its task carries `resource_type = cloudfoundry_service_instance`, `location: cf`, and `cf_space`, and depends on that specific Cloud Foundry space task

#### Scenario: subscription service
- **WHEN** a service entry in `specs/services.md` has `consumption_type: subscription`
- **THEN** its task carries `resource_type = btp_subaccount_subscription` with no `location` or `cf_space` fields

#### Scenario: entitlement-only service
- **WHEN** a service entry in `specs/services.md` has `consumption_type: entitlement-only`
- **THEN** the command creates only an entitlement assignment task with `resource_type = btp_subaccount_entitlement` and no instance or subscription task

### Requirement: carry quota_required flag into service task metadata
When a service entry in `specs/services.md` contains `quota_required: true`, the command SHALL copy that flag into the task metadata of the task created for that service entry, regardless of its `resource_type`. When the flag is absent the task metadata SHALL NOT contain it.

#### Scenario: quota_required propagated to service task
- **WHEN** a service entry in `specs/services.md` has `quota_required: true`
- **THEN** the task created for that service entry in `specs/tasks.md` contains `quota_required: true` in its task metadata, regardless of whether its `resource_type` is `btp_subaccount_service_instance`, `cloudfoundry_service_instance`, `btp_subaccount_subscription`, or `btp_subaccount_entitlement`

#### Scenario: quota_required absent — no field in task
- **WHEN** a service entry in `specs/services.md` does not have `quota_required: true`
- **THEN** the task created for that service entry does not contain a `quota_required` field

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
The command SHALL write all tasks with an unchecked checkbox (`- [ ]`). The `/sap-iac.generate` command is responsible for marking tasks complete (`- [x]`) as it generates each one.

#### Scenario: tasks written with unchecked checkboxes
- **WHEN** `specs/tasks.md` is written
- **THEN** every task entry uses `- [ ]` markdown checkbox syntax

### Requirement: stage annotation
Each task SHALL be annotated with the stage(s) it belongs to (e.g. `dev`, `test`, `prod`), derived from the input specs. All tasks are always written to `specs/tasks.md` regardless of stage; the stage annotation is consumed by `/sap-iac.generate` to filter which tasks are transferred into code.

#### Scenario: tasks annotated with stages
- **WHEN** `specs/tasks.md` is written
- **THEN** every task entry includes a `stage` annotation

### Requirement: generate role collection assignment tasks
For each role collection assignment entry in `specs/trust.md`, the command SHALL create one `btp_subaccount_role_collection_assignment` task. Each task SHALL depend on the corresponding `btp_subaccount_role_collection_base` task. Task metadata SHALL include: `resource_type = btp_subaccount_role_collection_assignment`, `subaccount`, `role_collection_name`, and either `user_name` (user assignments) or `group_name` (group assignments). When an `origin` field is present in the trust entry, the task metadata SHALL include `origin`.

#### Scenario: user assignment task created
- **WHEN** a user assignment entry exists in the trust file role collection assignments block
- **THEN** one btp_subaccount_role_collection_assignment task is created with user_name and optional origin, depending on the base task

#### Scenario: group assignment task created
- **WHEN** a group assignment entry exists in the trust file role collection assignments block
- **THEN** one btp_subaccount_role_collection_assignment task is created with group_name and optional origin, depending on the base task

#### Scenario: no assignments in trust file
- **WHEN** no role collection assignment block is present in specs/trust.md
- **THEN** no btp_subaccount_role_collection_assignment tasks are created

### Requirement: generate CF space role tasks
For each CF space role assignment entry in `specs/trust.md`, the command SHALL create one `cloudfoundry_space_role` task. Each entry represents one `(space_name, username, role)` tuple. Each task SHALL depend on the Cloud Foundry space task for the named space. Task metadata SHALL include: `resource_type = cloudfoundry_space_role`, `space_name`, `username`, `role_type`, `origin`.

#### Scenario: CF space role task created per tuple
- **WHEN** a CF space user assignment entry exists in specs/trust.md
- **THEN** one cloudfoundry_space_role task is created for that (space_name, username, role) tuple, depending on the CF space task

#### Scenario: dependency on CF space task
- **WHEN** generating a cloudfoundry_space_role task
- **THEN** the task depends on the Cloud Foundry space task whose name matches space_name

#### Scenario: no CF space assignments in trust file
- **WHEN** no CF space user assignments block is present in specs/trust.md
- **THEN** no cloudfoundry_space_role tasks are created

### Requirement: gate space-scoped tasks on CF space-role assignments
When a task set creates a Cloud Foundry space and contains one or more other Cloud Foundry-provider tasks scoped to that same space, `/sap-iac.tasks` SHALL place every `cloudfoundry_space_role` task for that space between the space task and each other space-scoped task. A task is space-scoped when its metadata identifies that Cloud Foundry space (currently `cf_space` for `cloudfoundry_service_instance` tasks); `cloudfoundry_space_role` itself is not an other space-scoped task. Each such task SHALL depend on the space task and on every role task for that space.

If a created space has another space-scoped task but `specs/trust.md` contains no CF space-role assignment for it, the command SHALL stop without writing a dependency-incomplete task list and instruct the user to record the needed assignments with `/sap-iac.security`. BTP-provider, organization-scoped, and resources not scoped to a created space SHALL retain their existing dependencies.

#### Scenario: CF service instance waits for all role assignments
- **WHEN** a created CF space has two `cloudfoundry_space_role` tasks and a `cloudfoundry_service_instance` task scoped to that space
- **THEN** the service-instance task depends on the space task and both role tasks, producing the order `space -> roles -> service instance`

#### Scenario: another space-scoped resource waits for its own space roles
- **WHEN** a created CF space has one or more role tasks and a generated Cloud Foundry-provider task whose metadata scopes it to that space
- **THEN** that task depends on every role task for that same space and does not depend on role tasks for another space

#### Scenario: required space roles are absent
- **WHEN** a created CF space has another space-scoped task but no CF space-role assignment in `specs/trust.md`
- **THEN** the command stops and directs the user to `/sap-iac.security` before writing `specs/tasks.md`

#### Scenario: organization-scoped and BTP resources are unchanged
- **WHEN** tasks are organization-scoped, BTP-provider-managed, or not scoped to a created CF space
- **THEN** the command does not add CF space-role dependencies to those tasks
