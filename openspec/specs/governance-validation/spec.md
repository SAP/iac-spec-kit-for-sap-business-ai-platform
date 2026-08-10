# Governance Validation Capability

## Purpose

Defines how downstream commands (`/btp-iac.accounts`, `/btp-iac.services`, `/btp-iac.security`, `/btp-iac.generate`) read and enforce governance rules from `memory/governance.md`.

## Requirements

### Requirement: governance file read before generating output
Each downstream command (`/btp-iac.accounts`, `/btp-iac.services`, `/btp-iac.security`, `/btp-iac.generate`) SHALL read `memory/governance.md` at the start of execution if it exists.

#### Scenario: governance file present
- **WHEN** `memory/governance.md` exists
- **THEN** the command loads all governance rules before making any decisions

#### Scenario: governance file absent
- **WHEN** `memory/governance.md` does not exist
- **THEN** the command proceeds without governance constraints and notes that no governance rules are in effect

### Requirement: hard block on violation
A downstream command SHALL stop and report a violation when a decision contradicts a governance rule, unless override is active.

#### Scenario: region violation blocked
- **WHEN** a subaccount is being placed in a region not in the allowed list
- **THEN** the command stops with a message identifying the forbidden region, the governance rule violated, and instructions to either change the region or set `- Override: true` in governance.md

#### Scenario: naming violation blocked
- **WHEN** a generated subaccount name does not match the required pattern
- **THEN** the command stops with a message showing the expected pattern and the non-conforming name

#### Scenario: service plan violation blocked
- **WHEN** a service instance is assigned a plan not permitted for its environment tier
- **THEN** the command stops with a message identifying the plan, the tier, and the permitted plans for that tier

#### Scenario: security violation blocked
- **WHEN** a configuration decision contradicts a security rule (e.g. no custom IdP configured when required)
- **THEN** the command stops with a message identifying the specific security rule violated

### Requirement: override bypasses hard block
When `memory/governance.md` contains `- Override: true`, downstream commands SHALL warn about violations but continue rather than blocking.

#### Scenario: override active with violation
- **WHEN** a governance violation is detected and `- Override: true` is set
- **THEN** the command logs a warning identifying the violation and the override
- **AND** continues generating output
