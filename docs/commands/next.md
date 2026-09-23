# `btp-iac.next`

!!! abstract "Summary"
    **Role:** Utility — run any time. · **Reads:** local project files · **Writes:** nothing

Inspects the local project state and recommends the next command in the workflow. It is a read-only helper: it looks only at which specification files exist on disk and never invokes a BTP CLI command or BTP MCP tool.

## When to run it

Whenever you lose track of where you are in the flow — after a break, when picking up someone else's project, or before deciding what to run next. It has no prerequisites and changes nothing.

## Inputs and outputs

| | |
|---|---|
| **Reads** | The presence of `memory/governance.md`, `specs/scenario.md`, `specs/landscape.md`, `specs/services.md`, `specs/trust.md`, `specs/connectivity.md`, `specs/tasks.md`, and any `terraform/**/*.tf`. |
| **Writes** | Nothing — it only prints a state summary and a recommendation. |

## Behaviour

- Walks up from the working directory to find the project root (a directory with `specs/`, `memory/`, and `terraform/`). If none is found, it tells you to run it from inside a `btp-iac init` project and stops.
- Checks each workflow output file in dependency order and picks the first missing one to recommend, treating `specs/connectivity.md` as optional.
- When both `specs/connectivity.md` and `specs/tasks.md` are missing, it emits a single two-option line offering `/btp-iac.connectivity` (optional) — define destinations and certificates — or `/btp-iac.tasks` to skip connectivity.
- Once at least one `terraform/**/*.tf` file exists, it prints a **Done** completion message ("Terraform code is in `terraform/`. Review, commit, and apply.") instead of a next-step recommendation.
- Prints all rows so you can see your full position in the flow, with `✓` for files that exist and `✗` for those that do not.

## Example

```
/btp-iac.next
```

For a project that has completed through `accounts` but not `services`, `next` prints:

```
BTP IaC — Project State
───────────────────────
✓ memory/governance.md
✓ specs/scenario.md
✓ specs/landscape.md
✗ specs/services.md
✗ specs/trust.md
✗ specs/connectivity.md
✗ specs/tasks.md
✗ terraform/**/*.tf

Next: /btp-iac.services — resolve which BTP services each subaccount needs.
```

## Related

- Recommends the next step in the workflow — see the [command overview](index.md).
- For the full sequence end to end, see the [usage walkthrough](../walkthrough.md).
