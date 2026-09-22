# Project instructions

## Highest priority: no assumptions

Do not guess. If a requirement, name, path, scope, or intent is unclear or ambiguous, **stop and ask**. This overrides every other rule below, including brevity and "just ship it" defaults.

## MCP tool preferences

- **Web fetches**: before using the built-in `WebFetch` tool, check if the `fetch` MCP server is available (tools prefixed `mcp__fetch__…`). If yes, use it. If not, fall back to `WebFetch` and mention once that the `fetch` MCP would be preferred.
- **GitHub operations** (issues, PRs, repos, file reads, searches): before using `gh` CLI or `WebFetch` against `github.com`, check if the `github` MCP server is available (tools prefixed `mcp__github__…` or `mcp__MCP_DOCKER__…`). If yes, use it. If not, fall back to `gh` and mention once that the `github` MCP would be preferred.

Do not attempt to start the MCP servers — they are the user's responsibility. Just detect and prefer.

## Keep this file current

After every change (code, config, tooling, workflow), check whether anything documented here — or that *should* be documented here — is now stale or missing. If yes, update `CLAUDE.md` in the same response, no confirmation needed. Only record durable project rules; skip conversation-specific notes.

## Project

- **Purpose**: Spec-Driven Development Toolkit for Terraform on SAP BTP
- **Module path**: `github.com/SAP/btp-iac-spec-kit`
- **Kind**: CLI tool (Go)
- **Go version**: pinned in `go.mod` (respect what's there; don't bump silently)
- **Supported agent adapters**: `claude`, `codex`, `cursor`, `copilot`
- **BTP agent safety**: Agent command files may use only BTP CLI read/list commands and BTP MCP tools explicitly documented as read/list lookups. `btp target --global-account <subdomain>` is allowed only as the account-selection prelude to those CLI calls; BTP mutations are prohibited.

## Go conventions

- Layout: consult the `golang-project-layout` skill; don't invent alternatives.
- CLI framework: consult `golang-spf13-cobra` before hand-rolling flag parsing.
- Errors: wrap with `fmt.Errorf("...: %w", err)`. No `panic` outside `main`.
- Testing: table-driven; `testify` only if already in `go.mod`, else stdlib.
- Lint/format: `gofmt`, `go vet`, `golangci-lint run` must pass before "done".
- Consult the relevant `golang-*` skill under `.claude/skills/` before writing non-trivial code (e.g. `golang-error-handling` before touching error paths). Skills override generic instincts.

## Commands

- Build: `go build ./...`
- Test: `go test ./...`
- Lint: `golangci-lint run`
- Run: _fill in once `main` exists_
- Docs: Material-themed static site under `docs/` (config `mkdocs.yml`). Built and served by **Zensical** (`zensical build` / `zensical serve`), which reads the existing `mkdocs.yml` unchanged; `mkdocs` also still builds it as a reversible fallback. Published to GitHub Pages by `.github/workflows/create-gh-page.yml` (manual `workflow_dispatch` / `workflow_call`), which runs `zensical build --clean --strict` and deploys via the GitHub Actions Pages artifact flow — so the repo's Pages source must be set to "GitHub Actions" (not a `gh-pages` branch). Mermaid diagrams render via the `pymdownx.superfences` custom fence; the `!!python/name:` tag in `mkdocs.yml` is declared in `.vscode/settings.json` so editors don't flag it. Keep user docs in `docs/`, not in the README — the README is a minimal developer entry point that links the Pages site.

## Definition of done

Compiles, tests pass, `go vet` and `golangci-lint` clean, exported identifiers have doc comments, `CLAUDE.md` still accurate.
