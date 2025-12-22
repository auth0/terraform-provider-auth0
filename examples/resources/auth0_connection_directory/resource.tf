resource "auth0_connection" "google_workspace" {
  name         = "google-workspace-connection"
  display_name = "Google Workspace"
  strategy     = "google-apps"

  options {
    client_id        = "your-google-client-id"
    client_secret    = "your-google-client-secret"
    domain           = "example.com"
    api_enable_users = true

  }
}

# Configure directory provisioning with default settings
resource "auth0_connection_directory" "default" {
  connection_id = auth0_connection.google_workspace.id
}

# Configure directory provisioning with custom mapping and auto-sync enabled
resource "auth0_connection_directory" "custom" {
  connection_id             = auth0_connection.google_workspace.id
  synchronize_automatically = true

  mapping {
    auth0 = "email"
    idp   = "primaryEmail"
  }

  mapping {
    auth0 = "family_name"
    idp   = "name.familyName"
  }

  mapping {
    auth0 = "given_name"
    idp   = "name.givenName"
  }

  mapping {
    auth0 = "external_id"
    idp   = "id"
  }
}

