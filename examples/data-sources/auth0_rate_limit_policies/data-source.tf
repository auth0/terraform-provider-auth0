data "auth0_rate_limit_policies" "all_by_resource" {
  resource = "oauth_authentication_api"
}
