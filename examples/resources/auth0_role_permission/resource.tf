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

locals {
  scopesList = [
    for scope in auth0_resource_server_scopes.resource_server_scopes.scopes : scope.name
  ]
}

resource "auth0_role_permission" "my_role_perm" {
  for_each = toset(local.scopesList)

  role_id                    = auth0_role.my_role.id
  resource_server_identifier = auth0_resource_server.resource_server.identifier
  permission                 = each.value
}
