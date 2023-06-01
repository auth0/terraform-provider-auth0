resource "auth0_action" "login_alert" {
  name   = "Alert after login"
  code   = <<-EOT
    exports.onContinuePostLogin = async (event, api) => {
      console.log("foo");
    };"
	EOT
  deploy = true

  supported_triggers {
    id      = "post-login"
    version = "v3"
  }
}

resource "auth0_trigger_action" "post_login_alert_action" {
  trigger   = "post-login"
  action_id = auth0_action.login_alert.id
}
