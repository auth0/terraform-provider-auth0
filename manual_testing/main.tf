terraform {
    required_version = ">= 1.5.0"
    required_providers {
        auth0 = {
            source  = "auth0/auth0"
            version = "1.7.3"
        }
    }
}


# Custom email provider resource:

resource "auth0_email_provider" "custom_email_provider" {
    depends_on = [auth0_action.custom-email-provider-1]
default_from_address = "account-update@notifications.grainger.com"
enabled = true
name = "custom"
credentials {
}
}


resource "auth0_action" "custom-email-provider-1" {
code = <<-EOT
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

name = "Custom Email Provider"

runtime = "node18"
deploy = true
supported_triggers {
id = "custom-email-provider"
version = "v1"
}
dependencies {
name = "@aws-sdk/client-ses"
version = "3.682.0"
}
}

# The results of hitting /api/v2/actions/triggers after creating the custom email provider action in the dashboard:
#
# {
# "id": "custom-email-provider",
# "version": "v1",
# "status": "CURRENT",
# "runtimes": [
# "node18-actions"
# ],
# "default_runtime": "node18",
# "binding_policy": "trigger-bound",
# "compatible_triggers": []
# }
