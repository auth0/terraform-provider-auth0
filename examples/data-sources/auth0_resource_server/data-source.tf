# An Auth0 Resource Server loaded using its identifier.
data "auth0_resource_server" "some-resource-server-by-identifier" {
  identifier = "https://my-api.com/v1"
}

# An Auth0 Resource Server loaded using its ID.
data "auth0_resource_server" "some-resource-server-by-id" {
  resource_server_id = "abcdefghkijklmnopqrstuvwxyz0123456789"
}
