# Troubleshooting

This page lists the errors and warnings `btp-iac init` can emit, and how to resolve each one. Errors stop the command; warnings are advisory and the project is still created.

## Errors (`init` stops)

### `no TTY detected: use --agent to specify agents (e.g. --agent claude,codex,cursor,copilot)`

You ran `init` without `--agent` in a non-interactive context (CI, a pipe, or some IDE terminals). Without a terminal, the interactive picker cannot be shown, so pass the agents explicitly:

```sh
btp-iac init my-project --agent claude
```

In a non-interactive context you must also pass the project **name** as an argument, because `init` cannot prompt for it either.

### `unknown agent "<id>" — supported: claude, codex, cursor, copilot`

`--agent` accepts only `claude`, `codex`, `cursor`, and `copilot`, comma-separated. Check for typos and stray spaces — for example, `--agent claude,cursor`.

### `--agent value is empty — supported: claude, codex, cursor, copilot`

You passed `--agent` with only whitespace (e.g. `--agent " "`). Provide at least one agent ID, comma-separated: `--agent claude`.

### `at least one agent must be selected`

The interactive multi-select was dismissed without a choice. Select at least one agent (space to toggle, enter to confirm), or pass `--agent`.

### `invalid project name "<name>": must not contain path separators or start with '.'`

The project name is validated before any scaffolding runs. Names containing a path separator (e.g. `btp-iac init foo/bar`) or starting with a dot (e.g. `btp-iac init .foo`) are rejected to prevent directory traversal and hidden-directory confusion. Choose a plain name: `btp-iac init my-project`.

## Warnings (`init` still succeeds)

!!! info "Warnings do not stop `init`"
    Each warning below is advisory. The project is created regardless; the warning tells you what to install or configure to unblock a later step.

### `Warning: "terraform" was not found on $PATH — install it before running terraform commands.`

Terraform (or OpenTofu) is not installed. You can scaffold now and [install Terraform](https://developer.hashicorp.com/terraform/install) before running the generated HCL.

### `Warning: "git" was not found on $PATH — run "git init" manually in the project directory.`

Git is not installed, so `init` skipped repository initialisation. The project is fine; run `git init` inside it once Git is available.

### `Warning: [<agent>] No Terraform or OpenTofu MCP server detected. Terraform operations in your AI agent may not work.`

Your agent has no MCP server configured for Terraform/OpenTofu. `init` does not set this up — configuring it is your responsibility. Without the Terraform server the agent cannot run Terraform. Configure the server in your agent, then reopen the project.

### `Warning: [<agent>] No BTP CLI or BTP MCP server detected. Platform validation will use user input without blocking.`

Your agent has no BTP CLI or per-agent BTP MCP route for platform validation. `init` does not set these up — configuring them is your responsibility. Without a BTP route, platform validation falls back to your input without blocking. Configure the BTP CLI or a BTP MCP server for your agent, then reopen the project.

## Other

### `git init` failed

If Git **is** installed but `git init` fails, `init` treats this as a hard error and removes the partially created project. Resolve the underlying problem (permissions, corrupt configuration) and re-run.

### The agent does not recognise the `btp-iac.*` commands

Confirm you scaffolded for the agent you are using and that its command directory exists (`.claude/commands/`, `.codex/prompts/`, `.cursor/rules/`, or `.github/instructions/`). If you recently updated the CLI, see [Updating](updating.md) to refresh the command files.
