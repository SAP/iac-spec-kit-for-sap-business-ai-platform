---
name: sap-iac-security
description: Identifies authentication and authorisation requirements — IdP trust, role collections, and role collection assignments.
license: Apache-2.0
metadata:
  author: SAP
  version: "1.1"
---

# BTP IaC — Security

Identifies the authentication and authorisation requirements — role collection definitions, user assignments, IdP trust configurations, and role collection assignments.

Reads `specs/scenario.md` and `specs/landscape.md`. Produces `specs/trust.md`.

## Platform Name Normalization

The following names all refer to the same platform and are semantically equivalent for all natural-language interpretation in this skill:
- **SAP Business Technology Platform** (and "Business Technology Platform")
- **SAP BTP** (and "BTP" used as a product name in prose)
- **SAP Business AI Platform** (and "Business AI Platform")
- **SAP BAIP** (and "BAIP")

Treat any of these aliases as identical when interpreting user intent. This normalization applies only to natural-language prose. Technical identifiers remain untouched: BTP CLI command tokens (`btp list`, `btp target`), Terraform provider names (`btp`, `hashicorp/btp`), resource type prefixes (`btp_subaccount`, `btp_service_instance`), region codes, and API paths.

## Tool preferences

Before using `WebFetch` to look up SAP documentation, check if the `sap-docs` MCP server is available (tools prefixed `mcp__sap-docs__*`). If yes, use it. If not, fall back to `WebFetch`.

## BTP platform validation

### BTP operation safety boundary

The prohibition on state-changing CLI commands excludes the permitted target prelude.

When invoking the BTP CLI or any BTP MCP tool, perform **only read or list retrievals**. For the BTP CLI, invoke only documented read/list commands (for example, `btp list ...`); for BTP MCP, invoke only a tool explicitly documented as a read/list lookup. `btp target --global-account <subdomain>` is the sole permitted account-selection prelude and may be used only immediately before those read/list CLI commands. Never invoke, suggest, or approve a BTP operation that creates, updates, deletes, assigns, unassigns, enables, disables, or otherwise mutates BTP state — even when requested by the user. Do not run login, config, profile, or any other state-changing CLI command.

If this workflow needs a live BTP availability check, read `<project-root>/.sap-iac/platform-validation.md` and prefer its recorded CLI route, then this agent's recorded BTP MCP route. If no route is recorded, retain user input without blocking. If a recorded route cannot authenticate, target, or complete its lookup, ask the user to resolve it before relying on platform data.

---

## Governance Check

**Before defining any security configuration**, locate the sap-iac project root by walking up from the current working directory until a directory containing `specs/`, `memory/`, and `terraform/` is found. Then check whether `<project-root>/memory/governance.md` exists.

**If it does not exist:** proceed without constraints and note: "No governance rules found — proceeding without enforcement."

**If it exists**, load all rules and validate every security decision against the `## Security` section.

### IdP validation
- If `Custom IdP: required` and no custom IdP is being configured: **STOP**
  > "GOVERNANCE VIOLATION: Governance requires a custom Identity Provider. Configure a custom IdP trust or add `- Override: true` to memory/governance.md."

### Role collection validation
- If `Default role collections` specifies assignments and the configuration does not include them: **STOP**
  > "GOVERNANCE VIOLATION: Governance requires default role collection assignments (`<required>`). Include these assignments or add `- Override: true` to memory/governance.md."

**If `- Override: true` is set in `<project-root>/memory/governance.md`:** log a warning for each violation and continue instead of stopping.

---

## Workflow

Read `specs/scenario.md` and `specs/landscape.md` to understand authentication flows, user roles, and external system integrations.

Define for each subaccount:
- IdP trust configurations (platform and application)
- Role collections (see two-step elicitation below)
- User and group assignments to role collections

### Custom IdP detection and URL collection

**Before any other elicitation**, check `specs/scenario.md` and `specs/landscape.md` for signals indicating a custom Identity Provider (custom IdP) is required or referenced (e.g. mentions of a custom identity provider, external IdP, SAP IAS tenant, or corporate IdP trust).

**If a custom IdP signal is detected:**
- Ask: **"What is the URL of the custom Identity Provider (IdP)?"**
- This question MUST be asked. It MUST NOT be omitted or skipped under any circumstances.
- Record the raw URL in `specs/trust.md`.

