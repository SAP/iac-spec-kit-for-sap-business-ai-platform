# FAQ

## Responsibilities and team structure

### Who is responsible for what when using the spec-kit?

The spec-kit assists you through the workflow but does not make decisions for you or touch your BTP account. Here is where your responsibility starts and ends at each step:

| Step | The spec-kit does | You are responsible for |
|---|---|---|
| `sap-iac init` | Scaffolds the project, installs agent commands | Choosing agents, providing global-account subdomain, configuring MCP servers |
| `sap-iac.govern` | Guides you through guardrail categories, writes `memory/governance.md` | Deciding the actual rules: regions, naming patterns, cost policies |
| `sap-iac.scenario` through `sap-iac.tasks` | Structures your inputs into spec files | Providing accurate inputs, reviewing and correcting spec files before moving on |
| `sap-iac.design` | Proposes a state-chunking architecture | Approving or overriding the architecture before `generate` runs |
| `sap-iac.generate` | Writes and validates Terraform HCL | Reviewing the generated files, committing them, managing credentials |
| `terraform plan` / `terraform apply` | Nothing — the spec-kit stops here | Running the plan, approving the apply, operating the provisioned resources |

The spec files in `specs/` and `memory/` are yours: read them, edit them, commit them, and use them to collaborate. The spec-kit writes them on your behalf but they are the source of truth for your project — not the agent session.

### Our organisation runs different steps in different teams — where do we draw the line?

The spec-kit is designed to work across team boundaries. The handover mechanism is a shared Git repository: each team commits its outputs, and the next team picks up from there.

A common split:

| Team | Steps they run | What they commit |
|---|---|---|
| Platform / architecture | `sap-iac.govern`, `sap-iac.scenario`, `sap-iac.analyse`, `sap-iac.accounts`, `sap-iac.services` | `memory/governance.md`, `specs/scenario.md`, `specs/landscape.md`, `specs/services.md` |
| Infrastructure engineering | `sap-iac.security`, `sap-iac.connectivity`, `sap-iac.tasks`, `sap-iac.design`, `sap-iac.generate` | `specs/trust.md`, `specs/connectivity.md`, `specs/tasks.md`, `terraform/` |
| Operations / SRE | `terraform plan` / `terraform apply` | state, pipeline configuration |

**How to make it work in practice:**

- Each team member runs `sap-iac init <repo> --agent <their-agent>` to install their own agent's command files. The spec files and governance file travel with the repo — no session state is transferred between teams.
- Use pull requests to gate handovers. The upstream team opens a PR; the downstream team reviews the spec files and merges before continuing.
- When picking up a project mid-flow, run `/sap-iac.next` — it reads the existing spec files and tells you exactly where things stand and what to run next.
- If the boundary between two teams falls mid-step (for example, both teams want to review the design before generation), use a draft PR or a feature branch for that review before merging to the main branch.

The split above is a suggestion, not a requirement. You can draw the line wherever your team structure demands — the spec files are the contract between teams.

---

## Before you start

### Do I need a BTP account before running anything?

You need a BTP global account to run `terraform apply`, but not to use the spec-kit itself. You can scaffold a project, run the full AI-agent workflow, and generate Terraform without a live account. The only steps that need a live account are the optional BTP CLI / BTP MCP validation checks during `sap-iac.govern` and `sap-iac.scenario` — and if you skip those or have no account subdomain configured, the commands continue without blocking. Enter the global-account subdomain when `init` prompts for it; leave it blank to skip live validation.

### Is `sap-iac.govern` mandatory?

No. Without it the downstream commands run without governance enforcement — they will not block on region, naming, or cost-centre violations. That is fine for exploration or small projects. For anything production-bound, running `sap-iac.govern` first means every downstream command validates against your rules automatically.

### Does Terraform need to be installed on my machine?

Recommended, not required for scaffolding. The `sap-iac init` command warns and continues when Terraform is missing. The generation step (`sap-iac.generate`) calls `terraform init`, `terraform fmt`, and `terraform validate` after writing the HCL — so you need Terraform (or the Terraform MCP server) available in your agent's environment for that step. You can still review the raw HCL and run Terraform yourself afterwards.

---

## Running the workflow

### Where do the spec files live, and can I read or edit them?

