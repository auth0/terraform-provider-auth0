# Migration Guide

## Upgrading from v0.48.0 → v0.49.0

There are deprecations in this update. Please ensure you read this guide thoroughly and prepare your potential
automated workflows before upgrading.

### Deprecations

- [Trigger Binding Renaming](#trigger-binding-renaming)

#### Trigger Binding Renaming

The `auth0_trigger_binding` resource has been renamed to `auth0_trigger_actions` for clarity and consistency with the `auth0_trigger_action` (1:1) resource. To migrate, simply rename the resource from `auth0_trigger_binding` to `auth0_trigger_actions`.

<table>
<tr>
<th>Before (v0.48.0)</th>
<th>After (v0.49.0)</th>
</tr>
<tr>
<td>

```terraform
resource auth0_trigger_binding login_flow {
	trigger = "post-login"

	actions {
		id = auth0_action.my_action.id
		display_name = auth0_action.my_action.name
	}
}
```

</td>
<td>

```terraform
resource auth0_trigger_actions login_flow {
	trigger = "post-login"

	actions {
		id = auth0_action.my_action.id
		display_name = auth0_action.my_action.name
	}
}
```

</td>
</tr>
</table>

## Upgrading from v0.47.0 → v0.48.0

There are deprecations in this update. Please ensure you read this guide thoroughly and prepare your potential
automated workflows before upgrading.

### Deprecations

- [Client Authentication Method](#client-authentication-method)
- [Resource Server Scopes](#resource-server-scopes)

#### Client Authentication Method

The `token_endpoint_auth_method` field on the `auth0_client` resource will continue to be available for managing the
client's authentication method. However, to ensure a smooth transition when we eventually remove the capability to
manage the authentication method through this field, we recommend proactively migrating to the newly introduced
`auth0_client_credentials` resource as this will also give you the possibility of managing the client secret.
This will help you stay prepared for future changes.

<table>
<tr>
<th>Before (v0.47.0)</th>
<th>After (v0.48.0)</th>
</tr>
<tr>
<td>

```terraform
# Example:
resource "auth0_client" "my_client" {
  name = "My Client"

  token_endpoint_auth_method = "client_secret_post"
}
```

</td>
<td>

```terraform
# Example:
resource "auth0_client" "my_client" {
  name = "My Client"
}

resource "auth0_client_credentials" "test" {
  client_id = auth0_client.my_client.id

  authentication_method = "client_secret_post"
}
```

</td>
</tr>
</table>

#### Resource Server Scopes

The `scopes` field on the `auth0_resource_server` resource will continue to be available for managing resource server scopes. However, to ensure a smooth transition when we eventually remove the capability to manage scopes through this field, we recommend proactively migrating to the newly introduced `auth0_resource_server_scope` resource. This will help you stay prepared for future changes.

<table>
<tr>
<th>Before (v0.47.0)</th>
<th>After (v0.48.0)</th>
</tr>
<tr>
<td>

```terraform
resource auth0_resource_server api {
  name       = "Example API"
  identifier = "https://api.travel0.com/"

  scopes {
    value       = "read:posts"
    description = "Can read posts"
  }

  scopes {
    value       = "write:posts"
    description = "Can write posts"
  }
}
```

</td>
<td>

```terraform
resource auth0_resource_server api {
  name       = "Example API"
  identifier = "https://api.travel0.com/"

  # Until we remove the ability to operate changes on
  # the scopes field it is important to have this
  # block in the config, to avoid diffing issues.
  lifecycle {
    ignore_changes = [scopes]
  }
}

# Use the auth0_resource_server_scopes to manage a 1:many
# relationship between the resource server and its scopes.
resource "auth0_resource_server_scopes" "my_api_scopes" {
  resource_server_identifier = auth0_resource_server.my_api.identifier

  scopes {
    name        = "read:posts"
    description = "Can read posts"
  }

  scopes {
    name        = "write:posts"
    description = "Can write posts"
  }
}

# Use the auth0_resource_server_scope to manage a 1:1
# relationship between the resource server and its scopes.
resource auth0_resource_server_scope read_posts {
  resource_server_identifier = auth0_resource_server.api.identifier
  scope       = "read:posts"
  description = "Can read posts"
}

resource auth0_resource_server_scope write_posts {
  resource_server_identifier = auth0_resource_server.api.identifier
  scope       = "write:posts"
  description = "Can write posts"
}
```

</td>
</tr>
</table>

## Upgrading from v0.46.0 → v0.47.0

There are deprecations in this update. Please ensure you read this guide thoroughly and prepare your potential
automated workflows before upgrading.

### Deprecations

- [User Roles](#user-roles)
- [Role Permissions](#role-permissions)

#### User Roles

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

resource "auth0_role" "owner" {
  name        = "owner"
  description = "Owner"
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
  roles   = [ auth0_role.admin.id, auth0_role.owner.id ]
}

# Use the auth0_user_role to manage a 1:1
# relationship between the user and its role.
resource auth0_user_role user_admin {
  user_id = auth0_user.user.id
  role_id = auth0_role.admin.id
}

resource auth0_user_role user_owner {
  user_id = auth0_user.user.id
  role_id = auth0_role.owner.id
}
```

</td>
</tr>
</table>

#### Role Permissions

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
  name       = "Example API"
  identifier = "https://api.travel0.com/"

  scopes {
    value       = "read:posts"
    description = "Can read posts"
  }

  scopes {
    value       = "write:posts"
    description = "Can write posts"
  }
}

resource auth0_role content_editor {
  name        = "Content Editor"
  description = "Elevated roles for editing content"

  permissions {
    name                       = "read:posts"
    resource_server_identifier = auth0_resource_server.api.identifier
  }

  permissions {
    name                       = "write:posts"
    resource_server_identifier = auth0_resource_server.api.identifier
  }
}
```

</td>
<td>

```terraform
resource auth0_resource_server api {
  name       = "Example API"
  identifier = "https://api.travel0.com/"

  scopes {
    value       = "read:posts"
    description = "Can read posts"
  }

  scopes {
    value       = "write:posts"
    description = "Can write posts"
  }
}

resource auth0_role content_editor {
  name        = "Content Editor"
  description = "Elevated roles for editing content"

  # Until we remove the ability to operate changes on
  # the permissions field it is important to have this
  # block in the config, to avoid diffing issues.
	lifecycle {
		ignore_changes = [ permissions ]
	}
}

# Use the auth0_role_permissions to manage a 1:many
# relationship between the role and its permissions.
resource "auth0_role_permissions" "editor_permissions" {
	role_id = auth0_role.content_editor.id

  permissions  {
    resource_server_identifier = auth0_resource_server.api.identifier
    name                       = "read:posts"
  }

  permissions  {
    resource_server_identifier = auth0_resource_server.api.identifier
    name                       = "write:posts"
  }
}

# Use the auth0_role_permission resource to manage a 1:1
# relationship between a role and its permissions.
resource "auth0_role_permission" "read_posts_permission" {
	role_id                    = auth0_role.content_editor.id
	resource_server_identifier = auth0_resource_server.api.identifier
	permission                 = "read:posts"
}

resource "auth0_role_permission" "write_posts_permission" {
	role_id                    = auth0_role.content_editor.id
	resource_server_identifier = auth0_resource_server.api.identifier
	permission                 = "write:posts"
}
```

</td>
</tr>
</table>
