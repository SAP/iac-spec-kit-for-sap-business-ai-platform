# Governance Service Restrictions Capability

## Purpose

Defines the optional `## Service Restrictions` section of `memory/governance.md`: its two input paths (linked Markdown document or free-text inline content), the allowlist/denylist schema, and the rule that service and plan names in this section are not validated against BTP.

## Requirements

### Requirement: collect service restrictions
The `govern` command SHALL ask the user whether org-wide service and service-plan restrictions apply before asking the per-tier Service Plans question.

#### Scenario: user declines restrictions
- **WHEN** the user answers no to the service-restrictions question
- **THEN** the command skips the `## Service Restrictions` section entirely
- **AND** proceeds to the existing per-tier Service Plans question

#### Scenario: user provides a document link
- **WHEN** the user answers yes and supplies a URL to a Markdown document
- **THEN** the command records only the URL under `## Service Restrictions`
- **AND** does not fetch or copy the document contents into `governance.md`

#### Scenario: user provides free-text restrictions
- **WHEN** the user answers yes and enters restrictions as free text
- **THEN** the command records the supplied text inline under `## Service Restrictions`

#### Scenario: user provides both a link and free-text
- **WHEN** the user answers yes and supplies both a URL and additional free-text
- **THEN** the command records the URL and the free-text inline under `## Service Restrictions`

### Requirement: no BTP validation for service restrictions input
The `govern` command SHALL NOT validate service names, plan names, or service-plan pairs supplied via the Service Restrictions input path against the BTP platform.

#### Scenario: service name provided via restrictions
- **WHEN** the user names a BTP service or service-plan pair as part of the Service Restrictions response (link or free-text)
- **THEN** the command accepts the value without issuing a BTP CLI lookup or MCP entitlement check for it

### Requirement: service restrictions schema
When the user declares service restrictions, `memory/governance.md` SHALL contain a `## Service Restrictions` section with the following structure:

```markdown
## Service Restrictions
- Policy source: link | inline | both
- Policy document: <URL, or "N/A" if inline only>
- Allowed services: <comma-separated BTP service names, or "all">
- Allowed service-plan pairs: <service/plan pairs, or "all">
- Denied services: <comma-separated BTP service names, or "none">
- Denied service-plan pairs: <service/plan pairs, or "none">
- Inline policy: |
    <free-text policy content, or omit if policy source is link>
```

#### Scenario: link-only policy
- **WHEN** the user supplies only a URL
- **THEN** `Policy source` is `link`, `Policy document` contains the URL, and `Inline policy` is omitted

#### Scenario: inline-only policy
- **WHEN** the user supplies only free-text
- **THEN** `Policy source` is `inline`, `Policy document` is `N/A`, and `Inline policy` contains the supplied text

#### Scenario: link and inline policy
- **WHEN** the user supplies both a URL and free-text
- **THEN** `Policy source` is `both`, `Policy document` contains the URL, and `Inline policy` contains the free-text
