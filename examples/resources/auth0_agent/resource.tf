# Available in Early Access (EA). Requires the tenant to be enabled for agents.
resource "auth0_agent" "my_agent" {
  name              = "My Agent"
  external_agent_id = "my-agent-001"
  metadata = jsonencode({
    team = "platform"
  })
}

# Link the agent to an application. Reference the agent as `auth0_agent.<name>.id`
# so the client is unlinked before the agent is deleted.
resource "auth0_client" "my_agent_client" {
  name     = "My Agent Client"
  app_type = "non_interactive"
  agent_id = auth0_agent.my_agent.id
}
