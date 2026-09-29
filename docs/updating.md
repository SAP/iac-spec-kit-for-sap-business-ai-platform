# Updating

This page explains how to update the `sap-iac` CLI and how to refresh the command files in an existing project.

## Update the CLI

If you installed with `go install`, re-run it to fetch the latest version:

```sh
go install github.com/SAP/btp-iac-spec-kit/cmd/sap-iac@latest
```

If you built from source, pull and rebuild:

```sh
git pull
go build -o sap-iac ./cmd/sap-iac
```

## Refresh an existing project

!!! info "Command files are embedded at build time"
    The `sap-iac.<skill>` command files are embedded into the CLI binary when it is built. Updating the CLI does not retroactively change the command files already scaffolded into your projects.

To pick up updated commands in an existing project:

1. Update the CLI, as described above.
2. From the parent directory, pass the existing project's directory name and the agents to refresh:
   ```sh
   sap-iac init my-project --agent claude
   ```
   Alternatively, run `sap-iac init` from inside the project and choose **Add / update AI agent in current project** from the interactive menu.

Your `specs/`, `memory/`, and `terraform/` content is untouched. The selected agent command files and `.sap-iac/platform-validation.md` are refreshed, and the managed `.gitignore` entries are ensured.
