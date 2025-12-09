# Phone Notification Template - OTP Enrollment
# Configure the OTP enrollment phone notification template with SMS and voice support.
resource "auth0_branding_phone_notification_template" "otp_enrollment" {
  type     = "otp_enroll"
  disabled = false

  content {
    from = "+1234567890"

    body {
      text  = "Your enrollment code is: @{code}"
      voice = "Your enrollment code is @{code}"
    }
  }
}

# Phone Notification Template - OTP Verification
# Configure the OTP verification phone notification template.
resource "auth0_branding_phone_notification_template" "otp_verification" {
  type     = "otp_verify"
  disabled = false

  content {
    from = "+1234567890"

    body {
      text  = "Your verification code is: @{code}"
      voice = "Your verification code is @{code}"
    }
  }
}
