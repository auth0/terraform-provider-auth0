# This is an example of an Apple connection.

resource "auth0_connection" "apple" {
  name     = "Apple-Connection"
  strategy = "apple"

  options {
    client_id     = "<client-id>"
    client_secret = "<private-key>"
    team_id       = "<team-id>"
    key_id        = "<key-id>"
    scopes        = ["email", "name"]
  }
}
