# Usage

The CLI exposes a single command, `sap-iac init`, which scaffolds a new project. All subsequent work happens inside your AI agent through the generated `sap-iac.*` commands.

## `sap-iac init`

```
sap-iac init [name] [--agent claude,codex,cursor,copilot]
```

Bootstraps a new Infrastructure-as-Code Specification Toolkit project in a directory named `[name]`.

### Argument

| Argument | Description |
|---|---|
| `[name]` | Optional. The project directory to use. In an interactive terminal, omitting it opens a menu to create a new project, adopt the current directory, or add/update agents in an existing project. In a non-interactive context (no TTY, such as CI) the command cannot prompt, so the name must be passed explicitly. |

### Flag

| Flag | Description |
|---|---|
| `--agent` | Comma-separated list of AI agents to configure: `claude`, `codex`, `cursor`, `copilot`. When omitted **in a terminal**, `init` shows an interactive multi-select. When omitted **without a TTY**, `init` reports an error asking you to pass `--agent`. An unknown agent id is rejected. |

### Examples

```sh
# Interactive: prompts for agents (and for the name if omitted)
sap-iac init my-iac-project

# Configure for Claude Code
sap-iac init my-iac-project --agent claude

# Configure for multiple agents
sap-iac init my-iac-project --agent claude,cursor

# Non-interactive (CI): name and --agent are both required
sap-iac init my-iac-project --agent claude,codex
```

## What `init` creates

```
my-iac-project/
├── .gitignore            # ignores .terraform/, *.tfstate, *.tfstate.backup,
│                         # .terraform.lock.hcl, .sap-iac/platform-validation.md,
│                         # and memory/global-account.md
├── specs/                # (empty) requirement specs the agent commands write
├── memory/               # global-account.md is written at init; governance.md
│                         # is authored later by sap-iac.govern
│   └── global-account.md # the global-account subdomain entered at init (optional)
├── terraform/            # (empty) generated Terraform HCL lands here
├── .sap-iac/             # internal state (platform-validation.md)
└── <agent command dir>/  # one per selected agent (see below)
```

During a fresh init (and when adopting an existing directory), `init` also prompts for an optional **global-account subdomain** and records it in `memory/global-account.md`. When `git` is available, `init` runs `git init` in the new directory.

!!! warning "Existing projects are updated, not overwritten"
    Running `init` against an existing directory does not error. If the directory is already a sap-iac project (it has `specs/` or `memory/`), `init` only adds or updates the selected agents' command files, leaving your specs and memory untouched. If it is some other existing directory, `init` adopts it — scaffolding the project structure in place.

### Agent command directories

Each selected agent receives a directory of `sap-iac.<skill>` command files — one per skill: `govern`, `scenario`, `analyse`, `accounts`, `services`, `security`, `connectivity`, `tasks`, `design`, `generate`, and the `next` utility.

| Agent | Directory | File extension |
|---|---|---|
| `claude` | `.claude/commands/` | `.md` |
| `codex` | `.codex/prompts/` | `.md` |
| `cursor` | `.cursor/rules/` | `.mdc` |
| `copilot` | `.github/instructions/` | `.instructions.md` |

For example, `--agent claude` produces `.claude/commands/sap-iac.scenario.md`, `.claude/commands/sap-iac.generate.md`, and so on.

## MCP and platform-validation checks

During `init`, the CLI checks whether your agent has an MCP server configured for **Terraform/OpenTofu**, and whether a **BTP CLI or BTP MCP** route is available for platform validation. Each missing capability prints an advisory warning, for example:

```
[claude] No Terraform or OpenTofu MCP server detected.
Terraform operations in your AI agent may not work.
```

```
[claude] No BTP CLI or BTP MCP server detected. Platform validation will use user input without blocking.
```

!!! info "Advisory only"
    These checks never stop `init`. Configuring the servers is what lets the agent actually run Terraform and validate against your account; setting them up is your responsibility.

## The agent workflow

After `init`, open the project in your AI agent and run the commands in order:

```mermaid
flowchart TD
    init([sap-iac init]) --> govern

    govern["sap-iac.govern<br/><i>optional</i>"] --> scenario[sap-iac.scenario]
    scenario --> analyse["sap-iac.analyse<br/><i>optional</i>"]
    analyse --> accounts[sap-iac.accounts]
    accounts --> services[sap-iac.services]
    services --> security[sap-iac.security]
    security --> connectivity["sap-iac.connectivity<br/><i>optional</i>"]
    connectivity --> tasks[sap-iac.tasks]
    tasks --> design[sap-iac.design]
    design --> generate[sap-iac.generate]
    generate --> tf([terraform/ HCL])

    next["sap-iac.next<br/><i>utility — run any time</i>"] -.-> scenario

    classDef optional fill:#f5f5f5,stroke:#999,stroke-dasharray:4 3;
    classDef util fill:#eef,stroke:#88a,stroke-dasharray:4 3;
    class govern,analyse,connectivity optional;
    class next util;
```

| Step | Command | Purpose |
|---|---|---|
| ○ | `sap-iac.govern` | *Optional.* Set guardrails — regions, naming, cost policies. Skip to use defaults. |
| 1 | `sap-iac.scenario` | Describe your application. |
| 2 | `sap-iac.analyse` | *Optional.* Scan source or Terraform code to extract service dependencies. |
| 3 | `sap-iac.accounts` | Map your application to BTP directories and subaccounts. |
| 4 | `sap-iac.services` | Resolve which BTP services each subaccount needs. |
| 5 | `sap-iac.security` | Set up IdP trust, role collections, and role-collection assignments. |
| 6 | `sap-iac.connectivity` | *Optional.* Define destinations and certificates for external/on-premise systems. |
| 7 | `sap-iac.tasks` | Build a dependency-ordered execution plan. |
| 8 | `sap-iac.design` | Plan the Terraform file and module layout. |
| 9 | `sap-iac.generate` | Write and validate all Terraform HCL. |
| — | `sap-iac.next` | *Utility.* Show the current project state and recommend the next command. Run any time. |

Each command reads and writes files under `specs/`, `memory/`, and `terraform/`. For a detailed description of every command — its inputs, outputs, and governance behaviour — see the [command reference](commands/index.md). For a concrete run-through, see the [usage walkthrough](walkthrough.md).
