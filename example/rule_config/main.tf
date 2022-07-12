terraform {
  required_providers {
    auth0 = {
      source = "auth0/auth0"
    }
  }
}

provider "auth0" {}

resource "auth0_rule" "my_rule" {
  name    = "empty-rule"
  script  = <<EOF
function (user, context, callback) {
  callback(null, user, context);
}
EOF
  enabled = true
}

resource "auth0_rule_config" "my_rule_config" {
  key   = "foo"
  value = "bar"
}
