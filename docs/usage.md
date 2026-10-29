# Usage

The CLI scaffolds projects with `sap-iac init`, reports its build with `sap-iac version`, and manages service-parameter catalogue entries with `sap-iac catalogue convert`. Infrastructure planning and generation happen inside your AI agent through the generated `sap-iac.*` commands.

## `sap-iac init`

```
sap-iac init [name] [--agent claude,codex,cursor,copilot,gemini] [--backend <type>]
```

Bootstraps a new Infrastructure-as-Code Specification Toolkit project in a directory named `[name]`.

When the target already exists, `init` updates its configured agent files if it is
already a sap-iac project, or offers to adopt it otherwise.

### Argument

| Argument | Description |
|---|---|
| `[name]` | Optional. The project directory to use. In an interactive terminal, omitting it opens a menu to create a new project, adopt the current directory, or add/update agents in an existing project. |

### Flags

| Flag | Description |
|---|---|
| `--agent` | Comma-separated list of AI agents to configure: `claude`, `codex`, `cursor`, `copilot`, `gemini`. When omitted **in a terminal**, `init` shows an interactive multi-select. When omitted **without a TTY**, `init` reports an error asking you to pass `--agent`. An unknown agent id is rejected. |
| `--backend` | Terraform backend type for remote state. Accepted values: `local`, `remote`, `azurerm`, `consul`, `cos`, `gcs`, `http`, `kubernetes`, `oci`, `oss`, `pg`, `s3`. When omitted during a fresh or adopt init, `init` shows an interactive select with a skip option. Skipped or omitted for agent-only updates. If `memory/backend.md` already exists it is never overwritten. |

### Examples

```sh
# Interactive: prompts for agents, backend, and name (if omitted)
sap-iac init my-iac-project

# Configure for Claude Code with an S3 backend
sap-iac init my-iac-project --agent claude --backend s3

# Configure Gemini CLI commands and native skills
sap-iac init my-iac-project --agent gemini

# Configure Gemini CLI alongside Claude Code
sap-iac init my-iac-project --agent claude,gemini

# Configure for multiple agents, skip backend
# The backend chooser is still shown; select "skip (no backend.tf)".
sap-iac init my-iac-project --agent claude,cursor

# Refresh an existing project without selecting agents interactively
sap-iac init my-iac-project --agent claude,codex
```

Fresh and adopt initialization prompt for the optional global-account subdomain; adoption also asks for confirmation. Omitting `--backend` adds the backend chooser, and a fresh project may prompt for optional remote Git setup. `--backend` skips only the backend chooser. Fresh and adopt modes therefore require an interactive input stream even when `--agent` and `--backend` are supplied. Only an agent-only refresh of an existing sap-iac project is fully non-interactive.

### Optional remote repository setup

For a fresh project, `init` can optionally connect a remote Git repository. In the interactive **Create a new project** flow, choosing a remote URL derives the local project directory from the repository name (for example, `https://github.com/acme/orders.git` creates `orders/`). A trailing slash and `.git` suffix are ignored for this purpose.

The CLI checks whether the remote has branches before creating files:

- For an existing remote, it clones the repository into the derived directory, creates and checks out the chosen working branch, then adds sap-iac files.
- For an empty remote, it scaffolds locally, adds `origin`, renames the initial branch to the chosen default branch, and creates the chosen working branch. Commit and push the generated files when you are ready.

The default working branch is `feature/iac-setup`; an empty remote also prompts for a default branch, `main` by default. Both inputs must be valid Git branch names. Remote setup is currently interactive; `init` has no remote URL or branch flags. Git and access to the remote are required when this option is selected.

## What `init` creates

