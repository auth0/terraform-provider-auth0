---
layout: "auth0"
page_title: "Data Source: auth0_config"
description: |-
Use this data source to access the configuration of the Auth0 provider.
---

# Data Source: auth0_config

Use this data source to access the configuration of the Auth0 provider.

## Example Usage

```hcl
data "auth0_config" "current" {}
```

## Argument Reference

No arguments accepted.

## Attribute Reference

* `domain` - String. Your Auth0 domain name.
* `management_api_identifier` - String. The identifier value of the built-in Management API resource server, which can be used as an audience when configuring client grants.
