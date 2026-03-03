# This is an example on how to set up the phone provider with Twilio.
resource "auth0_phone_provider" "twilio_phone_provider" {
  name     = "twilio"
  disabled = false
  credentials {
    auth_token = "secretAuthToken"
  }

  configuration {
    delivery_methods = ["text", "voice"]
    default_from     = "+1234567890"
    sid              = "ACXXXXXXXXXXXXXXXX"
    mssid            = "MSXXXXXXXXXXXXXXXX"
  }
}



# This is an example on how to set up the phone provider with a custom action.
# Make sure a corresponding action exists with custom-phone-provider as supported triggers
resource "auth0_action" "send_custom_phone" {
  name    = "Custom Phone Provider"
  runtime = "node22"
  deploy  = true
  code    = <<-EOT
    /**
     * Handler to be executed while sending a phone notification
     * @param {Event} event - Details about the user and the context in which they are logging in.
     * @param {CustomPhoneProviderAPI} api - Methods and utilities to help change the behavior of sending a phone notification.
     */
    exports.onExecuteCustomPhoneProvider = async (event, api) => {
        // Code goes here
        return;
    };
  EOT


  supported_triggers {
    id      = "custom-phone-provider"
    version = "v1"
  }
}

resource "auth0_phone_provider" "custom_phone_provider" {
  depends_on = [auth0_action.send_custom_phone] # Ensure the action is created first with `custom-phone-provider` as the supported_triggers
  name       = "custom"                         # Indicates a custom implementation
  disabled   = false                            # Disable the default phone provider
  configuration {
    delivery_methods = ["text", "voice"]
  }
  credentials {}
}
