provider "auth0" {}

resource "auth0_action" "my_action" {
  name    = format("Test Action %s", timestamp())
  runtime = "node16"
  code    = <<-EOT
	exports.onContinuePostLogin = async (event, api) => {
		console.log(event)
	};"
	EOT
  deploy  = true

  supported_triggers {
    id      = "post-login"
    version = "v3"
  }

  dependencies {
    name    = "lodash"
    version = "latest"
  }

  dependencies {
    name    = "request"
    version = "latest"
  }

  secrets {
    name  = "FOO"
    value = "Foo"
  }

  secrets {
    name  = "BAR"
    value = "Bar"
  }
}
