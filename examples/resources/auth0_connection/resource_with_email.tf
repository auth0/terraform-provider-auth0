# This is an example of an Email connection.

resource "auth0_connection" "passwordless_email" {
  strategy = "email"
  name     = "email"

  options {
    from                     = "{{ application.name }} \u003croot@auth0.com\u003e"
    subject                  = "Welcome to {{ application.name }}"
    syntax                   = "liquid"
    template                 = "<html>This is the body of the email</html>"
    disable_signup           = false
    brute_force_protection   = true
    set_user_root_attributes = []
    non_persistent_attrs     = []
    auth_params = {
      scope         = "openid email profile offline_access"
      response_type = "code"
    }

    totp {
      time_step = 300
      length    = 6
    }
  }
}
