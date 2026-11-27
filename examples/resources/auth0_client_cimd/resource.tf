resource "auth0_client_cimd" "minimal_client" {
  external_client_id = "https://mcp-agent1.example.com/oauth/metadata.json"
}

# Available in Early Access (EA). Referenced by `agent_id` below.
resource "auth0_agent" "my_agent" {
  name              = "My Agent"
  external_agent_id = "my-agent-001"
}

resource "auth0_client_cimd" "my_mcp_agent" {
  external_client_id         = "https://mcp-agent2.example.com/.well-known/client.json"
  external_client_id_version = 1
  description                = "MCP Agent - Production"
  app_type                   = "spa"
  oidc_conformant            = true

  # Reference the agent as `auth0_agent.<name>.id` so the client is unlinked before the agent is deleted.
  agent_id = auth0_agent.my_agent.id

  allowed_origins = ["https://mcp-agent2.example.com"]
  web_origins     = ["https://mcp-agent2.example.com"]

  grant_types = [
    "authorization_code",
    "refresh_token",
  ]

  organization_usage            = "require"
  organization_require_behavior = "post_login_prompt"

  client_metadata = {
    environment = "production"
  }

  jwt_configuration {
    lifetime_in_seconds = 300
    alg                 = "RS256"
  }

  refresh_token {
    rotation_type                = "rotating"
    expiration_type              = "expiring"
    token_lifetime               = 2592000
    idle_token_lifetime          = 1296000
    infinite_token_lifetime      = false
    infinite_idle_token_lifetime = false
    leeway                       = 0
  }

  # Show the Google FedCM prompt on login (EA only).
  fedcm_login {
    google {
      is_enabled = true
    }
  }

  # Allow the client to exchange ID-JAGs for access tokens, used for Cross App Access (EA only).
  identity_assertion_authorization_grant {
    active = true
  }

  # Limit the optional OIDC scopes the client may request.
  oidc_support {
    is_allowed     = true
    allowed_scopes = ["profile", "email"]
  }
}
