# Configure MFA enrollment backoff strategy using the default strategy.
resource "auth0_phone_provider_protection" "default_strategy" {
  type = "default"
}

# Configure MFA enrollment backoff strategy using exponential backoff.
resource "auth0_phone_provider_protection" "exponential_strategy" {
  type = "exponential"
}
