# Usage

The CLI scaffolds projects with `sap-iac init`, reports its build with `sap-iac version`, and manages service-parameter catalogue entries with `sap-iac catalogue convert`. Infrastructure planning and generation happen inside your AI agent through the generated `sap-iac.*` commands.

## `sap-iac init`

```
sap-iac init [name] [--agent claude,codex,cursor,copilot] [--backend <type>]
```

Bootstraps a new Infrastructure-as-Code Specification Toolkit project in a directory named `[name]`.

### Argument

| Argument | Description |
|---|---|
| `[name]` | Optional. The project directory to use. In an interactive terminal, omitting it opens a menu to create a new project, adopt the current directory, or add/update agents in an existing project. |

### Flags

| Flag | Description |
|---|---|
| `--agent` | Comma-separated list of AI agents to configure: `claude`, `codex`, `cursor`, `copilot`. When omitted **in a terminal**, `init` shows an interactive multi-select. When omitted **without a TTY**, `init` reports an error asking you to pass `--agent`. An unknown agent id is rejected. |
| `--backend` | Terraform backend type for remote state. Accepted values: `local`, `remote`, `azurerm`, `consul`, `cos`, `gcs`, `http`, `kubernetes`, `oci`, `oss`, `pg`, `s3`. When omitted during a fresh or adopt init, `init` shows an interactive select with a skip option. Skipped or omitted for agent-only updates. If `memory/backend.md` already exists it is never overwritten. |

### Examples

```sh
# Interactive: prompts for agents, backend, and name (if omitted)
sap-iac init my-iac-project

# Configure for Claude Code with an S3 backend
sap-iac init my-iac-project --agent claude --backend s3

# Configure for multiple agents, skip backend
sap-iac init my-iac-project --agent claude,cursor

# Refresh an existing project without selecting agents interactively
sap-iac init my-iac-project --agent claude,codex
```

Fresh and adopt initialization prompt for the optional global-account subdomain, the backend type, and adoption asks for confirmation. They therefore require an interactive input stream even when `--agent` is supplied. A fully non-interactive invocation requires both `--agent` and `--backend` (or pass `--backend` with a known type to skip the prompt).

## What `init` creates

```
my-iac-project/
├── .gitignore            # ignores .terraform/, *.tfstate, *.tfstate.backup,
│                         # .terraform.lock.hcl, .sap-iac/platform-validation.md,
│                         # and memory/global-account.md
├── specs/                # (empty) requirement specs the agent commands write
├── memory/               # project memory used by the generated commands
│   ├── global-account.md # optional global-account subdomain entered at init
│   ├── backend.md        # backend type + placeholder HCL (only if --backend chosen)
│   └── service-params-catalogue.yaml # editable service-instance parameters
├── terraform/            # generated Terraform HCL lands here
├── .sap-iac/             # internal state (platform-validation.md)
└── <agent command dir>/  # one per selected agent (see below)
```

During a fresh init (and when adopting an existing directory), `init` prompts for an optional **global-account subdomain** (recorded in `memory/global-account.md`) and an optional **Terraform backend type** (recorded in `memory/backend.md` as a type label and full placeholder HCL block with `# TODO` comments). When `git` is available, `init` runs `git init` in the new directory. `memory/backend.md` is never overwritten on subsequent runs — edit it freely. The `sap-iac.design` and `sap-iac.generate` commands read this file and copy the HCL block into each configuration unit's `backend.tf` at generation time.

!!! warning "Existing projects are updated, not overwritten"
    Running `init` against an existing directory does not error. If the directory is already a sap-iac project (it has `specs/` or `memory/`), `init` updates the selected agents' command files, refreshes `.sap-iac/platform-validation.md`, and ensures its managed `.gitignore` entries. Existing `specs/`, `memory/`, and `terraform/` content remains untouched. If it is some other existing directory, `init` offers to adopt it and scaffold the project structure in place.

The service-parameter catalogue is created during fresh and adopt initialization. Adoption preserves an existing catalogue, and agent-only refresh does not create or modify it. See [Service-parameter catalogue](catalogue.md).

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

| Step | Command | Purpose |
|---|---|---|
| ○ | `sap-iac.govern` | *Optional; recommended first.* Set guardrails — regions, naming, cost policies. Without it, downstream commands proceed without governance enforcement. |
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

## Other CLI commands

- `sap-iac version` prints the binary version.
- `sap-iac catalogue convert` converts a JSON Schema into a service-parameter catalogue entry. See [Service-parameter catalogue](catalogue.md).
