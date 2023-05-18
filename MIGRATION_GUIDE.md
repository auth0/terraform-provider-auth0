# Migration Guide

## Upgrading from v0.46.0 → v0.47.0

There are deprecations in this update. Please ensure you read this guide thoroughly and prepare your potential
automated workflows before upgrading.

### Deprecations

- [User Roles](#user-roles)
- [Role Permissions](#role-permissions)

### User Roles

The `roles` field on the `auth0_user` resource will continue to be available for managing user roles. However, to ensure
a smooth transition when we eventually remove the capability to manage roles through this field, we recommend
proactively migrating to the newly introduced `auth0_user_roles` or `auth0_user_role` resource. This will help you stay
prepared for future changes.

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

# Use the auth0_user_roles to manage a 1:many
# relationship between the user and its roles.
resource auth0_user_roles user_roles {
  user_id = auth0_user.user.id
  roles = [auth0_role.admin.id]
}

# Or the auth0_user_role to manage a 1:1
# relationship between the user and its role.
resource auth0_user_role user_roles {
  user_id = auth0_user.user.id
  roles = auth0_role.admin.id
}
```

</td>
</tr>
</table>

### Role Permissions

The `permissions` field on the `auth0_role` resource will continue to be available for managing role permissions. However, to ensure
a smooth transition when we eventually remove the capability to manage permissions through this field, we recommend
proactively migrating to the newly introduced `auth0_role_permission` resource. This will help you stay
prepared for future changes.

<table>
<tr>
<th>Before (v0.46.0)</th>
<th>After (v0.47.0)</th>
</tr>
<tr>
<td>

```terraform
resource auth0_resource_server api {
    name = "Example API"
    identifier = "https://api.travel0.com/"

    scopes {
        value = "read:posts"
        description = "Can read posts"
    }
    scopes {
        value = "write:posts"
        description = "Can write posts"
    }
}

resource auth0_role content_editor {
  name = "Content Editor"
  description = "Elevated roles for editing content"
  permissions {
    name = "read:posts"
    resource_server_identifier = auth0_resource_server.api.identifier
  }
  permissions {
    name = "write:posts"
    resource_server_identifier = auth0_resource_server.api.identifier
  }
}
```

</td>
<td>

```terraform
resource auth0_resource_server api {
    name = "Example API"
    identifier = "https://api.travel0.com/"

    scopes {
        value = "read:posts"
        description = "Can read posts"
    }
    scopes {
        value = "write:posts"
        description = "Can write posts"
    }
}

resource auth0_role content_editor {
  name = "Content Editor"
  description = "Elevated roles for editing content"

  # Until we remove the ability to operate changes on
  # the permissions field it is important to have this
  # block in the config, to avoid diffing issues.
	lifecycle {
		ignore_changes = [ permissions ]
	}
}

# Use the auth0_role_permission resource to manage a 1:1
# relationship between a role and its permissions.
resource "auth0_role_permission" "read_posts_permission" {
	role_id = auth0_role.content_editor.id
	resource_server_identifier = auth0_resource_server.api.identifier
	permission = "read:posts"
}

resource "auth0_role_permission" "write_posts_permission" {
	role_id = auth0_role.content_editor.id
	resource_server_identifier = auth0_resource_server.api.identifier
	permission = "write:posts"
}
```

</td>
</tr>
</table>
