# This is an example of an LinkedIn connection.

resource "auth0_connection" "linkedin" {
  name     = "Linkedin-Connection"
  strategy = "linkedin"

  options {
    client_id        = "<client-id>"
    client_secret    = "<client-secret>"
    strategy_version = 2
    scopes           = ["basic_profile", "profile", "email"]
  }
}
