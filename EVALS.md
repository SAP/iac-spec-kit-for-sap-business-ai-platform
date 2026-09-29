# Running skill evals locally

The Go test suite validates the eval specification and fixture materialization without contacting a model:

```sh
go test ./...
```

Live runs require authenticated `codex` and/or `claude` CLIs. They are opt-in because they use model capacity and can modify only temporary fixture projects. Run both providers with Codex as the AI judge:

```sh
go run ./cmd/sap-iac-eval -live -provider all -judge codex -timeout 30m
```

Use `-provider codex` or `-provider claude` to run one agent, and `-judge claude` to use Claude as the judge. Reports are written to `eval-artifacts/<provider>-report.json`; add `-keep` to retain the temporary project paths recorded in the report.

Each case creates a fresh temporary directory from `files`, sends each `turns` entry to the agent in sequence, snapshots the resulting files, then asks the judge to score the assertions using the transcript and snapshot. Eval 4 has no files and must run in an isolated temporary directory so it has no project-root ancestor.

The judge is probabilistic. Review its `reason` fields and preserved artifacts before treating a failure as a product regression.

The command exits non-zero if a provider call fails or any judge verdict fails. It still writes each provider report first, so CI logs and local runs retain the failure evidence.
