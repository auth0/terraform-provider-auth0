# This is an example of an OAuth2 connection.

resource "auth0_connection" "oauth2" {
  name     = "OAuth2-Connection"
  strategy = "oauth2"

  options {
    client_id              = "<client-id>"
    client_secret          = "<client-secret>"
    scopes                 = ["basic_profile", "profile", "email"]
    token_endpoint         = "https://auth.example.com/oauth2/token"
    authorization_endpoint = "https://auth.example.com/oauth2/authorize"
    pkce_enabled           = true
    icon_url = "https://auth.example.com/assets/logo.png"
    scripts = {
      fetchUserProfile = <<EOF
        function fetchUserProfile(accessToken, context, callback) {
          return callback(new Error("Whoops!"));
        }
      EOF
    }
    set_user_root_attributes = "on_each_login"
    non_persistent_attrs = ["ethnicity","gender"]
  }
}
