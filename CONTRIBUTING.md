# Contributing

Before you begin, read through the Terraform documentation on
[Extending Terraform](https://www.terraform.io/docs/extend/index.html) and
[Writing Custom Providers](https://learn.hashicorp.com/collections/terraform/providers).

Finally,
the [HashiCorp Provider Design Principles](https://www.terraform.io/docs/extend/hashicorp-provider-design-principles.html)
explore the underlying principles for the design choices of this provider.

## Prerequisites

- [Go 1.18+](https://go.dev/)

## Getting started

To work on the provider, you'll need [Go](http://www.golang.org) installed on your machine(version 1.18+ is *required*).
You'll also need to correctly set up a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin`
to your `$PATH`.

To compile the provider, run `make install VERSION=X.X.X`. This will build the provider and install the provider binary
in the `${HOME}/.terraform.d/plugins` directory, so it can be used directly in terraform `required_providers` block.

```sh
make install VERSION=0.2.0
...
~/.terraform.d/plugins/registry.terraform.io/auth0/auth0/0.2.0/darwin_amd64/terraform-provider-auth0_v0.2.0
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

## Running tests

To run the tests use the `make test` command. This will make use of the pre-recorded http interactions found in the
[recordings](./test/data/recordings) folder. To add new recordings run the tests against an Auth0 tenant
individually using the following env vars `AUTH0_HTTP_RECORDINGS=on TF_ACC=true`.

To run the tests against an Auth0 tenant, use the `make test-acc-e2e` command. Start by creating an
[M2M app](https://auth0.com/docs/applications/set-up-an-application/register-machine-to-machine-applications) in the
tenant, that has been authorized to call the Management API and has all the required permissions.

Then set the following environment variables:

* `AUTH0_DOMAIN`: The **Domain** of the M2M app
* `AUTH0_CLIENT_ID`: The **Client ID** of the M2M app
* `AUTH0_CLIENT_SECRET`: The **Client Secret** of the M2M app
* `AUTH0_DEBUG`: Set to `true` to call the Management API in debug mode, which dumps the HTTP requests and responses to
  the output

**Note:** The e2e acceptance tests make calls to a real Auth0 tenant, and create real resources. Certain tests also
require a paid Auth0 subscription to be able to run successfully, e.g. `TestAccCustomDomain` and the ones starting with
`TestAccLogStream*`.

**Note:** At the time of writing, the following configuration steps are also required for the test tenant:

* The `Username-Password-Authentication` connection must have _Requires Username_ option enabled for the user tests to
  successfully run.

## Documentation

The documentation found in the [docs](./docs) folder is generated using
[terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs). Please run `make docs` to regenerate
documentation for newly added resources or schema attributes.

```sh
make docs
```
