resource "auth0_connection" "my_conn" {
  name     = "My-Auth0-Connection"
  strategy = "auth0"
  # Avoid using the enabled_clients = [...],
  # if using the auth0_connection_client resource.
}

resource "auth0_client" "my_client" {
  name = "My-Auth0-Client"
}

resource "auth0_connection_client" "my_conn_client_assoc" {
  connection_id = auth0_connection.my_conn.id
  client_id     = auth0_client.my_client.id
}