**If no custom IdP signal is detected**, also check `memory/governance.md`: if it exists and contains `Custom IdP: required`, treat that as a custom IdP signal and ask for the URL. If neither source provides a signal, proceed without asking.

#### Origin derivation

From the IdP URL, derive the **origin** as follows:

> **Rule:** Take the first label of the hostname (the portion before the first `.`) and append `-platform`.

Example: `testsub12domain.accounts.ondemand.com` → origin is `testsub12domain-platform`

Record the derived origin alongside the IdP URL in `specs/trust.md`. The derived value is reused in role collection assignments (see below) and in Cloud Foundry space assignments (see below). It MUST NOT be used as the `origin` of `btp_subaccount_trust_configuration`; the identity-provider URL is sufficient for that resource unless the user explicitly supplied a trust-configuration origin.

After recording the derived origin, ask: **"The derived origin is `<derived-origin>`. Optionally, provide a different explicit origin for this IdP trust configuration; leave blank to omit the Terraform `origin` attribute."** Record a trust-configuration origin only when the user supplies it, followed by `Trust configuration origin explicit: true`. Do not create either field from the derived origin.

### Role collection elicitation

For each subaccount, ask: **"Are role collections required for this subaccount?"**

If **yes**:
- Collect the name and optional description of each role collection.
- Then ask: **"Should individual roles be assigned to this role collection? (optional — leave blank to skip)"**
  - If roles are specified, collect each role as three fields: `role_name`, `role_template_name`, `role_template_app_id`.
  - If no roles are specified, the role collection is valid with an empty roles list — do not ask again or infer assignments.

If **no**: record no role collections for that subaccount.

Do **not** use unstructured "role template assignments" text. Every role must be captured as a named, structured entry with all three fields.

### Role collection assignment elicitation

**This section applies whenever role collections are created (regardless of whether a custom IdP is in use).**

After collecting all role collections for a subaccount, ask:

> **"How should role collections be assigned — by individual user or by group?"**

Then collect the assignment entries:

- **User assignment**: for each entry collect `username` and `role_collection_name`.
- **Group assignment**: for each entry collect `group_name` and `role_collection_name`.

