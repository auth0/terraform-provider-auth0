# This is an example of a Facebook connection.

resource "auth0_connection" "facebook" {
  name     = "Facebook-Connection"
  strategy = "facebook"

  options {
    client_id     = "<client-id>"
    client_secret = "<client-secret>"
    scopes = [
      "public_profile",
      "email",
      "groups_access_member_info",
      "user_birthday"
    ]
    set_user_root_attributes = "on_each_login"
    non_persistent_attrs = ["ethnicity","gender"]
  }
}
