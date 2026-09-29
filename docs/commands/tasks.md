# `sap-iac.tasks`

!!! abstract "Summary"
    **Role:** Required. · **Reads:** `specs/landscape.md`, `specs/services.md`, `specs/trust.md` (+ `specs/connectivity.md`) · **Writes:** `specs/tasks.md`

Consolidates the landscape, services, and trust specifications — and, when present, the connectivity spec — into a single, dependency-ordered execution plan.

## When to run it

After `sap-iac.accounts`, `sap-iac.services`, and `sap-iac.security` have all produced their specs (and after the optional `sap-iac.connectivity`, if you ran it).

## Inputs and outputs

| | |
|---|---|
| **Reads** | `specs/landscape.md`, `specs/services.md`, and `specs/trust.md`; and `specs/connectivity.md` when it exists. |
| **Writes** | `specs/tasks.md` — an ordered task list with task IDs and markers showing which tasks can run in parallel. |

## Behaviour

- Merges the specs into one ordered list.
- Assigns task IDs and adds parallel-execution markers.
- Creates a task for every Cloud Foundry environment, Kyma environment, and Cloud Foundry space in the landscape; each Cloud Foundry space task depends on its Cloud Foundry environment.
- Creates one `btp_subaccount_trust_configuration` task for each custom IdP trust entry. The task always carries the IdP URL as `identity_provider`; it carries `origin` only when security recorded both an explicit trust-configuration origin and its explicit marker.
- Preserves each service instance's `location`; a CF-located instance retains its `cf_space`, depends on that space, and—when the space is created in the task list—depends on every recorded CF space-role task for that space. This creates `space -> roles -> same-space resource`. If a required space has no recorded roles, stops and directs you to rerun `sap-iac.security`; BTP and organization-scoped resources are unchanged.
- Annotates each task with the stage(s) it belongs to (e.g. `dev`, `test`, `prod`). All tasks are written regardless of stage; the annotation is consumed by [`sap-iac.generate`](generate.md) to filter which tasks are generated.
- Copies `usage` and `beta_enabled` from each subaccount, `quota_required` from service entries, and service-instance `parameters` into task metadata without changing them.
- Converts a derived `APPLICATION_RUNTIME` / `MEMORY` entry into one dependent `btp_subaccount_entitlement` task, preserving its calculated `amount`.
- Writes every task unchecked (`- [ ]`); [`sap-iac.generate`](generate.md) marks tasks `- [x]` as it completes them.
- When `specs/connectivity.md` exists, appends one task per destination and certificate, ordered after the landscape, service, and trust tasks they depend on.
- Adds no new requirements — it only sequences what the earlier commands resolved.

## Example

```
/sap-iac.tasks
```

`tasks` combines the subaccount, services, and trust specs for the HR leave-request project into `specs/tasks.md`. Each group becomes its own section with a dependency/parallel table followed by a checkbox list:

```markdown
## Accounts

| ID | Task | Stage | Depends on | Parallel |
|---|---|---|---|---|
| T-001 | Create subaccount hr-leave-prod | prod | — | ✦ |
| T-002 | Entitle and create xsuaa (application) | prod | T-001 | ✦ |
| T-003 | Entitle and create hana-cloud (hana) | prod | T-001 | ✦ |
| T-004 | Create role collection HR_Leave_Employee | prod | T-002 | — |
| T-005 | Assign role Employee to HR_Leave_Employee | prod | T-004 | — |

- [ ] T-001 `[prod]` Create subaccount hr-leave-prod
- [ ] T-002 `[prod]` Entitle and create xsuaa (application)
- [ ] T-003 `[prod]` Entitle and create hana-cloud (hana)
- [ ] T-004 `[prod]` Create role collection HR_Leave_Employee
  - Task metadata: `resource_type = btp_subaccount_role_collection_base`, `collection_name = HR_Leave_Employee`, `subaccount = hr-leave-prod`
- [ ] T-005 `[prod]` Assign role Employee to HR_Leave_Employee
  - Task metadata: `resource_type = btp_subaccount_role_collection_role`, `collection_name = HR_Leave_Employee`, `subaccount = hr-leave-prod`, `role_name = Employee`, `role_template_name = Employee`, `role_template_app_id = hr-leave-xsuaa!b1`
```

## Related

- Requires [`sap-iac.accounts`](accounts.md), [`sap-iac.services`](services.md), and [`sap-iac.security`](security.md); optionally reads [`sap-iac.connectivity`](connectivity.md).
- Feeds [`sap-iac.design`](design.md) and [`sap-iac.generate`](generate.md).
