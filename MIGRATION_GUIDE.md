# Migration Guide

## Upgrading from v0.46.0 → v0.47.0

There are deprecations in this update. Please ensure you read this guide thoroughly and prepare your potential 
automated workflows before upgrading.

### Deprecations

- [User Roles](#user-roles)

### User Roles

The `roles` field on the `auth0_user` resource will continue to be available for managing user roles. However, to ensure
a smooth transition when we eventually remove the capability to manage roles through this field, we recommend
proactively migrating to the newly introduced `auth0_user_roles` resource. This will help you stay prepared for future
changes.

<table>
<tr>
<th>Before (v0.46.0)</th>
<th>After (v0.47.0)</th>
</tr>
<tr>
<td>

```terraform
# Example:
resource "auth0_role" "admin" {
  name        = "admin"
  description = "Administrator"
}

resource "auth0_user" "user" {
  connection_name = "Username-Password-Authentication"
  username        = "unique_username"
  name            = "Firstname Lastname"
  email           = "test@test.com"
  password        = "passpass$12$12"
  roles           = [auth0_role.admin.id]
}
```

</td>
<td>

```terraform
# Example:
resource "auth0_role" "admin" {
  name        = "admin"
  description = "Administrator"
}

resource "auth0_user" "user" {
  connection_name = "Username-Password-Authentication"
  username        = "unique_username"
  name            = "Firstname Lastname"
  email           = "test@test.com"
  password        = "passpass$12$12"

  # Until we remove the ability to operate changes on
  # the roles field it is important to have this
  # block in the config, to avoid diffing issues.
  lifecycle {
    ignore_changes = [roles]
  }
}

resource auth0_user_roles user_roles {
  user_id = auth0_user.user.id
  roles = [auth0_role.admin.id]
}
```

</td>
</tr>
</table>
