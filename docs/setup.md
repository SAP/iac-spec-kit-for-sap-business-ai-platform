# Setup

This page describes the prerequisites for the `sap-iac` CLI and the supported ways to install it.

## Prerequisites

| Dependency | When you need it | Notes |
|---|---|---|
| AI agent | Always | [Claude Code](https://www.anthropic.com/claude-code), [Codex](https://openai.com/codex/), [Cursor](https://cursor.com/), [GitHub Copilot](https://github.com/features/copilot), or [Gemini CLI](https://geminicli.com/). The agent runs the generated `sap-iac.*` commands. |
| [Terraform](https://developer.hashicorp.com/terraform/install) | Recommended | The generation skill invokes the `terraform` CLI to initialise, format, and validate generated configuration. The resulting HCL may also be used with [OpenTofu](https://opentofu.org/), but the automated generation workflow currently calls Terraform. |
| [Git](https://git-scm.com/downloads) | Optional for local scaffolding; required for remote setup | When present, local fresh `init` initialises a repository automatically. Choosing the optional remote-repository flow also requires Git and access to that remote. |

!!! note "Missing Terraform or Git is not blocking for local scaffolding"
    `sap-iac init` warns and continues when Terraform or Git is missing during ordinary local scaffolding. If you choose optional remote-repository setup, Git is required to query, clone, and configure the remote.

## Install

### Package Manager

!!! Commands
    === "Windows"
        We have submitted the package to be released via `winget`. Once it is available you can install it using the command below. Please refer to the official [Winget](https://learn.microsoft.com/en-us/windows/package-manager/winget/) documentation websites for details.
        ```powershell
        winget install SAP.sap-iac
        ```

        Until it is officially available you can use a install script in this repository as a workaround to install the tool on your Windows machine. Open PowerShell and run the install script:
        ```powershell
        iwr -useb https://raw.githubusercontent.com/SAP/iac-spec-kit-for-sap-business-ai-platform/refs/heads/main/sap-iac-install.ps1 | iex
        ```
    === "Mac OS"
        Please refer to the [Homebrew](https://brew.sh/) website for details.
        ```bash
        brew tap SAP/iac-spec-kit-for-sap-business-ai-platform https://github.com/SAP/iac-spec-kit-for-sap-business-ai-platform
        brew install sap/iac-spec-kit-for-sap-business-ai-platform/sap-iac
        ```
    === "Linux"
        We’ve released `deb` and `rpm` packages to support installation on the most common Linux distributions. You can download the packages from the `assets` section of the [releases](https://github.com/SAP/iac-spec-kit-for-sap-business-ai-platform/releases) page.

        **For Debian-based distributions (like Ubuntu, Linux Mint, etc.):**

        ```bash
        sudo dpkg -i <path-to-download>/iac-spec-kit-for-sap-business-ai-platform_<latest-version>_linux_amd64.deb
        ```
        **For RPM-based distributions (like Fedora, RHEL, CentOS, openSUSE):**
        ```bash
        sudo rpm -i <path-to-download>/iac-spec-kit-for-sap-business-ai-platform_<latest-version>_linux_amd64.rpm
        ```

### Download Binaries

You can download the binaries directly from the [releases section](https://github.com/SAP/iac-spec-kit-for-sap-business-ai-platform/releases) of the GitHub repository.

Select the version that you want to use and download the binary that fits your operating system from the assets of the release. We recommend using the latest version.

### Local Build

```sh
git clone https://github.com/SAP/iac-spec-kit-for-sap-business-ai-platform.git
cd iac-spec-kit-for-sap-business-ai-platform
go build -o sap-iac ./cmd/sap-iac
```

This produces a `sap-iac` binary in the current directory. Move it onto your `PATH` (for example, `mv sap-iac /usr/local/bin/`) to run it from anywhere.

### Install with Go

```sh
go install github.com/SAP/iac-spec-kit-for-sap-business-ai-platform/cmd/sap-iac@latest
```

The binary is installed to `$(go env GOBIN)`, or `$(go env GOPATH)/bin` if `GOBIN` is unset.

!!! tip
    Ensure the install directory is on your `PATH`, otherwise your shell will not find the `sap-iac` command.

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
sap-iac --help
```

Continue to [Usage](usage.md).

For Gemini CLI, run `sap-iac init my-project --agent gemini` and open the project with `gemini`. Initialization installs both project-local slash commands and native skills; see [Gemini CLI integration](usage.md#gemini-cli-integration) for their locations and discovery commands.
