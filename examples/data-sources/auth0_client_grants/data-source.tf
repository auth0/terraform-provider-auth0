resource "auth0_client" "my_client" {
  name = "Example Application (Managed by Terraform)"
}

resource "auth0_resource_server" "my_resource_server" {
  name       = "Example Resource Server (Managed by Terraform)"
  identifier = "https://api.example.com/client-grant"
  authorization_details { type = "payment" }
  subject_type_authorization {
    user { policy = "allow_all" }
  }
}

resource "auth0_resource_server_scopes" "my_scopes" {
  depends_on                 = [auth0_resource_server.my_resource_server]
  resource_server_identifier = auth0_resource_server.my_resource_server.identifier
  scopes { name = "create:foo" }
}

resource "auth0_client_grant" "my_client_grant" {
  client_id                   = auth0_client.my_client.id
  audience                    = auth0_resource_server.my_resource_server.identifier
  authorization_details_types = ["payment"]
  subject_type                = "user"
  allow_all_scopes            = true
}

data "auth0_client_grants" "filter_by_client_id" {
  depends_on = [auth0_client_grant.my_client_grant]
  client_id  = auth0_client.my_client.id
}

data "auth0_client_grants" "filter_by_audience" {
  depends_on = [auth0_client_grant.my_client_grant]
  audience   = auth0_resource_server.my_resource_server.identifier
}

data "auth0_client_grants" "filter_by_client_id_and_audience" {
  depends_on = [auth0_client_grant.my_client_grant]
  client_id  = auth0_client.my_client.id
  audience   = auth0_resource_server.my_resource_server.identifier
}