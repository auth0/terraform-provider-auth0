resource "auth0_connection" "azure_ad" {
  name           = "connection-azure-ad"
  strategy       = "waad"
  show_as_button = true
  options {
    identity_api      = "azure-active-directory-v1.0"
    client_id         = "123456"
    client_secret     = "123456"
    strategy_version  = 2
    user_id_attribute = "oid"
    app_id            = "app-id-123"
    tenant_domain     = "example.onmicrosoft.com"
    domain            = "example.onmicrosoft.com"
    domain_aliases = [
      "example.com",
      "api.example.com"
    ]
    icon_url               = "https://example.onmicrosoft.com/assets/logo.png"
    use_wsfed              = false
    waad_protocol          = "openid-connect"
    waad_common_endpoint   = false
    max_groups_to_retrieve = 250
    api_enable_users       = true
    scopes = [
      "basic_profile",
      "ext_groups",
      "ext_profile"
    ]
    set_user_root_attributes               = "on_each_login"
    should_trust_email_verified_connection = "never_set_emails_as_verified"
    upstream_params = jsonencode({
      "screen_name" : {
        "alias" : "login_hint"
      }
    })
    non_persistent_attrs = ["ethnicity", "gender"]
  }
}
