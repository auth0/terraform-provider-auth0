# Migration Guide

## Upgrading from v0.49.0 → v0.50.0

There are deprecations in this update. Please ensure you read this guide thoroughly and prepare your potential
automated workflows before upgrading.

### Deprecations

- [Global Client](#global-client)
- [Tenant Pages](#tenant-pages)
- [Tenant Universal Login](#tenant-universal-login)

#### Global Client

The `auth0_global_client` resource and data source were introduced primarily to allow managing the `custom_login_page`
and `custom_login_page_on` attributes in order to manage the custom login page of a tenant. These are now deprecated in
favour of the `auth0_pages` resource.

To ensure a smooth transition when we eventually remove the capability to manage the custom 
login page through the `auth0_global_client`, we recommend proactively migrating to the `auth0_pages` resource. 
This will help you stay prepared for future changes.

<table>
<tr>
<th>Before (v0.49.0)</th>
<th>After (v0.50.0)</th>
</tr>
<tr>
<td>

```terraform
resource "auth0_global_client" "global" {
  custom_login_page_on = true
  custom_login_page    = "<html>My Custom Login Page</html>"
}
```

</td>
<td>

```terraform
resource "auth0_pages" "my_pages" {
  login {
    enabled = true
    html    = "<html><body>My Custom Login Page</body></html>"
  }
}
```

</td>
</tr>
</table>

#### Tenant Pages

The `change_password`, `guardian_mfa_page` and `error_page` attributes on the `auth0_tenant` have been deprecated in
favour of managing them with the `auth0_pages` resource. 

To ensure a smooth transition when we eventually remove the capability to manage these custom Auth0 pages through the
`auth0_tenant` resource, we recommend proactively migrating to the `auth0_pages` resource. This will help you stay
prepared for future changes.

<table>
<tr>
<th>Before (v0.49.0)</th>
<th>After (v0.50.0)</th>
</tr>
<tr>
<td>

```terraform
resource "auth0_tenant" "my_tenant" {
  change_password {
    enabled = true
    html    = "<html>My Custom Reset Password Page</html>"
  }

  guardian_mfa_page {
    enabled = true
    html    = "<html>My Custom MFA Page</html>"
  }

  error_page {
    html          = "<html>My Custom Error Page</html>"
    show_log_link = true
    url           = "https://example.com/errors"
  }
}
```

</td>
<td>

```terraform
resource "auth0_pages" "my_pages" {
  change_password {
    enabled = true
    html    = "<html><body>My Custom Reset Password Page</body></html>"
  }

  guardian_mfa {
    enabled = true
    html    = "<html><body>My Custom MFA Page</body></html>"
  }

  error {
    show_log_link = true
    html          = "<html><body>My Custom Error Page</body></html>"
    url           = "https://example.com"
  }
}
```

</td>
</tr>
</table>

#### Tenant Universal Login

The `universal_login` settings on the `auth0_tenant` have been deprecated in favour of managing them through the `auth0_branding` resource.

To ensure a smooth transition when we eventually remove the capability to manage these settings through the
`auth0_tenant` resource, we recommend proactively migrating to the `auth0_branding` resource. This will help you stay
prepared for future changes.

<table>
<tr>
<th>Before (v0.49.0)</th>
<th>After (v0.50.0)</th>
</tr>
<tr>
<td>

```terraform
resource "auth0_tenant" "my_tenant" {
  universal_login {
    colors {
      primary         = "#0059d6"
      page_background = "#000000"
    }
  }
}
```

</td>
<td>

```terraform
resource "auth0_branding" "my_branding" {
  colors {
    primary         = "#0059d6"
    page_background = "#000000"
  } 
}
```

</td>
</tr>
</table>

## Upgrading from v0.48.0 → v0.49.0

There are deprecations in this update. Please ensure you read this guide thoroughly and prepare your potential
automated workflows before upgrading.

### Deprecations

- [Trigger Binding Renaming](#trigger-binding-renaming)
- [Organization Member Roles](#organization-member-roles)

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

#### Organization Member Roles

The `roles` field on the `auth0_organization_member` resource will continue to be available for managing an organization
member's roles. However, to ensure a smooth transition when we eventually remove the capability to manage roles through
this field, we recommend proactively migrating to the newly introduced `auth0_organization_member_roles` resource.
This will help you stay prepared for future changes.

<table>
<tr>
<th>Before (v0.48.0)</th>
<th>After (v0.49.0)</th>
</tr>
<tr>
<td>

```terraform
resource "auth0_role" "reader" {
  name = "Reader"
}

resource "auth0_role" "writer" {
  name = "Writer"
}

resource "auth0_user" "user" {
  email           = "test-user@auth0.com"
  connection_name = "Username-Password-Authentication"
  email_verified  = true
  password        = "MyPass123$"
}

resource "auth0_organization" "my_org" {
  name         = "some-org"
  display_name = "Some Org"
}

resource "auth0_organization_member" "my_org_member" {
  organization_id = auth0_organization.my_org.id
  user_id         = auth0_user.user.id
  roles           = [ auth0_role.reader.id, auth0_role.writer.id ]
}
```

</td>
<td>

```terraform
resource "auth0_role" "reader" {
  name = "Reader"
}

resource "auth0_role" "writer" {
  name = "Writer"
}

resource "auth0_user" "user" {
  connection_name = "Username-Password-Authentication"
  email           = "test-user@auth0.com"
  password        = "MyPass123$"
}

resource "auth0_organization" "my_org" {
  name         = "some-org"
  display_name = "Some Org"
}

resource "auth0_organization_member" "my_org_member" {
  organization_id = auth0_organization.my_org.id
  user_id         = auth0_user.user.id

  # Until we remove the ability to operate changes on
  # the roles field it is important to have this
  # block in the config, to avoid diffing issues.
  lifecycle {
    ignore_changes = [ roles ]
  }
}

# Use the auth0_organization_member_roles to manage a 1:many
# relationship between the organization member and its roles.
resource "auth0_organization_member_roles" "my_org_member_roles" {
  organization_id = auth0_organization.my_org.id
  user_id         = auth0_user.user.id
  roles           = [ auth0_role.reader.id, auth0_role.writer.id ]
}

# Use the auth0_organization_member_role to manage a 1:1
# relationship between the organization member and its roles.
resource "auth0_organization_member_role" "role1" {
  organization_id = auth0_organization.my_org.id
  user_id         = auth0_user.user.id
  role_id         = auth0_role.reader.id
}

resource "auth0_organization_member_role" "role2" {
  organization_id = auth0_organization.my_org.id
  user_id         = auth0_user.user.id
  role_id         = auth0_role.writer.id
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
