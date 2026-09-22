# Updating

This page explains how to update the `btp-iac` CLI and how to refresh the command files in an existing project.

## Update the CLI

If you installed with `go install`, re-run it to fetch the latest version:

```sh
go install github.com/SAP/btp-iac-spec-kit/cmd/btp-iac@latest
```

If you built from source, pull and rebuild:

```sh
git pull
go build -o btp-iac ./cmd/btp-iac
```

## Refresh an existing project

!!! info "Command files are embedded at build time"
    The `btp-iac.<skill>` command files are embedded into the CLI binary when it is built. Updating the CLI does not retroactively change the command files already scaffolded into your projects.

To pick up updated commands in an existing project:

1. Update the CLI, as described above.
2. Run `init` from inside the project and let it update the agent command files in place:
   ```sh
   btp-iac init my-project --agent claude
   ```
   Because the directory is already a btp-iac project, `init` runs in "add / update AI agent" mode: it rewrites the selected agents' command files and leaves everything else alone. Run it without `--agent` in a terminal to pick the agents from a menu instead.

Your `specs/`, `memory/`, and `terraform/` content is untouched — only the agent command files are replaced.
