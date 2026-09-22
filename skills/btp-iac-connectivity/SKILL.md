---
name: btp-iac-connectivity
description: Elicits BTP destination and destination-certificate requirements and writes specs/connectivity.md.
license: Apache-2.0
metadata:
  author: SAP
  version: "1.0"
---

# BTP IaC — Connectivity

Elicits and structures BTP destination and destination-certificate requirements for a project.

Optionally reads `specs/scenario.md` when it exists (reusing any destinations already captured there). Produces `specs/connectivity.md`, which is an optional input to `/btp-iac.tasks`.

## Tool preferences

Before using `WebFetch` to look up SAP or Terraform documentation, check if the `sap-docs` MCP server (tools prefixed `mcp__sap-docs__*`) or the `terraform` MCP server is available. If yes, prefer those. If not, fall back to `WebFetch`.

## BTP platform validation

### BTP operation safety boundary

The prohibition on state-changing CLI commands excludes the permitted target prelude.

When invoking the BTP CLI or any BTP MCP tool, perform **only read or list retrievals**. For the BTP CLI, invoke only documented read/list commands (for example, `btp list ...`); for BTP MCP, invoke only a tool explicitly documented as a read/list lookup. `btp target --global-account <subdomain>` is the sole permitted account-selection prelude and may be used only immediately before those read/list CLI commands. Never invoke, suggest, or approve a BTP operation that creates, updates, deletes, assigns, unassigns, enables, disables, or otherwise mutates BTP state — even when requested by the user. Do not run login, config, profile, or any other state-changing CLI command.

If this workflow needs a live BTP availability check, read `<project-root>/.btp-iac/platform-validation.md` and prefer its recorded CLI route, then this agent's recorded BTP MCP route. If no route is recorded, retain user input without blocking. If a recorded route cannot authenticate, target, or complete its lookup, ask the user to resolve it before relying on platform data.

---

## Workflow

### Step 1 — Check for existing destinations in scenario.md

Locate the btp-iac project root by walking up from the current working directory until a directory containing `specs/`, `memory/`, and `terraform/` is found.

If `specs/scenario.md` exists, read it and look for any destinations already described (the scenario command captures these in its follow-up question). If destinations are found:

- Present them to the user with their captured details.
- Ask the user to confirm, extend, or modify each one.
- Collect any missing required fields (see Step 2) rather than re-asking for already-known information.

If `specs/scenario.md` does not exist, or exists but contains no destination information, proceed directly to Step 2.

### Step 2 — Elicit destination requirements

Ask the user how many destinations are needed (skip if already answered). For each destination, present the following **canonical patterns** as hints and ask which best describes their use case:

---

**Pattern A — HTTP / NoAuthentication**
Internet-accessible service with no credentials.
```
Type:          HTTP
ProxyType:     Internet
Authentication: NoAuthentication
URL:           https://myservice.example.com
```
*Required fields*: Name, URL.

---

**Pattern B — HTTP / BasicAuthentication via Cloud Connector**
On-premise system reached through SAP Cloud Connector.
```
Type:              HTTP
ProxyType:         OnPremise
Authentication:    BasicAuthentication
URL:               https://internal.corp/service
User:              <username>
Password:          <PLACEHOLDER — do not store in repo>
CloudConnectorLocationId: <location-id>
```
*Required fields*: Name, URL, User, CloudConnectorLocationId. Password is a placeholder.

---

**Pattern C — HTTP / OAuth2ClientCredentials**
Token-based machine-to-machine call.
```
Type:            HTTP
ProxyType:       Internet
Authentication:  OAuth2ClientCredentials
URL:             https://myservice.example.com
clientId:        <client-id>
tokenServiceURL: https://myauth.example.com/oauth/token
clientSecret:    <PLACEHOLDER — do not store in repo>
```
*Required fields*: Name, URL, clientId, tokenServiceURL. clientSecret is a placeholder.

---

**Pattern D — RFC**
SAP ABAP system via Java Connector (JCo).
```
Type:                                   RFC
jco.client.ashost:                      <hostname>
jco.client.sysnr:                       <system-number>
jco.client.client:                      <client>
jco.client.user:                        <user>
jco.client.passwd:                      <PLACEHOLDER — do not store in repo>
jco.destination.auth_type:              CONFIGURED_USER
jco.destination.proxy_type:             OnPremise
```
*Required fields*: Name, ashost, sysnr, client, user. passwd is a placeholder.

