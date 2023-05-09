resource "auth0_connection" "my_conn" {
  name     = "My-Auth0-Connection"
  strategy = "auth0"
}

resource "auth0_client" "my_first_client" {
  name = "My-First-Auth0-Client"
}

resource "auth0_client" "my_second_client" {
  name = "My-Second-Auth0-Client"
}

# One connection to many clients association.
# To prevent issues, avoid using this resource together with the `auth0_connection_client` resource.
resource "auth0_connection_clients" "my_conn_clients_assoc" {
  connection_id = auth0_connection.my_conn.id
  enabled_clients = [
    auth0_client.my_first_client.id,
    auth0_client.my_second_client.id
  ]
}
