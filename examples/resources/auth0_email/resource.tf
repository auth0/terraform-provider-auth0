# This is an example on how to set up the email provider with Amazon SES.
resource "auth0_email" "amazon_ses_email_provider" {
  name                 = "ses"
  enabled              = true
  default_from_address = "accounts@example.com"

  credentials {
    access_key_id     = "AKIAXXXXXXXXXXXXXXXX"
    secret_access_key = "7e8c2148xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
    region            = "us-east-1"
  }
}

# This is an example on how to set up the email provider with SMTP.
resource "auth0_email" "smtp_email_provider" {
  name                 = "smtp"
  enabled              = true
  default_from_address = "accounts@example.com"

  credentials {
    smtp_host = "your.smtp.host.com"
    smtp_port = 583
    smtp_user = "SMTP Username"
    smtp_pass = "SMTP Password"
  }
}

# This is an example on how to set up the email provider with Sendgrid.
resource "auth0_email" "sendgrid_email_provider" {
  name                 = "sendgrid"
  enabled              = true
  default_from_address = "accounts@example.com"

  credentials {
    api_key = "secretAPIKey"
  }
}
