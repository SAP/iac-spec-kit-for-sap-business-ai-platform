# `btp-iac.design`

!!! abstract "Summary"
    **Role:** Required. · **Reads:** `specs/tasks.md` · **Writes:** `specs/tasks.md` (in place)

Translates the task list into a concrete Terraform folder structure and annotates each task with the file it will be written to.

## When to run it

After `btp-iac.tasks`, and before `btp-iac.generate`.

## Inputs and outputs

| | |
|---|---|
| **Reads** | `specs/tasks.md`. |
| **Writes** | `specs/tasks.md`, updated in place. It does not create a separate output file. |

## Behaviour

- Decides how resources are split across files.
- Decides whether modules are introduced.
- Decides how per-environment variable files are organised.
- Annotates each task with its target file path, which `generate` then follows.
- Performs no BTP access for its core function — it only reads and rewrites `specs/tasks.md`, unlike the BTP-touching commands.

## Example

```
/btp-iac.design
```

`design` adds a file-path annotation to each task in `specs/tasks.md`:

```markdown
1. [T1] Create subaccount hr-leave-prod          -> subaccount.tf
2. [T2] Create xsuaa (application)               -> services.tf
3. [T3] Create hana-cloud (hana)                 -> services.tf
4. [T4] Create role collection HR_Leave_Employee -> security.tf
```

For a single small application it keeps everything in a flat set of files; a larger, multi-subaccount scenario is where it introduces modules and per-environment variable files.

## Related

- Requires [`btp-iac.tasks`](tasks.md).
- Feeds [`btp-iac.generate`](generate.md).
