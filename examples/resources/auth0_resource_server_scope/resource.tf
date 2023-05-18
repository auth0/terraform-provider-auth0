resource "auth0_resource_server" "resource_server" {
  name       = "Example Resource Server (Managed by Terraform)"
  identifier = "https://api.example.com"

  # Until we remove the ability to operate changes on
  # the scopes field it is important to have this
  # block in the config, to avoid diffing issues.
  lifecycle {
    ignore_changes = [scopes]
  }
}

resource "auth0_resource_server_scope" "read_posts" {
  resource_server_identifier = auth0_resource_server.resource_server.identifier
  scope                      = "read:posts"
}

resource "auth0_resource_server_scope" "write_posts" {
  resource_server_identifier = auth0_resource_server.resource_server.identifier
  scope                      = "write:posts"
}
