---
layout: "auth0"
page_title: "Auth0: auth0_organization_member"
description: |-
  This resource is used to manage the assignment of members and their roles within an organization.
---

# auth0_organization_member

This resource is used to manage the assignment of members and their roles within an organization.

## Example Usage

```hcl
resource auth0_organization_member acme_admin {
  organization_id = auth0_organization.acme.id
  user_id = auth0_user.acme_user.id
  roles = [ auth0_role.admin.id ] 
}
```

## Argument Reference

The following arguments are supported:

* `organization_id` - (Required) The ID of the organization
* `user_id` – (Required) The user ID of the member
* `roles` – (Optional) Set(string). List of role IDs to assign to member.


## Import

As this is not a resource identifiable by an ID within the Auth0 Management API, organization_connection can be imported
using a random string. We recommend [Version 4 UUID](https://www.uuidgenerator.net/version4) e.g.

```
$ terraform import auth0_organization_member.acme_admin 11f4a21b-011a-312d-9217-e291caca36c5
```
