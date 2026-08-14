resource "auth0_connection" "my_connection" {
  name     = "My-Google-Workspace-Connection"
  strategy = "google-apps"
  options {
    client_id         = "..."
    client_secret     = "..."
    domain            = "example.com"
    tenant_domain     = "example.com"
    api_enable_users  = true
    api_enable_groups = true
  }
}

resource "auth0_connection_directory" "my_directory" {
  connection_id      = auth0_connection.my_connection.id
  synchronize_groups = "selected"
}

resource "auth0_connection_directory_synchronized_groups" "my_groups" {
  depends_on    = [auth0_connection_directory.my_directory]
  connection_id = auth0_connection.my_connection.id

  groups {
    id = "group1"
  }
  groups {
    id                   = "group2"
    name                 = "test"
    email                = "test@test.com"
    direct_members_count = 123
  }
}
