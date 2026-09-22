# Setup

This page describes the prerequisites for the `btp-iac` CLI and the supported ways to install it.

## Prerequisites

| Dependency | When you need it | Notes |
|---|---|---|
| AI agent | Always | [Claude Code](https://www.anthropic.com/claude-code), [Codex](https://openai.com/codex/), [Cursor](https://cursor.com/), or [GitHub Copilot](https://github.com/features/copilot). The agent runs the generated `btp-iac.*` commands. |
| [Terraform](https://developer.hashicorp.com/terraform/install) | Recommended | Required later to apply the configuration the agent generates. [OpenTofu](https://opentofu.org/) is also supported. |
| [Git](https://git-scm.com/downloads) | Optional | When present, `init` initialises a repository in the new project automatically. |

!!! note "Terraform and Git are never blocking"
    `btp-iac init` does not hard-fail when Terraform or Git is missing. It surfaces a warning and continues, so you can scaffold first and install them when convenient.

## Install

### Build from source

```sh
git clone https://github.com/SAP/btp-iac-spec-kit.git
cd btp-iac-spec-kit
go build -o btp-iac ./cmd/btp-iac
```

This produces a `btp-iac` binary in the current directory. Move it onto your `PATH` (for example, `mv btp-iac /usr/local/bin/`) to run it from anywhere.

### Install with Go

```sh
go install github.com/SAP/btp-iac-spec-kit/cmd/btp-iac@latest
```

The binary is installed to `$(go env GOBIN)`, or `$(go env GOPATH)/bin` if `GOBIN` is unset.

!!! tip
    Ensure the install directory is on your `PATH`, otherwise your shell will not find the `btp-iac` command.

### Make targets

From a source checkout, the `Makefile` wraps the common tasks:

| Target | Action |
|---|---|
| `make build` | Compile the CLI (`go build -v ./...`). |
| `make install` | Build and install all commands (`go install ./...`)
| `make test` | Run the test suite. |
| `make lint` | Run `golangci-lint`. |
| `make fmt` | Format the code with `gofmt`. |

## Verify

Confirm the CLI is installed and runnable:

```sh
btp-iac --help
btp-iac init --help
```

You should see the `init` command and its `--agent` flag. Continue to [Usage](usage.md).
