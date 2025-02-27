# Example:
resource "auth0_flow" "my_flow" {
  actions = jsonencode([{
    action        = "UPDATE_USER"
    alias         = "user meta data"
    allow_failure = false
    id            = "update_user_PmSa"
    mask_output   = false
    params = {
      changes = {
        user_metadata = {
          full_name = "{{fields.full_name}}"
        }
      }
      connection_id = "<vault_connection_id>" #  Altenative ways: (connection_id = auth0_flow_vault_connection.my_connection.id) or using terraform variables
      user_id       = "{{context.user.user_id}}"
    }
    type = "AUTH0"
  }])
  name = "Flow KYC update data"
}

