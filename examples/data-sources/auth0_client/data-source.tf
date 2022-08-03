# An Auth0 Client loaded using its name.
data "auth0_client" "some-client-by-name" {
  name = "Name of my Application"
}

# An Auth0 Client loaded using its ID.
data "auth0_client" "some-client-by-id" {
  client_id = "abcdefghkijklmnopqrstuvwxyz0123456789"
}
