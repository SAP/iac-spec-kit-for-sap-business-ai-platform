# Usage walkthrough

This walkthrough is a complete end-to-end run: install the CLI, scaffold a project, drive the AI agent through the command workflow, and finish with generated Terraform. It uses the single-application scenario from the [scenario examples](scenarios.md); substitute any other scenario description to build something larger.

!!! note "Prerequisites"
    The `btp-iac` CLI (see [Setup](setup.md)) and an AI agent — Claude Code, Codex, Cursor, or GitHub Copilot. Terraform and Git are recommended but optional.

## 1 · Scaffold the project

```sh
btp-iac init hr-leave --agent claude
cd hr-leave
```

This creates the `specs/`, `memory/`, `terraform/`, and `.btp-iac/` (internal state) directories, a `.gitignore`, and the `btp-iac.*` command files for your agent (here, under `.claude/commands/`). It also prompts for an optional global-account subdomain, saved to `memory/global-account.md`. Missing Terraform, Git, or MCP servers produce warnings only — the scaffold still succeeds. If anything looks off, see [Troubleshooting](troubleshooting.md).

## 2 · Open the project in your agent

```sh
claude
```

Run the commands below as slash commands inside the agent (`/btp-iac.govern`, `/btp-iac.scenario`, and so on), in order.

## 3 · Run the command workflow

| # | Command | What it does | Writes |
|---|---|---|---|
| ○ | `btp-iac.govern` | Set guardrails — regions, naming, service plans, cost policy *(optional)* | `memory/governance.md` |
| 1 | `btp-iac.scenario` | Describe the application (see below) | `specs/scenario.md` |
| 2 | `btp-iac.analyse` | Scan existing source or Terraform to enrich the scenario *(optional)* | `specs/scenario.md` |
| 3 | `btp-iac.accounts` | Map the application to directories and subaccounts | `specs/landscape.md` |
| 4 | `btp-iac.services` | Resolve the BTP services each subaccount needs | `specs/services.md` |
| 5 | `btp-iac.security` | Set up IdP trust, role collections, and assignments | `specs/trust.md` |
| 6 | `btp-iac.connectivity` | Define destinations and certificates *(optional)* | `specs/connectivity.md` |
| 7 | `btp-iac.tasks` | Build a dependency-ordered execution plan | `specs/tasks.md` |
| 8 | `btp-iac.design` | Annotate each task with its target Terraform file | `specs/tasks.md` |
| 9 | `btp-iac.generate` | Write and validate all Terraform HCL | `terraform/` |

For step 1, provide a scenario description. Using the single-application example:

!!! example "Single-application scenario"
    "An internal HR leave-request app built with CAP (Node.js) on Cloud Foundry. It stores leave requests in SAP HANA Cloud and authenticates employees via XSUAA. Single environment, one subaccount, region eu10."

When `memory/governance.md` is present, `accounts`, `services`, `security`, and `generate` validate against it and **block on violations**, so the guardrails set with `btp-iac.govern` are enforced throughout the run.

!!! tip "Lost track of where you are?"
    Run `/btp-iac.next` any time — it inspects the project files and tells you which command to run next.

## 4 · Review and apply

The toolkit generates and validates the configuration; applying it against your BTP account remains under your control.

```sh
cd terraform
terraform init
terraform plan
```

## Try a larger scenario

Repeat from step 1 with a different name and a more complex description from the [scenario examples](scenarios.md) — multiple environments, external integrations, a multi-application directory, or a full enterprise landing zone. The commands stay the same; only what each one produces grows.
