# /btp-iac:services

Resolves the infrastructure requirements from `specs/scenario.md` into a dependency-ordered list of BTP entitlements, subscriptions, and service instances per subaccount.

Reads `specs/scenario.md` and `specs/landscape.md`. Produces `specs/services.md` for the team to review and adjust before any code is generated.

---

## Governance Check

**Before resolving any service dependencies**, locate the btp-iac project root by walking up from the current working directory until a directory containing `specs/`, `memory/`, and `terraform/` is found. Then check whether `<project-root>/memory/governance.md` exists.

**If it does not exist:** proceed without constraints and note: "No governance rules found — proceeding without enforcement."

**If it exists**, load all rules and validate every service plan decision against the `## Service Plans` section.

### Service plan validation
For each service instance you are about to define, identify its environment tier from `specs/landscape.md` and check the plan:
- If the plan is **not** in the `Permitted` list for that tier: **STOP**
  > "GOVERNANCE VIOLATION: Plan `<plan>` is not permitted for environment `<tier>` (permitted: `<permitted-plans>`). Change the plan or add `- Override: true` to memory/governance.md."
- If the plan appears in the `Forbidden` list for that tier: **STOP** with the same message.

**If `- Override: true` is set in `<project-root>/memory/governance.md`:** log a warning for each violation and continue instead of stopping.

---

## Workflow

Read `specs/scenario.md` and `specs/landscape.md` to understand the application requirements and the account topology.

For each subaccount, resolve:
- Required entitlements (service + plan)
- Subscriptions (SaaS applications)
- Service instances with configuration parameters
- Dependencies between services (ordered)

Write `specs/services.md` with the full dependency-ordered list. This file is the direct input to `/btp-iac:tasks`.
