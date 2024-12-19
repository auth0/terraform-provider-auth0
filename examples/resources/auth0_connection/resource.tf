# This is an example of an Auth0 connection.

resource "auth0_connection" "my_connection" {
  name                 = "Example-Connection"
  is_domain_connection = true
  strategy             = "auth0"
  metadata = {
    key1 = "foo"
    key2 = "bar"
  }

  options {
    password_policy                = "excellent"
    brute_force_protection         = true
    strategy_version               = 2
    enabled_database_customization = true
    import_mode                    = false
    requires_username              = true
    disable_signup                 = false
    custom_scripts = {
      get_user = <<EOF
        function getByEmail(email, callback) {
          return callback(new Error("Whoops!"));
        }
      EOF
    }
    configuration = {
      foo = "bar"
      bar = "baz"
    }
    upstream_params = jsonencode({
      "screen_name" : {
        "alias" : "login_hint"
      }
    })

    password_history {
      enable = true
      size   = 3
    }

    password_no_personal_info {
      enable = true
    }

    password_dictionary {
      enable     = true
      dictionary = ["password", "admin", "1234"]
    }

    password_complexity_options {
      min_length = 12
    }

    validation {
      username {
        min = 10
        max = 40
      }
    }

    mfa {
      active                 = true
      return_enroll_settings = true
    }

    authentication_methods {
      passkey {
        enabled = true
      }
      password {
        enabled = true
      }
    }
  }
}
