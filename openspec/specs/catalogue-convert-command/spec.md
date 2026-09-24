# Catalogue Convert Command Capability

## Purpose

Defines the behaviour of `btp-iac catalogue convert`: converting a JSON Schema file into a catalogue YAML entry suitable for appending to `memory/service-params-catalogue.yaml`, with full specification of argument handling, conversion rules, and project-root discovery for the `--append` flag.

## Requirements

### Requirement: Command accepts schema file, service, and plans
`btp-iac catalogue convert <schema.json>` SHALL accept a single positional argument (path to a JSON Schema file) and two required flags: `--service` (the `service_offering_name`) and `--plans` (comma-separated plan names). All three SHALL be required; the command SHALL fail with a usage error if any are missing.

#### Scenario: All arguments provided
- **WHEN** the user runs `btp-iac catalogue convert schema.json --service xsuaa --plans application,broker`
- **THEN** the command reads `schema.json`, converts it, and outputs a catalogue entry

#### Scenario: Missing --service flag
- **WHEN** the user runs `btp-iac catalogue convert schema.json --plans application`
- **THEN** the command exits with a usage error indicating `--service` is required

#### Scenario: Missing schema file argument
- **WHEN** the user runs `btp-iac catalogue convert --service xsuaa --plans application`
- **THEN** the command exits with a usage error indicating the schema file argument is required

### Requirement: JSON Schema properties mapped to required/optional catalogue keys
The command SHALL map JSON Schema `required[]` properties to `parameters.required` entries and all other top-level properties to `parameters.optional` entries. Each entry SHALL carry `key` (the property name), `description` (from the schema `description` field, or synthesized from `type`/`enum` when absent), and `example` (from `default`, then `enum[0]`, then `"<key-name>"`).

#### Scenario: Required property mapped correctly
- **WHEN** a JSON Schema property name appears in the schema's `required` array
- **THEN** the converted entry places it under `parameters.required`

#### Scenario: Optional property mapped correctly
- **WHEN** a JSON Schema property name does not appear in the schema's `required` array
- **THEN** the converted entry places it under `parameters.optional`

#### Scenario: Description synthesized from enum when absent
- **WHEN** a property has no `description` field but has an `enum` array
- **THEN** the synthesized description lists the allowed values

#### Scenario: Example from default value
- **WHEN** a property has a `default` value
- **THEN** `example` is set to that default value

#### Scenario: Example falls back to placeholder
- **WHEN** a property has neither `default` nor `enum`
- **THEN** `example` is set to `"<key-name>"`

### Requirement: Nested objects converted to nested YAML maps
The command SHALL recursively convert nested JSON Schema object properties into nested YAML maps under the `example` field, preserving arbitrary nesting depth.

#### Scenario: Nested object produces nested example map
- **WHEN** a property has `"type": "object"` with its own `properties`
- **THEN** the `example` field for that key is a YAML map, not a string

### Requirement: oneOf and anyOf simplified to first branch
When a property or the schema root uses `oneOf` or `anyOf`, the command SHALL use the first branch as the canonical form and emit a `# simplified from oneOf` (or `anyOf`) YAML comment adjacent to the affected entry.

#### Scenario: oneOf simplified
- **WHEN** a schema property uses `oneOf` with multiple branches
- **THEN** the converted entry uses the first branch and includes a `# simplified from oneOf` comment

### Requirement: $ref resolved inline
The command SHALL resolve `$ref` references that point within the same document (internal refs). It SHALL NOT fetch remote URLs or files outside the provided schema file. An unresolvable ref SHALL be emitted as `example: "<unresolved $ref>"` with a comment.

#### Scenario: Internal $ref resolved
- **WHEN** a schema property uses `$ref` pointing to a definition in the same file
- **THEN** the referenced definition is inlined in the converted entry

#### Scenario: External $ref not followed
- **WHEN** a schema property uses `$ref` pointing to an external URL or file
- **THEN** the command emits `example: "<unresolved $ref>"` with a comment and does not fail

### Requirement: Output to stdout by default, --append writes to catalogue
Without `--append`, the command SHALL print the generated YAML entry to stdout. With `--append`, the command SHALL locate the nearest btp-iac project root by walking up from the current working directory and append the entry to `memory/service-params-catalogue.yaml` in that root. If no project root is found, the command SHALL fail with an error.

#### Scenario: Default output to stdout
- **WHEN** the user runs `btp-iac catalogue convert` without `--append`
- **THEN** the YAML entry is printed to stdout and no file is written

#### Scenario: --append writes to catalogue
- **WHEN** the user runs `btp-iac catalogue convert` with `--append` inside a btp-iac project
- **THEN** the entry is appended to `memory/service-params-catalogue.yaml`

#### Scenario: --append outside project root fails
- **WHEN** the user runs `btp-iac catalogue convert` with `--append` outside any btp-iac project
- **THEN** the command exits with an error indicating no project root was found

### Requirement: Generated entry carries source: user-defined
Entries produced by `btp-iac catalogue convert` SHALL be emitted with `source: user-defined`.

#### Scenario: Converted entry has correct source
- **WHEN** the command converts a JSON Schema and outputs a catalogue entry
- **THEN** the entry contains `source: user-defined`
