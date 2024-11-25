# Contributing

We appreciate feedback and contribution to this provider.
Before you submit a pull request, there are a couple requirements to satisfy.

## Prerequisites

- [Go 1.18+](https://go.dev/)
- [Terraform](https://developer.hashicorp.com/terraform/downloads)

## Reading Material

- [Extending Terraform](https://www.terraform.io/docs/extend/index.html) - design, develop, and test plugins that connect Terraform to external services.
- [Writing Custom Providers](https://learn.hashicorp.com/collections/terraform/providers) - tutorials on interacting with APIs using Terraform providers.
- [HashiCorp Provider Design Principles](https://www.terraform.io/docs/extend/hashicorp-provider-design-principles.html) - explore the underlying principles for the design choices of this provider.

## Getting Started

To work on the provider, you'll need [Go](https://go.dev/) installed on your machine (version 1.18+ is *required*).
You'll also need to set up a [`$GOPATH`](https://go.dev/doc/code.html#GOPATH), and add `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make install VERSION=X.X.X`. This will build the provider and install the provider binary
in the `${HOME}/.terraform.d/plugins` directory, so it can be used directly in terraform `required_providers` block.

```shell
make install VERSION=0.2.0
...
# On macOS, this will install the provider in a location like the following:
# ~/.terraform.d/plugins/registry.terraform.io/auth0/auth0/0.2.0/darwin_amd64/terraform-provider-auth0_v0.2.0
...
```

```hcl
terraform {
    required_providers {
        auth0 = {
            source  = "auth0/auth0"
            version = "0.2.0"
        }
    }
}
```

## Documentation

The documentation found in the [docs](./docs) folder is generated using
[terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs).
Do not edit files within this folder manually.

Run `make docs` to regenerate documentation for newly added resources and schema attributes or if changes are needed to existing schemas.

```shell
make docs
```

## Signing your Commits
We required all commits on the contributing PR to be signed.

- [Learn more about signing commits](https://docs.github.com/en/authentication/managing-commit-signature-verification/signing-commits)
- [Signing old commits](https://stackoverflow.com/questions/41882919/is-there-a-way-to-gpg-sign-all-previous-commits)


## Running the Tests

The tests can be run using the following make commands:

- `make test-unit` - runs all the unit tests.
- `make test-acc` - runs the tests with http recordings. To run a specific test pass the `FILTER` var. Usage `make test-acc FILTER="TestAccResourceServer"`.
- `make test-acc-e2e` - runs the tests against a real Auth0 tenant. To run a specific test pass the `FILTER` var. Usage `make test-acc-e2e FILTER="TestAccResourceServer"`.

> **Note**
> The http test recordings can be found in the [recordings](./test/data/recordings) folder.

To run the tests against an Auth0 tenant start by creating an
[M2M app](https://auth0.com/docs/applications/set-up-an-application/register-machine-to-machine-applications) in the
tenant, that has been authorized to request access tokens for the Management API and has all the required permissions.

Then set the following environment variables on your machine:

* `AUTH0_DOMAIN`: The **Domain** of the M2M app
* `AUTH0_CLIENT_ID`: The **Client ID** of the M2M app
* `AUTH0_CLIENT_SECRET`: The **Client Secret** of the M2M app
* `AUTH0_DEBUG`: Set to `true` to call the Management API in debug mode, which dumps the HTTP requests and responses to the output

> **Warning** 
> The e2e acceptance tests make calls to a real Auth0 tenant, and create real resources. 
> Certain tests also require a paid Auth0 subscription to be able to run successfully,
> e.g. `TestAccCustomDomain` and the ones starting with `TestAccLogStream*`.

> **Note** 
> At the time of writing, the following configuration steps are also required for the test tenant:
> - The `Username-Password-Authentication` connection must have _Requires Username_ option enabled for the user tests to successfully run.

## Adding New HTTP Test Recordings

When creating a new Terraform resource or adding a new resource property, the http recordings will need to be re-recorded. 
To add new http test recordings or to regenerate old ones, use the `make test-acc-record` command.

> **Warning**
> If you need to regenerate an old recording, make sure to delete the corresponding recording file first.

To add only one specific http test recording pass the `FILTER` var, for example `make test-acc-record FILTER="TestAccResourceServer"`.

> **Warning**
> Recording a new http test interaction will make use of a real Auth0 test tenant. 

## Resetting the Test Tenant

All resources created through running the tests against an Auth0 tenant can be removed by running `make test-sweep`.
