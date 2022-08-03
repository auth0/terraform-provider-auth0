resource "auth0_action" "my_action" {
  name    = format("Test Action %s", timestamp())
  runtime = "node16"
  deploy  = true
  code    = <<-EOT
	/**
	 * Handler that will be called during the execution of a PostLogin flow.
	 *
	 * @param {Event} event - Details about the user and the context in which they are logging in.
	 * @param {PostLoginAPI} api - Interface whose methods can be used to change the behavior of the login.
	 */
	 exports.onExecutePostLogin = async (event, api) => {
		 console.log(event)
	 };
	EOT

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
