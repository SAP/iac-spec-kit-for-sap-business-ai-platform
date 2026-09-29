# platform-name-normalization Specification

## Purpose
Defines the set of natural-language aliases for the SAP platform that are semantically equivalent within this tool, and the rule that agent skills SHALL normalize them before interpreting user intent, while leaving all technical identifiers untouched.

## Requirements

### Requirement: platform aliases are semantically equivalent
The agent skills SHALL treat the following natural-language names as referring to the same platform when interpreting user intent:
- "SAP Business Technology Platform" (and "Business Technology Platform")
- "SAP BTP" (and "BTP" used as a product name in prose)
- "SAP Business AI Platform" (and "Business AI Platform")
- "SAP BAIP" (and "BAIP")

No request using any of these aliases SHALL be rejected, misrouted, or given a response scoped to a different platform.

#### Scenario: user uses legacy brand name
- **WHEN** a user says "set up Business Technology Platform services"
- **THEN** the skill interprets the request identically to "set up BTP services"

#### Scenario: user uses new brand name
- **WHEN** a user says "deploy to SAP Business AI Platform"
- **THEN** the skill interprets the request identically to "deploy to SAP BTP"

#### Scenario: user uses BAIP abbreviation
- **WHEN** a user says "create a BAIP subaccount"
- **THEN** the skill interprets the request identically to "create a BTP subaccount"

### Requirement: technical identifiers are not normalized
The platform alias normalization SHALL apply only to natural-language prose. The following technical identifiers SHALL remain untouched in all generated output and tool invocations:
- BTP CLI command tokens (e.g. `btp list`, `btp target`)
- Terraform provider names (e.g. `btp`, `hashicorp/btp`)
- Terraform resource type prefixes (e.g. `btp_subaccount`, `btp_service_instance`)
- BTP region codes (e.g. `eu10`, `us10`)
- API paths, hostnames, and SDK package names

#### Scenario: generated Terraform resource names are not rewritten
- **WHEN** a skill generates Terraform code after a user prompt using any platform alias
- **THEN** the generated resource types still use `btp_` prefixes, not substitutes

#### Scenario: BTP CLI commands are not rewritten
- **WHEN** a skill invokes a BTP CLI command in response to a user prompt using any platform alias
- **THEN** the CLI invocation still uses `btp list`, `btp target`, and similar canonical tokens

### Requirement: normalization rule is present in every skill
Every embedded agent skill file SHALL include the platform alias normalization rule so the behavior is consistent regardless of which skill is active.

#### Scenario: alias recognized in accounts skill
- **WHEN** the accounts skill is active and the user references "Business AI Platform"
- **THEN** the skill treats the request as if the user had said "BTP"

#### Scenario: alias recognized in generate skill
- **WHEN** the generate skill is active and the user references "SAP BAIP"
- **THEN** the skill treats the request as if the user had said "SAP BTP"
