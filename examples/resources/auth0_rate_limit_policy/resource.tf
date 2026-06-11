resource "auth0_rate_limit_policy" "noisy_app" {
  resource          = "oauth_authentication_api"
  consumer          = "client"
  consumer_selector = "client_id:abc123"

  configuration {
    action       = "redirect"
    limit        = 1000
    redirect_uri = "https://example.com/rate-limited"
  }
}
