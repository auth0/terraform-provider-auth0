# This is an example of an GitHub connection.

resource "auth0_connection" "github" {
  name     = "GitHub-Connection"
  strategy = "github"

  options {
    client_id                = "<client-id>"
    client_secret            = "<client-secret>"
    scopes                   = ["email", "profile", "public_repo", "repo"]
    set_user_root_attributes = "on_each_login"
    non_persistent_attrs     = ["ethnicity", "gender"]
  }
}
