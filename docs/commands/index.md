# Commands

After `sap-iac init` scaffolds a project, all further work happens inside your AI agent through the `sap-iac.*` commands. Each command is a slash command (for example `/sap-iac.scenario`) that reads and writes plain-text specification files under `specs/`, `memory/`, and `terraform/`.

This section documents every command on its own page — what it reads, what it writes, how it behaves, and a worked example. The examples share a single running scenario (an HR leave-request application) so you can follow one project across the whole workflow.

!!! info "Run them in order"
    Each command builds on the files written by the previous ones. `govern` is optional but recommended first; `analyse` and `connectivity` are optional. The other seven commands are required and expect their inputs to exist. If a required input is missing, the command tells you which command to run first and stops.

## At a glance

| # | Command | Reads | Writes | Required |
|---|---|---|---|---|
| ○ | [`sap-iac.govern`](govern.md) | `memory/global-account.md` | `memory/governance.md` | Optional; recommended |
| 1 | [`sap-iac.scenario`](scenario.md) | your description | `specs/scenario.md` | Required |
| ○ | [`sap-iac.analyse`](analyse.md) | `specs/scenario.md`, your code | `specs/scenario.md` | Optional |
| 2 | [`sap-iac.accounts`](accounts.md) | `specs/scenario.md` | `specs/landscape.md` | Required |
| 3 | [`sap-iac.services`](services.md) | `specs/scenario.md`, `specs/landscape.md` | `specs/services.md` | Required |
| 4 | [`sap-iac.security`](security.md) | `specs/scenario.md`, `specs/landscape.md` | `specs/trust.md` | Required |
| ○ | [`sap-iac.connectivity`](connectivity.md) | `specs/scenario.md` | `specs/connectivity.md` | Optional |
| 5 | [`sap-iac.tasks`](tasks.md) | `specs/landscape.md`, `specs/services.md`, `specs/trust.md`, `specs/connectivity.md` | `specs/tasks.md` | Required |
| 6 | [`sap-iac.design`](design.md) | `specs/tasks.md` | `specs/tasks.md` | Required |
| 7 | [`sap-iac.generate`](generate.md) | `specs/tasks.md` | `terraform/` | Required |
| — | [`sap-iac.next`](next.md) | local project files | *(nothing)* | Utility |

When `memory/governance.md` is present, `accounts`, `services`, `security`, and `generate` validate their output against it and **block on applicable violations**. Adding `- Override: true` to the governance file downgrades blocks to warnings. Without the file, those commands continue and report that governance enforcement is inactive.

Run [`sap-iac.next`](next.md) at any point to see which files exist and which command to run next.

## Next steps

- Follow the [usage walkthrough](../walkthrough.md) to run the full sequence end to end.
- Browse the [scenario examples](../scenarios.md) for descriptions to feed `sap-iac.scenario`.
- If a command reports an error or warning, see [Troubleshooting](../troubleshooting.md).
