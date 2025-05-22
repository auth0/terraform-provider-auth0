resource "auth0_resource_server" "resource_server" {
  name       = "Example Resource Server (Managed by Terraform)"
  identifier = "https://api.example.com"
}

resource "auth0_resource_server_scope" "read_posts" {
  resource_server_identifier = auth0_resource_server.resource_server.identifier
  scope                      = "read:posts"
}

resource "auth0_resource_server_scope" "write_posts" {
  resource_server_identifier = auth0_resource_server.resource_server.identifier
  scope                      = "write:posts"
}
