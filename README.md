<h1 align="center">Auth0 Terraform Provider</h1>

<div align="center">

[![GoDoc](https://pkg.go.dev/badge/github.com/auth0/terraform-provider-auth0.svg)](https://pkg.go.dev/github.com/auth0/terraform-provider-auth0)
[![Go Report Card](https://goreportcard.com/badge/github.com/auth0/terraform-provider-auth0?style=flat-square)](https://goreportcard.com/report/github.com/auth0/terraform-provider-auth0)
[![Release](https://img.shields.io/github/v/release/auth0/terraform-provider-auth0?logo=terraform&include_prereleases&style=flat-square)](https://github.com/auth0/terraform-provider-auth0/releases)
[![Codecov](https://img.shields.io/codecov/c/github/auth0/terraform-provider-auth0?logo=codecov&style=flat-square)](https://codecov.io/gh/auth0/terraform-provider-auth0)
[![License](https://img.shields.io/github/license/auth0/terraform-provider-auth0.svg?logo=fossa&style=flat-square)](https://github.com/auth0/terraform-provider-auth0/blob/main/LICENSE)
[![Build Status](https://img.shields.io/github/workflow/status/auth0/terraform-provider-auth0/Main%20Workflow/main?logo=github&style=flat-square)](https://github.com/auth0/terraform-provider-auth0/actions?query=branch%3Amain)

The Auth0 Terraform Provider is the official plugin for managing Auth0 tenant configuration through the
[Terraform](https://www.terraform.io/) tool.

</div>

---

## 📚 Documentation

- [Quickstart Guide](./docs/guides/quickstart.md)
- [Official Docs](https://registry.terraform.io/providers/auth0/auth0/latest/docs)

## 🎻 Getting Started

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
      version = ">= 0.34" # Refer to docs for latest version
    }
  }
}

provider "auth0" {}
```

```sh
$ terraform init
```

## 👋 Contributing

Feedback and contributions to this project are welcome! Before you get started, please review the following:

- [Auth0 Contribution Guidelines](https://github.com/auth0/open-source-template/blob/master/GENERAL-CONTRIBUTING.md)
- [Auth0 Contributor Code of Conduct](https://github.com/auth0/open-source-template/blob/master/CODE-OF-CONDUCT.md)
- [Contribution Guide](CONTRIBUTING.md)

## 🙇 Support & Feedback

### Raise an Issue

If you have found a bug or if you have a feature request, please raise an issue on our
[issue tracker](https://github.com/auth0/terraform-provider-auth0/issues).

### Vulnerability Reporting

Please do not report security vulnerabilities on the public GitHub issue tracker.
The [Responsible Disclosure Program](https://auth0.com/whitehat) details the procedure for disclosing security issues.


---

<div align="center">

<img alt="Auth0 logo and word-mark in black on transparent background" src="https://user-images.githubusercontent.com/28300158/183676042-b9d92893-8fff-408f-9a36-63e77b14be30.png#gh-light-mode-only"  width="20%" height="20%">

<img alt="Auth0 logo and word-mark in white on transparent background" src="https://user-images.githubusercontent.com/28300158/183676141-bea463f9-af82-40ce-b18c-3a1030183d58.png#gh-dark-mode-only"  width="20%" height="20%">

</div>

<br/>

<div align="center">

Auth0 is an easy to implement, adaptable authentication and authorization platform. To learn more checkout
[Why Auth0?](https://auth0.com/why-auth0)

This project is licensed under the MPL-2.0 license. See the [LICENSE](LICENSE) file for more info or
[auth0-terraform-provider.pdf](https://www.okta.com/sites/default/files/2022-03/auth0-terraform-provider.pdf) for a full
report.

</div>
