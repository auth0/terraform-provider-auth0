data "auth0_rate_limit_policies" "filtered" {
  resource          = "oauth_authentication_api"
  consumer          = "client"
  consumer_selector = "default"
}
