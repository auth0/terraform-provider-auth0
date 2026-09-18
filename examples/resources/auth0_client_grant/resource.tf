# The following example grants a client the "create:foo" and "create:bar" permissions (scopes).

resource "auth0_client" "my_client" {
  name = "Example Application - Client Grant (Managed by Terraform)"
}

resource "auth0_resource_server" "my_resource_server" {
  name       = "Example Resource Server - Client Grant (Managed by Terraform)"
  identifier = "https://api.example.com/client-grant"

  authorization_details {
    type = "payment"
  }
  authorization_details {
    type = "shipping"
  }
  subject_type_authorization {
    user {
      policy = "allow_all"
    }
    client {
      policy = "require_client_grant"
    }
  }
}


resource "auth0_resource_server_scopes" "my_scopes" {
  depends_on = [auth0_resource_server.my_resource_server]

  resource_server_identifier = auth0_resource_server.my_resource_server.identifier

  scopes {
    name        = "read:foo"
    description = "Can read Foo"
  }

  scopes {
    name        = "create:foo"
    description = "Can create Foo"
  }
}

resource "auth0_client_grant" "my_client_grant" {
  client_id                   = auth0_client.my_client.id
  audience                    = auth0_resource_server.my_resource_server.identifier
  scopes                      = ["create:foo", "read:foo"]
  subject_type                = "user"
  authorization_details_types = ["payment", "shipping"]
  allow_all_scopes            = false
}

resource "auth0_client_grant" "default_3p_grant" {
  default_for = "third_party_clients"
  audience    = auth0_resource_server.my_resource_server.identifier
  scopes      = ["read:foo"]
}

# The following example grants a client access to anonymous access tokens for an API whose
# anonymous_user policy is set to "require_client_grant". The subject_type "anonymous_user" is
# only accepted at creation time and cannot be combined with organization_usage,
# allow_any_organization, authorization_details_types, or default_for.
# Anonymous Sessions is an Early Access feature.
resource "auth0_client_grant" "my_anonymous_grant" {
  client_id    = auth0_client.my_client.id
  audience     = auth0_resource_server.my_resource_server.identifier
  scopes       = ["read:foo"]
  subject_type = "anonymous_user"
}
