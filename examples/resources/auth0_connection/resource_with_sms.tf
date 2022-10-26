# This is an example of an SMS connection.

resource "auth0_connection" "sms" {
  name     = "SMS-Connection"
  strategy = "sms"

  options {
    name                   = "SMS OTP"
    twilio_sid             = "<twilio-sid>"
    twilio_token           = "<twilio-token>"
    from                   = "<phone-number>"
    syntax                 = "md_with_macros"
    template               = "Your one-time password is @@password@@"
    messaging_service_sid  = "<messaging-service-sid>"
    disable_signup         = false
    brute_force_protection = true
    forward_request_info   = true

    totp {
      time_step = 300
      length    = 6
    }

    provider    = "sms_gateway"
    gateway_url = "https://somewhere.com/sms-gateway"
    gateway_authentication {
      method                = "bearer"
      subject               = "test.us.auth0.com:sms"
      audience              = "https://somewhere.com/sms-gateway"
      secret                = "4e2680bb72ec2ae24836476dd37ed6c2"
      secret_base64_encoded = false
    }
  }
}

# This is an example of an SMS connection with a custom SMS gateway.

resource "auth0_connection" "sms" {
  name                 = "custom-sms-gateway"
  is_domain_connection = false
  strategy             = "sms"

  options {
    disable_signup         = false
    name                   = "sms"
    from                   = "+15555555555"
    syntax                 = "md_with_macros"
    template               = "@@password@@"
    brute_force_protection = true
    provider               = "sms_gateway"
    gateway_url            = "https://somewhere.com/sms-gateway"
    forward_request_info   = true

    totp {
      time_step = 300
      length    = 6
    }

    gateway_authentication {
      method                = "bearer"
      subject               = "test.us.auth0.com:sms"
      audience              = "https://somewhere.com/sms-gateway"
      secret                = "4e2680bb74ec2ae24736476dd37ed6c2"
      secret_base64_encoded = false
    }
  }
}
