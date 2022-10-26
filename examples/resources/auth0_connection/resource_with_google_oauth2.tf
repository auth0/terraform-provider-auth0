# This is an example of a Google OAuth2 connection.

resource "auth0_connection" "google_oauth2" {
  name     = "Google-OAuth2-Connection"
  strategy = "google-oauth2"

  options {
    client_id                = "<client-id>"
    client_secret            = "<client-secret>"
    allowed_audiences        = ["example.com", "api.example.com"]
    scopes                   = ["email", "profile", "gmail", "youtube"]
    set_user_root_attributes = "on_each_login"
    non_persistent_attrs = ["ethnicity","gender"]
  }
}
