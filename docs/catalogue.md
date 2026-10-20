# Service-parameter catalogue

`memory/service-params-catalogue.yaml` describes configuration parameters for service instances. Fresh and adopted projects receive the built-in catalogue; adopting a directory preserves an existing file, and an agent-only refresh leaves it unchanged.

The `sap-iac.services` command matches entries by `service` and plan name. For a matching service instance, it asks for every required parameter and offers each optional parameter. Supplied values are stored in `specs/services.md` and later emitted as Terraform `parameters = jsonencode(...)`.

## Convert a JSON Schema

```sh
sap-iac catalogue convert schema.json --service <offering> --plans <plan[,plan]>
```

The command requires a valid JSON file and converts top-level required and optional JSON Schema properties into a catalogue YAML entry. Generated entries use `source: user-defined` and are written to standard output unless `--append` is supplied. The `--service` and `--plans` values name the output entry; they are required and `--plans` accepts a comma-separated list.

Conversion preserves useful schema information as catalogue examples:

- A property's `default` is preferred; otherwise the first `enum` value is used, followed by a placeholder.
- Nested objects become nested example maps and arrays receive a one-item placeholder example.
- `oneOf` and `anyOf` use their first branch and are marked with a YAML comment.
- Internal `$ref` values under `$defs` or `definitions` are resolved. External or unresolvable references remain explicit placeholders and are not fetched.

The converter also accepts a BTP service-catalog wrapper containing `servicePlans`. In that form, it reads the first service plan's `service_instance.create.parameters` schema. Choose the output `--plans` values deliberately: they are not inferred from that wrapper.

To append the entry to the nearest sap-iac project's catalogue:

```sh
sap-iac catalogue convert schema.json \
  --service xsuaa \
  --plans application,broker \
  --append
```

With `--append`, the command walks upward from the current directory to find a project containing `specs/` or `memory/`, then appends to `memory/service-params-catalogue.yaml`. It fails if no project root is found.