---

**Pattern E — LDAP**
Directory service lookup.
```
Type:                 LDAP
ldap.url:             ldap://ldap.example.com:389
ldap.proxyType:       Internet
ldap.authentication:  BasicAuthentication
ldap.user:            <user>
ldap.password:        <PLACEHOLDER — do not store in repo>
```
*Required fields*: Name, ldap.url, ldap.user. ldap.password is a placeholder.

---

**Pattern F — MAIL**
SMTP relay for outbound email.
```
Type:              MAIL
Authentication:    BasicAuthentication
ProxyType:         OnPremise
mail.user:         user@example.com
mail.password:     <PLACEHOLDER — do not store in repo>
```
*Required fields*: Name, mail.user. mail.password is a placeholder.

---

**Pattern G — TCP**
Raw TCP connection to an on-premise host.
```
Type:        TCP
Address:     host:1234
ProxyType:   OnPremise
```
*Required fields*: Name, Address (host:port).

---

For each destination, also ask:
- **Scope**: is it at the subaccount level, or tied to a specific service instance? If service-instance-scoped, collect the service instance name/ID.
- **Description**: optional free-text description.

> **Credential rule**: Never write actual passwords, client secrets, certificates, or tokens to `specs/connectivity.md`. Record `<PLACEHOLDER>` with a note: `# ⚠ Supply this value at apply time — do not commit to the repository`.

### Step 3 — Elicit certificate requirements (optional)

Ask whether any destination certificates (`btp_subaccount_destination_certificate`) are needed.

If yes, for each certificate present the following patterns:

---

**Certificate Pattern 1 — Subaccount-level certificate**
```
subaccount_id:       <subaccount>
certificate_name:    my-cert.pem    # must include valid extension: .pem, .p12, .jks, .pfx
certificate_content: filebase64("<path-to-cert-file>")
```
*Required fields*: certificate_name (with extension), file path or indication the content is supplied at apply time.

---

**Certificate Pattern 2 — Service-instance-level certificate**
```
subaccount_id:       <subaccount>
service_instance_id: <service-instance-id>
certificate_name:    my-cert.pem
certificate_content: filebase64("<path-to-cert-file>")
```
*Required fields*: Same as above, plus service_instance_id.

---

For each certificate collect:
- Certificate name including its extension (`.pem`, `.p12`, `.jks`, or `.pfx`).
- Subaccount scope and, if applicable, service instance ID.
- File path of the certificate — or, if the file is not yet available, a note that content will be supplied at apply time.

> **Credential rule**: Do not write raw certificate bytes or base64-encoded content to `specs/connectivity.md`. Record the file path or `<PLACEHOLDER — supply at apply time>`.

If the user confirms no certificates are needed, omit the certificates section.

### Step 4 — Write specs/connectivity.md

Write the collected requirements to `<project-root>/specs/connectivity.md` using the following structure:

```markdown
# Connectivity

## Destinations

### <destination-name>

- **Type**: <HTTP|RFC|LDAP|MAIL|TCP>
- **Subaccount**: <subaccount name or reference>
- **Service Instance** (if applicable): <service instance name or ID>
- **ProxyType**: <Internet|OnPremise>
- **URL / Address**: <value>
- **Authentication**: <method>
- **CloudConnectorLocationId** (if OnPremise HTTP): <value>
- **Credentials**: ⚠ PLACEHOLDER — supply at apply time, do not commit to the repository
- **Description**: <optional>

<!-- repeat for each destination -->

## Certificates

### <certificate-name>

- **Subaccount**: <subaccount name or reference>
- **Service Instance** (if applicable): <service instance name or ID>
- **File**: <path or PLACEHOLDER — supply at apply time>

<!-- repeat for each certificate; omit this section if no certificates -->
```

This file is the direct input to `/btp-iac.tasks` (read when it exists) and informs `/btp-iac.generate` which produces `btp_subaccount_destination_generic` and `btp_subaccount_destination_certificate` Terraform resources.
