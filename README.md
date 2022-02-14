# Auth0 Terraform Provider

[![GoDoc](https://pkg.go.dev/badge/github.com/auth0/terraform-provider-auth0.svg)](https://pkg.go.dev/github.com/auth0/terraform-provider-auth0)
[![License](https://img.shields.io/github/license/auth0/terraform-provider-auth0.svg?style=flat-square)](https://github.com/auth0/terraform-provider-auth0/blob/main/LICENSE)
[![Release](https://img.shields.io/github/v/release/auth0/terraform-provider-auth0?include_prereleases&style=flat-square)](https://github.com/auth0/terraform-provider-auth0/releases)
[![Build Status](https://img.shields.io/endpoint.svg?url=https%3A%2F%2Factions-badge.atrox.dev%2Fauth0%2Fterraform-provider-auth0%2Fbadge%3Fref%3Dmain&style=flat-square)](https://github.com/auth0/terraform-provider-auth0/actions?query=branch%3Amain)

---

Terraform Provider for the [Auth0](https://auth0.com/) platform.

_Note: This Provider was previously maintained under
[alexkappa/terraform-provider-auth0](https://github.com/alexkappa/terraform-provider-auth0)._

-------------------------------------

## Table of Contents

- [Installation](#installation)
- [Documentation](#documentation)
- [Usage](#usage)
- [Contributing](#contributing)
- [What is Auth0?](#what-is-auth0)
- [Create a free Auth0 Account](#create-a-free-auth0-account)
- [Issue Reporting](#issue-reporting)
- [Author](#author)
- [License](#license)

-------------------------------------

## Installation

**Terraform 0.13+**

Terraform 0.13 and higher uses the [Terraform Registry](https://registry.terraform.io/) to download and install
providers. To install this provider, copy and paste this code into your Terraform configuration.
Then, run `terraform init`.

```tf
terraform {
  required_providers {
    auth0 = {
      source  = "auth0/auth0"
      version = "0.17.1"
    }
  }
}

provider "auth0" {}
```

```sh
$ terraform init
```

**Terraform 0.12.x**

For older versions of Terraform, binaries are available at the
[releases](https://github.com/alexkappa/terraform-provider-auth0/releases) page. Download one that corresponds to your
operating system / architecture, and move it to the `~/.terraform.d/plugins/` directory. Finally, run terraform init.

```tf
provider "auth0" {}
```

```sh
$ terraform init
```

[[table of contents]](#table-of-contents)

## Documentation

See the [Auth0 Provider Documentation](https://registry.terraform.io/providers/auth0/auth0/latest/docs) for all the
available resources.

## Usage

You can find examples on usage under the [example](example) folder.

[[table of contents]](#table-of-contents)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

[[table of contents]](#table-of-contents)

## What is Auth0?

Auth0 helps you to:

- Add authentication with [multiple authentication sources](https://docs.auth0.com/identityproviders), either social like **Google, Facebook, Microsoft Account, LinkedIn, GitHub, Twitter, Box, Salesforce, amont others**, or enterprise identity systems like **Windows Azure AD, Google Apps, Active Directory, ADFS or any SAML Identity Provider**.
- Add authentication through more traditional **[username/password databases](https://docs.auth0.com/mysql-connection-tutorial)**.
- Add support for **[linking different user accounts](https://docs.auth0.com/link-accounts)** with the same user.
- Support for generating signed [Json Web Tokens](https://docs.auth0.com/jwt) to call your APIs and **flow the user identity** securely.
- Analytics of how, when and where users are logging in.
- Pull data from other sources and add it to the user profile, through [JavaScript rules](https://docs.auth0.com/rules).

[[table of contents]](#table-of-contents)

## Create a free Auth0 Account

1.  Go to [Auth0](https://auth0.com) and click "Try Auth0 for Free".
2.  Use Google, GitHub or Microsoft Account to login.

[[table of contents]](#table-of-contents)

## Issue Reporting

If you have found a bug or if you have a feature request, please report them at this repository issues section.
Please do not report security vulnerabilities on the public GitHub issue tracker.
The [Responsible Disclosure Program](https://auth0.com/whitehat) details the procedure for disclosing security issues.

[[table of contents]](#table-of-contents)

## Author

[Auth0](https://auth0.com/)

[[table of contents]](#table-of-contents)

## License

This project is licensed under the MPL-2.0 license. See the [LICENSE](LICENSE) file for more info.

[[table of contents]](#table-of-contents)
