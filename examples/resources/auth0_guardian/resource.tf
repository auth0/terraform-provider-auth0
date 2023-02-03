resource "auth0_guardian" "my_guardian" {
  policy        = "all-applications"
  email         = true
  otp           = true
  recovery_code = true

  webauthn_platform {
    enabled = true
  }

  webauthn_roaming {
    enabled           = true
    user_verification = "required"
  }

  phone {
    enabled       = true
    provider      = "auth0"
    message_types = ["sms", "voice"]

    options {
      enrollment_message   = "{{code}} is your verification code for {{tenant.friendly_name}}. Please enter this code to verify your enrollment."
      verification_message = "{{code}} is your verification code for {{tenant.friendly_name}}."
    }
  }

  push {
    enabled  = true
    provider = "sns"

    amazon_sns {
      aws_access_key_id                 = "test1"
      aws_region                        = "us-west-1"
      aws_secret_access_key             = "secretKey"
      sns_apns_platform_application_arn = "test_arn"
      sns_gcm_platform_application_arn  = "test_arn"
    }

    custom_app {
      app_name        = "CustomApp"
      apple_app_link  = "https://itunes.apple.com/us/app/my-app/id123121"
      google_app_link = "https://play.google.com/store/apps/details?id=com.my.app"
    }
  }

  duo {
    enabled         = true
    integration_key = "someKey"
    secret_key      = "someSecret"
    hostname        = "api-hostname"
  }
}
