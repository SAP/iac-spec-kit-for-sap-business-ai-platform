---
name: btp-iac-govern
description: Establishes governance guardrails — regions, naming, service plans, security, and cost controls.
license: Apache-2.0
metadata:
  author: SAP
  version: "1.0"
---

# BTP IaC — Govern

Establishes governance guardrails for the project. Records approved regions, naming conventions, permitted service plans, security requirements, and cost controls into `memory/governance.md`. All subsequent commands validate their outputs against this file and hard-block on violations.

---

## Workflow

### Step 0 — Locate project root

Walk up from the current working directory, checking each ancestor for the presence of `specs/`, `memory/`, and `terraform/` subdirectories. The first matching directory is the project root. All file reads and writes in this command are anchored to that root.

**If no project root is found:**
> "Could not locate a btp-iac project root from the current directory. Run this command from within a project created by `btp-iac init`."
Stop. Do not create or modify any files.

---

### Step 1 — Detect existing governance file

Check whether `<project-root>/memory/governance.md` exists.

**If it exists:**
- Read the file and display a summary of the current rules grouped by category
- Ask the user: "Here are the current governance rules. What would you like to change, or type 'done' to keep them as-is."
- If the user says done or makes no changes, stop here
- Otherwise collect the changes and proceed to Step 3

**If it does not exist:**
- Proceed to Step 2

---

### Step 2 — Evaluate prompt completeness

Assess whether the user's invocation message already covers all five governance categories:

| Category | Covered if the prompt mentions… |
|---|---|
| **Regions** | specific BTP region codes, data residency, geographic restrictions, compliance reason |
| **Naming** | subaccount name patterns, token definitions, environment tier names, max length |
| **Service Plans** | plan names per environment, forbidden plans and reasons, default plan per tier |
| **Security** | IdP, custom identity provider type, role collections and scope, destinations and auth type |
| **Cost Controls** | metered services list, cost centre tag key name, whether tag is mandatory |

For each **missing** category, ask the user for that category's details. **Ask one category at a time**: emit a single question, wait for the answer, then move to the next missing category. Never bundle multiple categories into one question or one tool call — batching produces oversized, schema-invalid parameters.

Keep each question to a single short prompt. The details to gather per category are listed below as sub-points for your own reference — surface them concisely (e.g. as the prompt plus its examples), not as five separate questions crammed into one string.

- **Regions** — approved region codes (e.g. eu10, eu20); any forbidden regions; data residency requirement (e.g. EU only) and its compliance reason (e.g. GDPR); preferred default region.
- **Naming** — subaccount pattern (e.g. `{org}-{env}-{app}`) and what each token means; environment tiers (e.g. dev, test, prod); flat or hierarchical directory structure; max subaccount name length.
- **Service Plans** — permitted plans per environment tier; any forbidden plans and why (e.g. free plan has no SLA); default plan per tier when several are allowed.
- **Security** — whether a custom IdP is required and its type (SAML 2.0 / OIDC); default role collections and their user groups; role-collection scope (platform / application / both); destination restrictions and required auth type (e.g. OAuth2ClientCredentials).
- **Cost Controls** — whether metered services trigger a warning and which ones (e.g. hana-cloud, ai-core); whether a cost centre tag is required and its Terraform attribute name (e.g. cost_center).

If the prompt already covers all five categories, skip to Step 3 immediately.

---

### Step 3 — Write `<project-root>/memory/governance.md`

Write (or overwrite) `<project-root>/memory/governance.md` using the following structure. Fill in each section from the information provided by the user.

```markdown
# BTP Governance

## Regions
- Allowed: <comma-separated region codes, e.g. eu10, eu20>
- Forbidden: <comma-separated region codes, or "none">
- Data residency: <classification, e.g. "EU only">
- Compliance reason: <regulatory driver, e.g. "GDPR — customer data must not leave the EU">
- Preferred region: <default region when none is specified, e.g. eu10>

## Naming
- Subaccount pattern: <pattern, e.g. {org}-{env}-{app}>
- Pattern tokens: <token definitions, e.g. "{org} = organisation short code, {env} = environment tier, {app} = application identifier">
- Environments: <comma-separated tiers in order, e.g. dev, test, prod>
- Directory structure: <flat or hierarchical>
- Max name length: <character limit, e.g. 60>

## Service Plans
### <env-tier-1>
- Permitted: <comma-separated plan names>
- Forbidden: <comma-separated plan names, or "none">
- Forbidden reason: <why these plans are not allowed, or omit if none forbidden>
- Default plan: <plan to use when multiple are permitted and none is specified>
### <env-tier-2>
- Permitted: <comma-separated plan names>
- Forbidden: <comma-separated plan names, or "none">
- Forbidden reason: <why these plans are not allowed, or omit if none forbidden>
- Default plan: <plan to use when multiple are permitted and none is specified>

## Security
- Custom IdP: <required / optional / not required>
- IdP type: <SAML 2.0 / OIDC / N/A>
- Default role collections: <name → user group, e.g. "BTP_OPERATOR → platform admins, BTP_VIEWER → developers">
- Role collection scope: <platform / application / both>
- Destinations: <restriction description, e.g. "internal systems only" or "unrestricted">
- Destination auth type: <required auth method, e.g. OAuth2ClientCredentials, BasicAuthentication, any>

## Cost Controls
- Metered service warning: <enabled / disabled>
- Metered services: <comma-separated BTP service names, or "all">
- Cost centre tag: <required / optional / not required>
- Tag key name: <Terraform attribute name, e.g. cost_center>
```

Do **not** include `- Override: true` unless the user explicitly requests it. When present, this flag causes all downstream commands to warn instead of block on violations.

---

### Step 4 — Confirm

After writing the file, output:

```
Governance rules saved to memory/governance.md.

The following rules are now in effect:
  Regions:        <summary>
  Naming:         <summary>
  Service Plans:  <summary>
  Security:       <summary>
  Cost Controls:  <summary>

All subsequent commands (/btp-iac.accounts, /btp-iac.services,
/btp-iac.security, /btp-iac.generate) will validate their output
against these rules and stop on any violation.

To bypass enforcement, add "- Override: true" to memory/governance.md.
```