```
my-iac-project/
├── .gitignore            # ignores .terraform/, *.tfstate, *.tfstate.backup,
│                         # .terraform.lock.hcl, .sap-iac/platform-validation.md,
│                         # and memory/global-account.md
├── specs/                # (empty) requirement specs the agent commands write
├── memory/               # project memory used by the generated commands
│   ├── global-account.md # created at init; its global-account subdomain is optional
│   ├── backend.md        # backend type + placeholder HCL (only if --backend chosen)
│   └── service-params-catalogue.yaml # editable service-instance parameters
├── terraform/            # generated Terraform HCL lands here
├── .sap-iac/             # internal state (platform-validation.md)
└── <agent integration>/  # selected agent commands; Gemini also receives skills
```

During a fresh init (and when adopting an existing directory), `init` prompts for an optional **global-account subdomain**. It always creates `memory/global-account.md` for these modes, with a blank subdomain when none was supplied. It also prompts for an optional **Terraform backend type** (recorded in `memory/backend.md` as a type label and full placeholder HCL block with `# TODO` comments). For a local fresh project, `init` runs `git init` when Git is available. When remote setup is selected, it instead clones an existing remote or configures an empty one as described above. `memory/backend.md` is never overwritten on subsequent runs — edit it freely. The `sap-iac.design` and `sap-iac.generate` commands read this file and copy the HCL block into each configuration unit's `backend.tf` at generation time.

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
| `gemini` | `.gemini/commands/` | `.toml` |

For example, `--agent claude` produces `.claude/commands/sap-iac.scenario.md`, `.claude/commands/sap-iac.generate.md`, and so on.

### Gemini CLI integration

Selecting `gemini` installs both integration forms for all eleven workflows:

| Integration | Example file | How to use it |
|---|---|---|
| Slash command | `.gemini/commands/sap-iac.scenario.toml` | Run `/sap-iac.scenario` inside Gemini CLI. |
| Native skill | `.gemini/skills/sap-iac-scenario/SKILL.md` | Gemini discovers its name and description and can activate it when your request matches. |

The command TOML contains a description and the workflow prompt. Native skills retain their metadata and bundled runtime resources, including the services catalogue. Both forms use the same workflow instructions, adapted for Gemini tools. The editable service-parameter catalogue remains in `memory/service-params-catalogue.yaml`.

Start `gemini` from the project directory. Use `/commands list` to inspect slash commands and `/skills list` to inspect native skills. After refreshing files, run `/commands reload` and `/skills reload`, or start a new session. Native skills complement the explicit `/sap-iac.*` workflow; they do not replace its slash commands.

Files are installed only within the selected project. Gemini also supports `.agents/skills/` as an alias, which takes precedence over `.gemini/skills/` for a same-named workspace skill. See the official [custom command](https://geminicli.com/docs/cli/custom-commands/) and [skill discovery](https://geminicli.com/docs/cli/skills/) documentation.

## MCP and platform-validation checks

During `init`, the CLI checks whether your agent has an MCP server configured for **Terraform/OpenTofu**, and whether a **BTP CLI or BTP MCP** route is available for platform validation. Each missing capability prints an advisory warning, for example:

For Gemini, configuration discovery reads `mcpServers` in user `~/.gemini/settings.json` and the initialized project's `.gemini/settings.json`, including when you launch `init` from its parent directory. Server names containing `terraform`, `opentofu`, or `btp` are matched case-insensitively. Detection inspects configuration; it does not connect to servers. Missing, unreadable, or malformed files do not block initialization. Configure servers using the official [Gemini MCP guide](https://geminicli.com/docs/tools/mcp-server/).

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

| Command | Purpose | Output / effect |
|---|---|---|
| `sap-iac version` | Print the binary version. | Writes the version to standard output. |
| `sap-iac catalogue convert <schema.json> --service <offering> --plans <plan[,plan]>` | Convert a JSON Schema into a service-parameter catalogue entry. | Writes YAML to standard output by default. |
| `sap-iac catalogue convert ... --append` | Add the converted entry to a project catalogue. | Walks upward from the current directory to the nearest directory containing `specs/` or `memory/`, then appends to `memory/service-params-catalogue.yaml`; it fails when no project is found. |

See [Service-parameter catalogue](catalogue.md) for conversion rules and examples.
