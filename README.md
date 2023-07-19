<div align="center">
  <h1>Auth0 Terraform Provider</h1>

[![GoDoc](https://pkg.go.dev/badge/github.com/auth0/terraform-provider-auth0.svg)](https://pkg.go.dev/github.com/auth0/terraform-provider-auth0)
[![Go Report Card](https://goreportcard.com/badge/github.com/auth0/terraform-provider-auth0?style=flat-square)](https://goreportcard.com/report/github.com/auth0/terraform-provider-auth0)
[![Release](https://img.shields.io/github/v/release/auth0/terraform-provider-auth0?logo=terraform&include_prereleases&style=flat-square)](https://github.com/auth0/terraform-provider-auth0/releases)
[![Codecov](https://img.shields.io/codecov/c/github/auth0/terraform-provider-auth0?logo=codecov&style=flat-square)](https://codecov.io/gh/auth0/terraform-provider-auth0)
[![License](https://img.shields.io/github/license/auth0/terraform-provider-auth0.svg?logo=fossa&style=flat-square)](https://github.com/auth0/terraform-provider-auth0/blob/main/LICENSE)
[![Build Status](https://img.shields.io/github/actions/workflow/status/auth0/terraform-provider-auth0/main.yml?branch=main)](https://github.com/auth0/terraform-provider-auth0/actions?query=branch%3Amain)

</div>

-------------------------------------

The Auth0 Terraform Provider is the official plugin for managing Auth0 tenant configuration through the
[Terraform](https://www.terraform.io/) tool.

📚 [Documentation](#documentation) • 🚀 [Getting Started](#getting-started) • 💬 [Feedback](#feedback)

-------------------------------------

## Documentation

- [Official Docs](https://registry.terraform.io/providers/auth0/auth0/latest/docs)
- Guides
  - [Quickstart](./docs/guides/quickstart.md)
  - [List available triggers for actions](./docs/guides/action_triggers.md)
  - [Zero downtime client credentials rotation](./docs/guides/client_secret_rotation.md)

## Getting Started

### Requirements

- [Terraform](https://www.terraform.io/downloads)
- An [Auth0](https://auth0.com) account

### Installation

Terraform uses the [Terraform Registry](https://registry.terraform.io/) to download and install providers. To install
this provider, copy and paste the following code into your Terraform configuration. Then, run `terraform init`.

```terraform
terraform {
  required_providers {
    auth0 = {
      source  = "auth0/auth0"
      version = "1.0.0-beta.0" # Refer to docs for latest version
    }
  }
}

provider "auth0" {}
```

```shell
$ terraform init
```

## Feedback

### Contributing

We appreciate feedback and contribution to this repo! Before you get started, please see the following:

- [Contribution Guide](./CONTRIBUTING.md)
- [Auth0's General Contribution Guidelines](https://github.com/auth0/open-source-template/blob/master/GENERAL-CONTRIBUTING.md)
- [Auth0's Code of Conduct Guidelines](https://github.com/auth0/open-source-template/blob/master/CODE-OF-CONDUCT.md)

### Raise an issue

To provide feedback or report a bug, [please raise an issue on our issue tracker](https://github.com/auth0/terraform-provider-auth0/issues).

### Vulnerability reporting

Please do not report security vulnerabilities on the public GitHub issue tracker.
The [Responsible Disclosure Program](https://auth0.com/responsible-disclosure-policy) details the procedure for disclosing security issues.

---

<div align="center">
  <picture>
    <source media="(prefers-color-scheme: light)" srcset="https://cdn.auth0.com/website/sdks/logos/auth0_light_mode.png" width="150">
    <source media="(prefers-color-scheme: dark)" srcset="https://cdn.auth0.com/website/sdks/logos/auth0_dark_mode.png" width="150">
    <img alt="Auth0 Logo" src="https://cdn.auth0.com/website/sdks/logos/auth0_light_mode.png" width="150">
  </picture>
</div>

<div align="center">

Auth0 is an easy to implement, adaptable authentication and authorization platform. To learn more checkout
[Why Auth0?](https://auth0.com/why-auth0)

This project is licensed under the MPL-2.0 license. See the [LICENSE](LICENSE) file for more info or
[auth0-terraform-provider.pdf](https://www.okta.com/sites/default/files/2022-03/auth0-terraform-provider.pdf) for a full
report.

</div>
