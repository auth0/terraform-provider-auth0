#!/bin/bash
# Example: Import an existing Auth0 phone notification template into Terraform state

# Replace TEMPLATE_ID with the actual template ID from your Auth0 tenant

terraform import auth0_branding_phone_notification_template.otp_enrollment "tem_xxxxxxxxxxxxxxxxxxx"

