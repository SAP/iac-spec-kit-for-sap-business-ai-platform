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

For each **missing** category, ask exactly one targeted question:

| Missing category | Question to ask |
|---|---|
| Regions | "Which BTP regions are approved for this project (e.g. eu10, eu20)? Are any explicitly forbidden? What data residency requirement applies (e.g. EU only), and what is the compliance reason (e.g. GDPR — customer data must not leave the EU)? Which region should be the preferred default?" |
| Naming | "What naming convention should subaccounts follow (e.g. `{org}-{env}-{app}`)? What does each token mean (e.g. `{org}` = organisation short code)? What environment tiers exist (e.g. dev, test, prod)? Is the directory structure flat or hierarchical? Is there a maximum subaccount name length?" |
| Service Plans | "For each environment tier, which BTP service plans are permitted? Are any plans explicitly forbidden, and if so why (e.g. free plan has no SLA — forbidden in prod)? What is the default plan for each tier when multiple are allowed?" |
| Security | "Is a custom Identity Provider required, and if so what type — SAML 2.0 or OIDC? Which role collections should be assigned by default and to which user groups? Should role collections apply at platform level, application level, or both? Are destinations restricted (e.g. internal systems only), and what authentication type is required (e.g. OAuth2ClientCredentials)?" |
| Cost Controls | "Should metered services trigger a warning before use? If so, which specific services should trigger it (e.g. hana-cloud, ai-core, build-workzone)? Is a cost centre tag required on all subaccounts? If so, what is the tag attribute name in Terraform (e.g. cost_center)?" |

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
