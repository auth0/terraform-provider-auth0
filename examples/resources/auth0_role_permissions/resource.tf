# Example:
resource "auth0_resource_server" "resource_server" {
  name       = "test"
  identifier = "test.example.com"
}
resource "auth0_resource_server_scopes" "resource_server_scopes" {
  resource_server_identifier = auth0_resource_server.resource_server.identifier

  scopes {
    name = "store:create"
  }
  scopes {
    name = "store:read"
  }
  scopes {
    name = "store:update"
  }
  scopes {
    name = "store:delete"
  }
}

resource "auth0_role" "my_role" {
  name = "My Role"
}

resource "auth0_role_permissions" "my_role_perms" {
  role_id = auth0_role.my_role.id

  dynamic "permissions" {
    for_each = auth0_resource_server_scopes.resource_server_scopes.scopes
    content {
      name                       = permissions.value.name
      resource_server_identifier = auth0_resource_server.resource_server.identifier
    }
  }
}
