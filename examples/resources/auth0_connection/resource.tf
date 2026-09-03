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
    passkey_options {
      challenge_ui                   = "both"
      local_enrollment_enabled       = true
      progressive_enrollment_enabled = true
    }
  }
}

# The strategy's client secret can be set as a write-only argument so it is never persisted to
# Terraform state. It can be sourced from an ephemeral value (e.g. a secrets manager) and is
# mutually exclusive with `options.client_secret`. Bump `options_client_secret_wo_version` to
# rotate the secret.
#
# NOTE: Write-only arguments require Terraform 1.11 or later.
resource "auth0_connection" "my_connection_write_only_secret" {
  name     = "Example-Connection-Write-Only-Secret"
  strategy = "oidc"

  options_client_secret_wo         = var.connection_client_secret # can be an ephemeral value
  options_client_secret_wo_version = 1

  options {
    client_id              = "1234567"
    type                   = "back_channel"
    issuer                 = "https://www.paypalobjects.com"
    jwks_uri               = "https://api.paypal.com/v1/oauth2/certs"
    discovery_url          = "https://www.paypalobjects.com/.well-known/openid-configuration"
    token_endpoint         = "https://api.paypal.com/v1/oauth2/token"
    userinfo_endpoint      = "https://api.paypal.com/v1/oauth2/token/userinfo"
    authorization_endpoint = "https://www.paypal.com/signin/authorize"
    scopes                 = ["openid", "email"]
  }
}
