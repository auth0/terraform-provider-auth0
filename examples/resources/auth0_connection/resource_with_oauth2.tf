# This is an example of an OAuth2 connection.

resource "auth0_connection" "oauth2" {
  name     = "OAuth2-Connection"
  strategy = "oauth2"

  options {
    client_id              = "<client-id>"
    client_secret          = "<client-secret>"
    token_endpoint         = "https://auth.example.com/oauth2/token"
    authorization_endpoint = "https://auth.example.com/oauth2/authorize"
    pkce_enabled           = true
    scripts = {
      fetchUserProfile = <<EOF
        function function(accessToken, ctx, cb) {
          return callback(new Error("Whoops!"))
        }
      EOF
    }
  }
}
