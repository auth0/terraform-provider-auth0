# Data Source Example: Retrieve Auth0 Phone Notification Templates
# This example shows how to fetch existing phone notification templates from your Auth0 tenant

# Retrieve the OTP enrollment template
data "auth0_branding_phone_notification_template" "otp_enrollment" {
  template_id = "tem_xxxxxxxxxxxxxxxxx"
}

# Output the template ID
output "otp_enrollment_id" {
  description = "The ID of the OTP enrollment phone notification template"
  value       = data.auth0_branding_phone_notification_template.otp_enrollment.id
}
