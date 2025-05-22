resource "auth0_resource_server" "my_api" {
  name       = "Example Resource Server (Managed by Terraform)"
  identifier = "https://api.example.com"
}

resource "auth0_resource_server_scopes" "my_api_scopes" {
  resource_server_identifier = auth0_resource_server.my_api.identifier

  scopes {
    name        = "create:appointments"
    description = "Ability to create appointments"
  }

  scopes {
    name        = "read:appointments"
    description = "Ability to read appointments"
  }
}
