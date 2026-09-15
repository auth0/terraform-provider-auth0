# Docs Update Rules

## Tracked Docs

| Doc | What it covers | Exists |
|-----|---------------|--------|
| `README.md` | Installation, Terraform Registry source, quickstart links, feedback channels | ✅ Present |
| `examples/resources/<name>/` | HCL usage examples per resource (one `.tf` file per resource) | ✅ Present |
| `examples/data-sources/<name>/` | HCL usage examples per data source | ✅ Present |
| `docs/` | Full resource and data-source reference docs (auto-generated — never edit manually) | ✅ Present |

`EXAMPLES.md` does not exist; the `examples/` directory serves that role.

## Code-to-Docs Mapping

This is a Terraform provider. The "public surface" is resources, data sources, and provider configuration options.

| When this changes | Update these docs |
|-------------------|-------------------|
| New resource added | Create `examples/resources/<name>/resource.tf`; run `make docs` to generate `docs/resources/<name>.md` |
| New data source added | Create `examples/data-sources/<name>/data-source.tf`; run `make docs` to generate `docs/data-sources/<name>.md` |
| Schema field added or changed on a resource | Update `examples/resources/<name>/resource.tf` to show the field; run `make docs` |
| Schema field added or changed on a data source | Update `examples/data-sources/<name>/data-source.tf`; run `make docs` |
| Provider-level configuration option added or changed | Update `README.md` if it affects installation or authentication setup; run `make docs` |
| Resource or data source removed | Remove the `examples/` directory for that resource/data source; run `make docs` |

> **Always run `make docs` in the same PR as the schema change.** The CI `check-docs` job will fail the PR if `docs/` is out of sync with the current source.

## Docs Generation Pipeline

`make docs` runs two steps via `go generate`:

1. `terraform fmt -recursive ./examples/` — formats all HCL example files.
2. `go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs` — generates `docs/` from:
   - Schema `Description` fields in `internal/auth0/<name>/resource.go` (Markdown supported)
   - `templates/resources/<name>.md.tmpl` (optional structural overrides in `templates/`)
   - `examples/resources/<name>/resource.tf` (embedded as the usage example)
