# Service Params Catalogue Capability

## Purpose

Defines the format and lifecycle of `memory/service-params-catalogue.yaml`, the per-service parameter catalogue shipped with every sap-iac project, including its YAML entry shape, source attribution, and the rules governing how the `/sap-iac.services` skill reads and applies it to produce a `parameters:` block in `specs/services.md`.

## Requirements

### Requirement: Catalogue file written on init
The `sap-iac init` command SHALL write `memory/service-params-catalogue.yaml` from the embedded binary during fresh and adopt initialization. It SHALL NOT write or overwrite the file during agent-only initialization. If the file already exists in an adopt scenario, it SHALL be left unchanged.

#### Scenario: Fresh init writes catalogue
- **WHEN** the user runs `sap-iac init` in fresh mode
- **THEN** `memory/service-params-catalogue.yaml` exists in the new project directory with the builtin entries from the embedded binary

#### Scenario: Adopt init writes catalogue when absent
- **WHEN** the user runs `sap-iac init` in adopt mode and no catalogue file exists
- **THEN** `memory/service-params-catalogue.yaml` is written to `memory/`

#### Scenario: Adopt init preserves existing catalogue
- **WHEN** the user runs `sap-iac init` in adopt mode and `memory/service-params-catalogue.yaml` already exists
- **THEN** the existing file is left unchanged

#### Scenario: Agent-only init does not touch catalogue
- **WHEN** the user runs `sap-iac init` in agent-only mode
- **THEN** `memory/service-params-catalogue.yaml` is not created or modified

### Requirement: Catalogue YAML entry shape
Each entry in `memory/service-params-catalogue.yaml` SHALL conform to the following shape:

```yaml
- service: <service_offering_name>
  plans: [<plan1>, <plan2>]
  source: builtin | user-defined
  parameters:
    required:
      - key: <param-key>
        description: "<human-readable description of the parameter and its valid values>"
        example: <scalar value or nested YAML map>
    optional:
      - key: <param-key>
        description: "<human-readable description>"
        example: <scalar value or nested YAML map>
```

`service` SHALL match the `service_offering_name` used in BTP. `plans` SHALL list the plan names for which these parameters apply. `source` SHALL be `builtin` for entries shipped with the binary and `user-defined` for entries added by the user. `parameters.required` and `parameters.optional` SHALL each be a list; either list MAY be empty. The `example` field for a nested object SHALL be a YAML map, supporting arbitrary nesting depth.

#### Scenario: Entry with flat required parameters
- **WHEN** a catalogue entry has two required scalar parameters
- **THEN** each appears under `parameters.required` with `key`, `description`, and `example` fields

#### Scenario: Entry with nested optional parameter
- **WHEN** a catalogue entry has an optional parameter whose value is a JSON object
- **THEN** the `example` field for that parameter is a nested YAML map, not a string

#### Scenario: Entry scoped to specific plans
- **WHEN** a service has different parameter requirements per plan
- **THEN** separate catalogue entries with different `plans` lists are used to represent them

### Requirement: User-defined entries treated identically to builtin
The `/sap-iac.services` skill SHALL treat entries with `source: user-defined` identically to entries with `source: builtin` when matching and prompting. The `source` field is informational only and SHALL NOT affect lookup or prompting behaviour.

#### Scenario: User-defined entry matched
- **WHEN** a service instance task matches a `user-defined` catalogue entry
- **THEN** the skill prompts for parameters exactly as it would for a `builtin` entry

### Requirement: Catalogue lookup in /services
Before prompting about consumption type, the `/sap-iac.services` skill SHALL read `memory/service-params-catalogue.yaml` from the sap-iac project root. For each service instance task:
- If the task's `service_offering_name` and plan name match a catalogue entry, the skill SHALL prompt the user for each required parameter value (one key at a time), present the description and example, then offer optional parameters one at a time ("add this optional parameter?"). The collected values SHALL be written to the task's entry in `specs/services.md` as a `parameters:` block.
- If no catalogue entry matches, the skill SHALL proceed silently with no `parameters:` block for that task.

#### Scenario: Matching entry triggers parameter prompting
- **WHEN** a service instance task's offering and plan match a catalogue entry with two required keys and one optional key
- **THEN** the skill prompts for both required keys and offers the optional key before writing `specs/services.md`

#### Scenario: Non-matching service produces no parameters block
- **WHEN** a service instance task's offering and plan do not match any catalogue entry
- **THEN** the skill writes the `specs/services.md` entry for that task with no `parameters:` block

#### Scenario: User declines optional parameter
- **WHEN** the skill offers an optional parameter and the user declines
- **THEN** that key is omitted from the `parameters:` block in `specs/services.md`

### Requirement: Parameters block written to services.md
When at least one parameter value has been collected for a service instance task, the `/sap-iac.services` skill SHALL write a `parameters:` block on that task's entry in `specs/services.md` containing the user-supplied key-value pairs. The block SHALL carry final user-supplied values, not the catalogue scaffold or placeholders.

#### Scenario: Parameters block present for matched service
- **WHEN** the user supplies values for all required parameters of a matched service
- **THEN** `specs/services.md` contains a `parameters:` block on that service entry with the supplied values

#### Scenario: No parameters block for unmatched service
- **WHEN** a service has no catalogue entry
- **THEN** `specs/services.md` contains no `parameters:` block on that service entry
