resource "auth0_connection" "my_connection-1" {
  name     = "My Connection 1"
  strategy = "auth0"
}

resource "auth0_connection" "my_connection-2" {
  name     = "My Connection 2"
  strategy = "auth0"
}

resource "auth0_organization" "my_organization" {
  name         = "my-organization"
  display_name = "My Organization"
}

resource "auth0_organization_connections" "one-to-many" {
  organization_id = auth0_organization.my_organization.id

  enabled_connections {
    connection_id              = auth0_connection.my_connection-1.id
    assign_membership_on_login = true
  }

  enabled_connections {
    connection_id              = auth0_connection.my_connection-2.id
    assign_membership_on_login = true
  }
}
