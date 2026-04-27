resource "auth0_client_cimd" "minimal_client" {
  external_client_id = "https://mcp-agent1.example.com/oauth/metadata.json"
}

resource "auth0_client_cimd" "my_mcp_agent" {
  external_client_id         = "https://mcp-agent2.example.com/.well-known/client.json"
  external_client_id_version = 1
  description                = "MCP Agent - Production"
  app_type                   = "spa"
  oidc_conformant            = true

  allowed_origins = ["https://mcp-agent2.example.com"]
  web_origins     = ["https://mcp-agent2.example.com"]

  grant_types = [
    "authorization_code",
    "refresh_token",
  ]

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
}
