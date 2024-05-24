# Example:
resource "auth0_resource_server" "resource_server" {
  name                                            = "test"
  identifier                                      = "test.example.com"
  signing_alg                                     = "RS256"
  token_lifetime                                  = 86400
  token_lifetime_for_web                          = 7200
  enforce_policies                                = true
  skip_consent_for_verifiable_first_party_clients = true
  allow_offline_access                            = false
  token_dialect                                   = "access_token"
}
resource "auth0_resource_server_scopes" "resource_server_scopes" {
  resource_server_identifier = auth0_resource_server.resource_server.identifier

  scopes {
    name = "access:store_contrib"
  }

  scopes {
    name = "access:store_read"
  }

  scopes {
    name = "access:serve_read"
  }
}

resource "auth0_role" "my_role" {
  name = "My Role"
}

resource "auth0_role_permission" "permission" {
  role_id                    = auth0_role.my_role.id
  resource_server_identifier = auth0_resource_server.resource_server.identifier
  permission                 = tolist(auth0_resource_server_scopes.resource_server_scopes.scopes)[0].name
}

