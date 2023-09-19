# This is an example on how to set up the email provider with Amazon SES.
resource "auth0_email_provider" "amazon_ses_email_provider" {
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
resource "auth0_email_provider" "smtp_email_provider" {
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
resource "auth0_email_provider" "sendgrid_email_provider" {
  name                 = "sendgrid"
  enabled              = true
  default_from_address = "accounts@example.com"

  credentials {
    api_key = "secretAPIKey"
  }
}


# This is an example on how to set up the email provider with Azure CS.
resource "auth0_email_provider" "smtp_email_provider" {
  name                 = "azure_cs"
  enabled              = true
  default_from_address = "accounts@example.com"

  credentials {
    azure_cs_connection_string = "azure_cs_connection_string"
  }
}


# This is an example on how to set up the email provider with MS365.
resource "auth0_email_provider" "smtp_email_provider" {
  name                 = "ms365"
  enabled              = true
  default_from_address = "accounts@example.com"

  credentials {
    ms365_tenant_id     = "ms365_tenant_id"
    ms365_client_id     = "ms365_client_id"
    ms365_client_secret = "ms365_client_secret"
  }
}
