# Testing Reference

## Framework

`terraform-plugin-testing` (`github.com/hashicorp/terraform-plugin-testing`) with `go-vcr.v3` for HTTP cassette recording/playback.

## Test Tiers

| Tier | Command | Mode | Credentials needed? |
|------|---------|------|---------------------|
| Unit | `make test-unit` | No `TF_ACC` | None |
| Acceptance (recorded) | `make test-acc` | Cassette playback | None — uses `test/data/recordings/` |
| Acceptance (live e2e) | `make test-acc-e2e` | Real tenant | `AUTH0_DOMAIN`, `AUTH0_CLIENT_ID`, `AUTH0_CLIENT_SECRET` |
| Record new cassettes | `make test-acc-record FILTER=...` | Real tenant, writes cassettes | Same as e2e |

`make test-acc` (recordings mode) is the default in CI — no credentials required.

The `FILTER` variable works with all test targets: `make test-acc FILTER="TestAccMyResource"`.

## Test File Layout

Each resource package contains its own test file in package `<name>_test`:

```
internal/auth0/<name>/
├── resource.go
├── expand.go
├── flatten.go
└── resource_test.go   # package <name>_test
```

## Naming Convention

- Test functions: `TestAcc<ResourceOrDataSource>` — must start with `TestAcc` for the acceptance framework to pick them up.
- Config constants: descriptive names per lifecycle step, e.g. `testAccActionCreate`, `testAccActionUpdate`.
- Use `{{.testName}}` placeholder in HCL config strings — `acctest.Test()` injects the test name automatically.

## Writing a New Acceptance Test

```go
func TestAccMyResource(t *testing.T) {
    acctest.Test(t, resource.TestCase{
        Steps: []resource.TestStep{
            {
                Config: testAccMyResourceCreate,
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr("auth0_my_resource.test", "field", "value"),
                ),
            },
            {
                Config: testAccMyResourceUpdate,
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr("auth0_my_resource.test", "field", "updated"),
                ),
            },
        },
    })
}

const testAccMyResourceCreate = `
resource "auth0_my_resource" "{{.testName}}" {
  name = "Acceptance Test - {{.testName}}"
}
`
```

`acctest.Test()` automatically handles the parallel (recordings) vs sequential (live) execution mode difference.

## Cassette Workflow

1. Write the test function and config constants.
2. Run `make test-acc-record FILTER="TestAccMyResource"` against the dev tenant.
3. The cassette is saved to `test/data/recordings/TestAccMyResource.yaml`.
4. Inspect the cassette to verify sensitive values are redacted (auth headers, client secrets, domain names replaced with placeholders).
5. Commit the cassette alongside the test code.

**To re-record:** delete `test/data/recordings/<TestName>.yaml` first — `go-vcr` runs in `ModeRecordOnce` and will not overwrite an existing cassette.

## Cassette Domain

Recorded cassettes use `terraform-provider-auth0-dev.eu.auth0.com` as the tenant domain. Use the `acctest.RecordingsDomain` constant in any test config that needs to reference the domain, rather than hardcoding it.

## Coverage

`make test-unit` writes `coverage.out`. Codecov picks it up from CI. There is no enforced coverage threshold configured locally.
