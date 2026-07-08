resource "auth0_action" "my_action" {
  name    = format("Test Action %s", timestamp())
  runtime = "node22"
  deploy  = true
  code    = <<-EOT
  /**
   * Handler that will be called during the execution of a PostLogin flow.
   *
   * @param {Event} event - Details about the user and the context in which they are logging in.
   * @param {PostLoginAPI} api - Interface whose methods can be used to change the behavior of the login.
   */
   exports.onExecutePostLogin = async (event, api) => {
     console.log(event);
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

# Creates an action with write-only secrets (recommended for security).
# Secret values are never stored in Terraform state. They can be supplied via
# regular sensitive variables or via ephemeral variables/resources.
variable "action_api_key" {
  description = "API key passed to the post-login action."
  type        = string
  sensitive   = true
}

resource "auth0_action" "my_secure_action" {
  name    = format("Secure Action %s", timestamp())
  runtime = "node22"
  deploy  = true
  code    = <<-EOT
   exports.onExecutePostLogin = async (event, api) => {
     console.log(event);
   };
  EOT

  supported_triggers {
    id      = "post-login"
    version = "v3"
  }

  secrets_wo {
    name  = "API_KEY"
    value = var.action_api_key
  }

  secrets_wo_version = 1
}
