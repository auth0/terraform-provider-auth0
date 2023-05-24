resource "auth0_resource_server" "my_api" {
  name       = "Example Resource Server (Managed by Terraform)"
  identifier = "https://api.example.com"

  # Until we remove the ability to operate changes on
  # the scopes field it is important to have this
  # block in the config, to avoid diffing issues.
  lifecycle {
    ignore_changes = [scopes]
  }
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
