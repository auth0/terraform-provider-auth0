# An Auth0 Connection loaded using its name.
data "auth0_connection" "some-connection-by-name" {
  name = "Acceptance-Test-Connection-{{.testName}}"
}

# An Auth0 Connection loaded using its ID.
data "auth0_connection" "some-connection-by-id" {
  connection_id = "con_abcdefghkijklmnopqrstuvwxyz0123456789"
}
