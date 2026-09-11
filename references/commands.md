# Commands Reference

## Build

```bash
make build VERSION=x.x.x   # Compile binary to out/ with version ldflags
make install VERSION=x.x.x  # Build + install to ~/.terraform.d/plugins/registry.terraform.io/auth0/auth0/<version>/<os>_<arch>/
```

## Testing

```bash
make test-unit                           # Unit tests only — no TF_ACC, 30s timeout, outputs coverage.out
make test-acc                            # Acceptance tests with HTTP recordings (CI default)
make test-acc FILTER="TestAccMyTest"     # Run a specific recorded acceptance test
make test-acc-e2e                        # Live acceptance tests against a real Auth0 tenant (ask first)
make test-acc-record FILTER="TestName"  # Record new cassettes against a real tenant (ask first)
make test-sweep                          # Delete test resources from the Auth0 tenant (ask first)
```

## Lint & Checks

```bash
make lint          # golangci-lint run -v --fix -c .golangci.yml ./... (auto-fixes some issues)
make check-docs    # go generate and fail if docs/ has uncommitted changes (runs in CI)
make check-vuln    # govulncheck ./...
```

## Docs

```bash
make docs   # go generate → tfplugindocs; regenerates docs/ from source + templates/ + examples/
```

Never edit `docs/` files by hand — `make docs` overwrites them entirely.

## Dependencies

```bash
make deps     # go mod vendor -v
make deps-rm  # remove vendor/ directory
```

## Environment Variables (live tests)

| Variable | Purpose |
|----------|---------|
| `AUTH0_DOMAIN` | Auth0 tenant domain |
| `AUTH0_CLIENT_ID` | Client ID (machine-to-machine app) |
| `AUTH0_CLIENT_SECRET` | Client secret |
| `AUTH0_TOKEN` | Static management API token (alternative to client credentials) |
| `AUTH0_DEBUG` | Enable API debug logging |
| `AUTH0_HTTP_RECORDINGS` | Set to `on` to use cassette playback (set automatically by `make test-acc`) |
| `TF_ACC` | Set to `1` to enable acceptance tests (set automatically by all `make test-acc*` targets) |
