resource "auth0_hook" "my_hook" {
  name       = "My Pre User Registration Hook"
  script     = <<EOF
    function (user, context, callback) {
      callback(null, { user });
    }
  EOF
  trigger_id = "pre-user-registration"
  enabled    = true
  secrets = {
    foo = "bar"
  }
  dependencies = {
    auth0 = "2.30.0"
  }
}
