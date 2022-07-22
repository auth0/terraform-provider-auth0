---
layout: "auth0"
page_title: "Auth0: auth0_organization_connection"
description: |-
  With this resource, you can manage enabled connections on an organization.
---

# auth0_organization_connection

With this resource, you can manage enabled connections on an organization.

## Example Usage

```hcl
resource "auth0_organization_connection" "example" {
  organization_id = "org_XXXXXXXXXX"
  connection_id = "con_XXXXXXXXXX"
  assign_membership_on_login = true
}
```

## Argument Reference

The following arguments are supported:

* `assign_membership_on_login` - (Optional) When true, all users that log in with this connection will be automatically granted membership in the organization. When false, users must be granted membership in the organization before logging in with this connection.
* `connection_id` - (Required) The ID of the connection to enable for the organization.
* `organization_id` - (Required) The ID of the organization to enable the connection for.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are
exported:

* `name` - The name of the enabled connection.
* `strategy` - The strategy of the enabled connection.

## Import

As this is not a resource identifiable by an ID within the Auth0 Management API, organization_connection can be imported
using a random string. We recommend [Version 4 UUID](https://www.uuidgenerator.net/version4) e.g.

```
$ terraform import auth0_organization_connection.example 11f4a21b-011a-312d-9217-e291caca36c4
```