All spec files are plain Markdown files in the `specs/` directory of your project. You can open, read, and edit them at any point — they are designed to be human-readable. The key files are:

| File | Written by |
|---|---|
| `memory/governance.md` | `sap-iac.govern` |
| `specs/scenario.md` | `sap-iac.scenario` |
| `specs/landscape.md` | `sap-iac.accounts` |
| `specs/services.md` | `sap-iac.services` |
| `specs/trust.md` | `sap-iac.security` |
| `specs/connectivity.md` | `sap-iac.connectivity` |
| `specs/tasks.md` | `sap-iac.tasks` / `sap-iac.design` |

Editing a spec file and re-running the next command is the normal way to course-correct.

### I lost track of where I am in the workflow. What do I do?

Run `/sap-iac.next`. It inspects which spec files exist on disk and prints your current position with a recommendation for the next command. It is read-only and changes nothing.

### I made a mistake in an earlier step — do I have to start over?

No. Each command reads and overwrites its own output file. To redo a step, just re-run the relevant command. For example, if `specs/scenario.md` is wrong, run `/sap-iac.scenario` again with a corrected description — it overwrites the file and downstream commands will pick up the updated content when you re-run them too. `specs/`, `memory/`, and `terraform/` are never touched by `sap-iac init`, so re-scaffolding is not needed.

### I skipped `sap-iac.govern` — can I add it later?

Yes. Run `/sap-iac.govern` at any point. Once `memory/governance.md` exists, all governance-aware commands (`accounts`, `services`, `security`, `generate`) start enforcing it from that point forward. If you have already generated Terraform, re-run `generate` to apply the guardrails to the output.

### A command blocked on a governance violation. How do I override it?

Add `- Override: true` to `memory/governance.md`. This turns hard blocks into warnings so the command proceeds. Use it deliberately — it disables enforcement for the entire project until you remove the line. Deprecation warnings are the one exception: `- Override: true` does not suppress them.

### Where is the generated Terraform, and what do I do with it?

It lands in the `terraform/` directory of your project, in subdirectories matching the state-chunking layout from `sap-iac.design` (for example `terraform/prod/btp/` and `terraform/prod/cf/`). The files are left uncommitted — the agent never commits or pushes unless you explicitly ask. Review them, commit, and then run:

```sh
cd terraform/prod/btp
terraform init
terraform plan
```

### Can I regenerate just one stage?

Yes. When you run `/sap-iac.generate`, it asks which stage(s) to generate. Answer with the stage name (e.g. `dev`, `prod`) or `all`. Tasks outside the chosen stage are skipped and left as spec-only entries in `specs/tasks.md`.

---

## Working with multiple agents or teammates

### My teammate uses Codex; I use Claude. Same project — how does that work?

Both agents can work from the same repository. Each person runs `sap-iac init` against the shared repo with their own `--agent` flag:

```sh
# You
sap-iac init my-project --agent claude

# Your teammate
sap-iac init my-project --agent codex
```

`init` adds command files for the specified agent without touching the other agent's files or any `specs/`, `memory/`, or `terraform/` content. Both sets of command files coexist in the repo. Each person opens the project in their own agent and runs the same `/sap-iac.*` commands; the spec files they produce are identical regardless of which agent wrote them.

### My teammate updated the CLI. Do I need to update mine too?

Not immediately — each person's CLI version only affects the command files it installs. But the safest approach is to keep versions in sync. After updating, re-run `sap-iac init` to refresh the command files:

```sh
sap-iac init my-project --agent claude
```

This overwrites only the agent command files with the latest version. `specs/`, `memory/`, and `terraform/` are untouched. See [Updating](updating.md) for the full update steps.

### How do I add a second agent to a project that is already set up?

Run `sap-iac init` again with the new agent:

```sh
sap-iac init my-project --agent codex
```

Or run it from inside the project directory and choose **Add / update AI agent in current project** from the interactive menu. Only the new agent's files are added; everything else is left unchanged.

### My teammate set up governance and ran the scenario. I need to pick up from there. What do I do?

Pull the latest changes from the shared repository so you have the current `specs/` and `memory/` files. Run `sap-iac init my-project --agent <your-agent>` to install your agent's command files if you have not already. Then run `/sap-iac.next` inside your agent — it will read the existing spec files and tell you exactly which command to run next.
