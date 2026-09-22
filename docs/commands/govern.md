# `btp-iac.govern`

!!! abstract "Summary"
    **Role:** Optional — the foundation. · **Reads:** `memory/governance.md`, `memory/global-account.md` (if present) · **Writes:** `memory/governance.md`

Establishes the governance guardrails that every later command validates against: preferred infrastructure provider, allowed regions, permitted runtime environments, naming conventions, service plans, security requirements, and cost controls. Run it first if your organisation has rules that the generated landscape must respect.

## When to run it

Before `btp-iac.scenario`, at the very start of a project — or any time you want to review or change the guardrails already in place. It is entirely optional: skip it and the workflow proceeds without enforcement.

## Inputs and outputs

| | |
|---|---|
| **Reads** | `memory/governance.md`, if it already exists, so it can summarise and edit it; and `memory/global-account.md` for the global-account subdomain set at `init`. |
| **Writes** | `memory/governance.md`, with six sections: Regions (including preferred infrastructure provider), Account Setup, Naming, Service Plans, Security, and Cost Controls. |
| **Asks** | One question at a time, in a fixed order, for each of the seven categories your prompt has not already covered: Preferred Infrastructure → Region → Environments → Naming → Service Plans → Security → Cost Controls. On an existing file, it asks what to change. |

## Behaviour

- Must run inside a project created by `btp-iac init`: it walks up from the current directory looking for a root that contains `specs/`, `memory/`, and `terraform/`, and stops without creating or modifying anything if none is found.
- Inspects your invocation prompt to see which categories you have already described, and asks one question at a time only for the ones still uncovered.
- When your prompt names specific regions or service plans and a global-account subdomain is configured in `memory/global-account.md`, it validates them against the live BTP account using read/list lookups only, warns on an infrastructure-provider mismatch, and may ask for a valid replacement if a value is unavailable. With no subdomain configured, it keeps your input without a live check.
- Writes the guardrails to a fixed section structure that downstream commands know how to read. (The *Environments* answer is split across `## Account Setup` and `## Naming`.)
- When present, `accounts`, `services`, `security`, and `generate` enforce these rules and stop on violations — except an infrastructure-provider mismatch, which is advisory only.

!!! tip "Set guardrails once, enforce everywhere"
    Add `- Override: true` to `memory/governance.md` to turn hard blocks into warnings — useful when you deliberately need an exception.

## Example

Invoke the command with the rules you already know, and answer its follow-up questions for anything you left out:

```
/btp-iac.govern Only EU regions (eu10, eu20). Subaccounts named <project>-<env>.
Production databases must use the "large" plan. Every subaccount needs a cost-centre tag.
```

The command confirms the rules it captured and writes them to `memory/governance.md`, for example:

```markdown
## Regions
- Allowed: eu10, eu20

## Naming
- Subaccount pattern: <project>-<env>
- Environments: dev, test, prod
```

From here on, if `accounts` is asked to place a subaccount in `us10`, it will stop and report the violation.

## Related

- Next in the workflow: [`btp-iac.scenario`](scenario.md).
- Guardrail-aware commands: [`btp-iac.accounts`](accounts.md), [`btp-iac.services`](services.md), [`btp-iac.security`](security.md), [`btp-iac.generate`](generate.md).