**When a custom IdP is in use** (i.e. an IdP URL was collected and an origin was derived), each assignment entry MUST also include the `origin` field (the derived value from [Origin derivation](#origin-derivation)).

Example — user assignment **without** custom IdP:
```
username: john.doe@example.com, role_collection_name: MyCollection
```

Example — user assignment **with** custom IdP (origin derived as `testsub12domain-platform`):
```
username: john.doe@example.com, origin: testsub12domain-platform, role_collection_name: MyCollection
```

Example — group assignment **with** custom IdP:
```
group_name: developers, origin: testsub12domain-platform, role_collection_name: MyCollection
```

### Cloud Foundry space user assignment elicitation

**This section applies only when a Cloud Foundry space creation is detected in `specs/landscape.md` or `specs/scenario.md`.**

If a CF space is in scope, ask:

> **"Which users should be assigned to roles in each Cloud Foundry space? For each entry provide: space name (as defined in the landscape), username, and one or more roles."**

The `space_name` MUST match the CF space name in `specs/landscape.md`.

**Origin rules for CF space assignments:**
- **With custom IdP**: automatically use the origin derived from the IdP URL (see [Origin derivation](#origin-derivation)); do not ask the user to provide or override it.
- **Without custom IdP** (default SAP IAS / default IdP): use `sap.ids` as the origin.

Accepted role values (fixed list — multiple roles may be selected per user per space; each becomes a separate resource):
- `space_auditor`
- `space_developer`
- `space_manager`
- `space_supporter`

Example — with custom IdP (origin `testsub12domain-platform`):
```
space_name: dev-space, username: jane.smith@example.com, origin: testsub12domain-platform, roles: space_developer, space_manager
space_name: dev-space, username: bob.jones@example.com, origin: testsub12domain-platform, roles: space_auditor
space_name: prod-space, username: jane.smith@example.com, origin: testsub12domain-platform, roles: space_manager
```

Example — without custom IdP (origin defaults to `sap.ids`):
```
space_name: dev-space, username: jane.smith@example.com, roles: space_developer, space_manager
space_name: dev-space, username: bob.jones@example.com, roles: space_auditor
```

Collect entries as free-form input and normalize each to the structured format before writing to `specs/trust.md`. When no custom IdP is in use, set `origin: sap.ids` on each normalized entry. Each `(space_name, username, role)` combination becomes one `cloudfoundry_space_role` resource.

If no CF space is in scope, do not ask this question.

Write `specs/trust.md` with the complete security configuration. This file is the direct input to `/sap-iac.tasks`.

### `specs/trust.md` format

#### IdP configuration block

When a custom IdP is in use, write an IdP block at the top of each subaccount section:

```markdown
## <subaccount-name>

### Custom IdP
- IdP URL: <raw-url>
- Origin: <derived-origin>
- Trust configuration origin: <explicit-user-supplied-origin> # omit this line unless the user supplied it
- Trust configuration origin explicit: true # omit this line unless the user supplied the origin
```

`Trust configuration origin explicit: true` is the explicit-origin marker for the IdP trust configuration. It is independent of the derived `Origin` field: do not write either trust-configuration line from derivation, and do not write an empty placeholder when the user leaves the optional value blank.

#### Role collections

Role collections MUST be written using the following structured format. The `roles` list may be empty.

```markdown
### Role collections

- **Role collection**: <collection-name>
  - Description: <optional description — omit this line when not specified>
  - Roles:
    - role_name: <role-name>, role_template_name: <template-name>, role_template_app_id: <app-id>
    - role_name: <role-name>, role_template_name: <template-name>, role_template_app_id: <app-id>

- **Role collection**: <collection-name>
  - Description: <optional description — omit this line when not specified>
  - Roles: (none)
```

Do **not** write role collections using free-form "role template assignments" prose. Each role MUST appear as a structured list entry with all three fields: `role_name`, `role_template_name`, `role_template_app_id`.

> **Migration note**: If a `specs/trust.md` file exists that uses the old unstructured "role template assignments" format, it must be regenerated by re-running `/sap-iac.security` before proceeding to `/sap-iac.tasks`.

#### Role collection assignments

Write assignments after the role collections block. Use the variant that matches the collected assignment type.

**User assignments — without custom IdP:**
```markdown
### Role collection assignments (users)

- username: john.doe@example.com, role_collection_name: MyCollection
```

**User assignments — with custom IdP:**
```markdown
### Role collection assignments (users)

- username: john.doe@example.com, origin: testsub12domain-platform, role_collection_name: MyCollection
```

**Group assignments — without custom IdP:**
```markdown
### Role collection assignments (groups)

- group_name: developers, role_collection_name: MyCollection
```

**Group assignments — with custom IdP:**
```markdown
### Role collection assignments (groups)

- group_name: developers, origin: testsub12domain-platform, role_collection_name: MyCollection
```

#### Cloud Foundry space user assignments

Write CF space assignments when a CF space is in scope. Each `(space_name, username, role)` combination is a separate entry — do not collapse multiple roles into one line. The `origin` field is always present: use the derived custom-IdP origin when a custom IdP is in use; use `sap.ids` when no custom IdP is configured.

**With custom IdP:**
```markdown
### CF space user assignments

- space_name: dev-space, username: jane.smith@example.com, origin: testsub12domain-platform, role: space_developer
- space_name: dev-space, username: jane.smith@example.com, origin: testsub12domain-platform, role: space_manager
- space_name: dev-space, username: bob.jones@example.com, origin: testsub12domain-platform, role: space_auditor
- space_name: prod-space, username: jane.smith@example.com, origin: testsub12domain-platform, role: space_manager
```

**Without custom IdP (origin defaults to `sap.ids`):**
```markdown
### CF space user assignments

- space_name: dev-space, username: jane.smith@example.com, origin: sap.ids, role: space_developer
- space_name: dev-space, username: jane.smith@example.com, origin: sap.ids, role: space_manager
- space_name: dev-space, username: bob.jones@example.com, origin: sap.ids, role: space_auditor
```

## Next step

Next: `/sap-iac.connectivity` — define destinations and certificates.
