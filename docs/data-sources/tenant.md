---
layout: "auth0"
page_title: "Data Source: auth0_tenant"
description: |-
Use this data source to access information about the tenant this provider is configured to access.
---

# Data Source: auth0_tenant

Use this data source to access information about the tenant this provider is configured to access.

## Example Usage

```hcl
data "auth0_tenant" "current" {}
```

## Argument Reference

No arguments accepted.

## Attribute Reference

* `domain` - String. Your Auth0 domain name.
* `management_api_identifier` - String. The identifier value of the built-in Management API resource server, which can be used as an audience when configuring client grants.
