resource "auth0_token_exchange_profile" "my_token_exchange_profile" {
  name               = "token-exchange-prof"
  subject_token_type = "https://acme.com/cis-token"
  action_id          = auth0_action.my_action.id
  type               = "custom_authentication"
}


# Below action is created with custom-token-exchange as supported_triggers
# This action is then linked using the action_id param to the token-exchange profile
resource "auth0_action" "my_action" {
  name   = "TokenExchange-Action"
  code   = <<-EOT
		exports.onExecuteCustomTokenExchange = async (event, api) => {
			console.log("foo")
		};"
		EOT
  deploy = true
  supported_triggers {
    id      = "custom-token-exchange"
    version = "v1"
  }
}
