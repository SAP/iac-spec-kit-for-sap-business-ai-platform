# Service-parameter catalogue

`memory/service-params-catalogue.yaml` describes configuration parameters for service instances. Fresh and adopted projects receive the built-in catalogue; adopting a directory preserves an existing file, and an agent-only refresh leaves it unchanged.

The `sap-iac.services` command matches entries by `service` and plan name. For a matching service instance, it asks for every required parameter and offers each optional parameter. Supplied values are stored in `specs/services.md` and later emitted as Terraform `parameters = jsonencode(...)`.

## Convert a JSON Schema

```sh
sap-iac catalogue convert schema.json --service <offering> --plans <plan[,plan]>
```

The command converts top-level required and optional JSON Schema properties into a catalogue YAML entry and writes it to standard output. Generated entries use `source: user-defined`. Internal `$ref` values are resolved; external references are not followed.

To append the entry to the nearest sap-iac project's catalogue:

```sh
sap-iac catalogue convert schema.json \
  --service xsuaa \
  --plans application,broker \
  --append
```

With `--append`, the command walks upward from the current directory to find a project containing `specs/` or `memory/`, then appends to `memory/service-params-catalogue.yaml`. It fails if no project root is found.
