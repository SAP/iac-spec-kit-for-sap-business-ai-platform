# Development Setup

If you want to contribute to the Infrastructure-as-Code Specification Toolkit for SAP Business AI Platform be aware of the [contribution guidelines](CONTRIBUTING.md) available in this repository.

First, you need to set up your development environment. The following sections describe the options you have.

## GitHub Codespaces

**Step 1:** Open the repository in GitHub Codespaces via the button:

[![Open in GitHub Codespaces](https://github.com/codespaces/badge.svg)](https://github.com/codespaces/new?hide_repo_select=true&ref=main&repo=1334471983)

**Step 2:** There is no step 2 😎.

The Codespace uses the configuration in [.devcontainer/default/devcontainer.json](.devcontainer/default/devcontainer.json), which provisions Go, Terraform, OpenTofu, and the GitHub CLI.

## Dev Container

> **Note**: To use dev containers, you must have a container runtime running on the machine. See the official documentation about [Developing inside a Container](https://code.visualstudio.com/docs/devcontainers/containers).

Clone the repository:

```bash
git clone https://github.com/SAP/btp-iac-spec-kit.git
```

Open the cloned repository in [Visual Studio Code](https://code.visualstudio.com/), press the "Open a Remote Window" button in the lower-left corner, and choose "Reopen in Container". This starts the dev container defined in [.devcontainer/default/devcontainer.json](.devcontainer/default/devcontainer.json) with Go, Terraform, OpenTofu, and the GitHub CLI preinstalled.

> **Note**: On the first run, downloading the container image might take a while — maybe time to grab a cup of coffee ☕.

## Local Setup

Ensure you have the following tools installed on your local machine.

* [git](https://git-scm.com/)
* [go](https://go.dev/) (the version pinned in [go.mod](go.mod))
* [golangci-lint](https://github.com/golangci/golangci-lint)
* [make](https://www.gnu.org/software/make/)
* [terraform](https://www.terraform.io/) or [OpenTofu](https://opentofu.org/) (optional — only needed to run the generated configuration)

### macOS (Homebrew)

```bash
brew install git golang golangci-lint make terraform
```

### Windows (Chocolatey)

```bash
choco install git golang golangci-lint make terraform
```

### Cloning the Repository

```bash
git clone https://github.com/SAP/btp-iac-spec-kit.git
```

Navigate into the directory of the cloned repository.

## Build and Install Locally

The [Makefile](Makefile) wraps the common tasks:

| Target | Action |
|---|---|
| `make build` | `go build -v ./...` |
| `make install` | Build and `go install -v ./...` (installs `sap-iac` to `$(go env GOPATH)/bin`) |
| `make test` | Run the test suite (`go test -v -cover -tags=all ./...`) |
| `make lint` | Run `golangci-lint` |
| `make fmt` | Format the code with `gofmt` |

To build and install the CLI:

```bash
make install
```

## Verify the Setup

Confirm the `sap-iac` binary is on your `PATH` and runs:

```bash
sap-iac --help
sap-iac init --help
```

You should see the `init` command and its `--agent` flag. If `sap-iac` is not found, make sure `$(go env GOPATH)/bin` is on your `PATH`.

If you are still stuck, feel free to ask for support by raising a [question](https://github.com/SAP/btp-iac-spec-kit/discussions/) in the [GitHub Discussions](https://github.com/SAP/btp-iac-spec-kit/discussions/) of this repository.

## Updating

To update an existing installation of the CLI and refresh the scaffolded command files in a project, follow the [Updating guide](docs/updating.md).

## Documentation

User-facing documentation is a Material-themed site under [docs/](docs/), configured in [mkdocs.yml](mkdocs.yml). It is built with [Zensical](https://zensical.org/) (by the Material for MkDocs team), which reads the existing `mkdocs.yml` unchanged. Preview it locally with:

```sh
python3 -m venv .venv && source .venv/bin/activate
pip install zensical
zensical serve      # live preview at http://127.0.0.1:8000
zensical build --clean --strict   # one-off build into ./site
```

It is published to GitHub Pages by the [`create-gh-page`](.github/workflows/create-gh-page.yml) workflow, which runs `zensical build` and deploys via the GitHub Actions Pages artifact flow. (`mkdocs build` still works too, as a reversible fallback.)

## How to Commit

Once you're done, ensure the tests still pass (`make test`) and the code is linted and formatted (`make lint`, `make fmt`). Then open a pull request. We follow the [conventional commits specification](https://www.conventionalcommits.org/en/v1.0.0/) — enforced on PR titles by the [Semantic PR Check](.github/workflows/semantic-pr.yml) workflow — so the pull-request title has to be structured like:

* `fix: typo in the documentation`
* `feat: add a new scaffold agent`
* `refactor!: rename a command flag`
* `feat(init): a scoped feature`

For more examples, see the [conventional commits specification](https://www.conventionalcommits.org/en/v1.0.0/).
