resource "auth0_resource_server" "resource_server" {
  name       = "Example Resource Server (Managed by Terraform)"
  identifier = "https://api.example.com"
  scopes {
    value       = "create:foo"
    description = "Create foos"
  }

  scopes {
    value       = "read:foo"
    description = "Read foos"
  }
}

resource "auth0_user" "user" {
  connection_name = "Username-Password-Authentication"
  user_id         = "12345"
  username        = "unique_username"
  name            = "Firstname Lastname"
  nickname        = "some.nickname"
  email           = "test@test.com"
  email_verified  = true
  password        = "passpass$12$12"
  picture         = "https://www.example.com/a-valid-picture-url.jpg"
}

resource "auth0_user_permissions" "all_user_permissions" {
  user_id = auth0_user.user.id

  permissions {
    name                       = tolist(auth0_resource_server.resource_server.scopes)[0]
    resource_server_identifier = auth0_resource_server.resource_server.identifier
  }

  permissions {
    name                       = tolist(auth0_resource_server.resource_server.scopes)[1]
    resource_server_identifier = auth0_resource_server.resource_server.identifier
  }
}
