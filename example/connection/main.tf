terraform {
  required_providers {
    auth0 = {
      source = "auth0/auth0"
    }
  }
}

provider "auth0" {}

resource "auth0_connection" "my_connection" {
  name     = "Example-Connection"
  strategy = "auth0"

  options {
    password_policy                = "excellent"
    brute_force_protection         = true
    enabled_database_customization = true
    custom_scripts = {
      get_user = <<EOF
function getByEmail (email, callback) {
  return callback(new Error("Whoops!"))
}
EOF
    }
    configuration = {
      foo = "bar"
      bar = "baz"
    }

    password_history {
      enable = true
      size   = 3
    }

    validation {
      username {
        min = 5
        max = 20
      }
    }
  }
}
