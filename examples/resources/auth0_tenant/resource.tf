resource "auth0_tenant" "my_tenant" {
  friendly_name           = "Tenant Name"
  picture_url             = "http://example.com/logo.png"
  support_email           = "support@example.com"
  support_url             = "http://example.com/support"
  allowed_logout_urls     = ["http://example.com/logout"]
  session_lifetime        = 8760
  sandbox_version         = "22"
  enabled_locales         = ["en"]
  default_redirection_uri = "https://example.com/login"

  flags {
    disable_clickjack_protection_headers   = true
    enable_public_signup_user_exists_error = true
    use_scope_descriptions_for_consent     = true
    no_disclose_enterprise_connections     = false
    disable_management_api_sms_obfuscation = false
    disable_fields_map_fix                 = false
  }

  session_cookie {
    mode = "non-persistent"
  }

  sessions {
    oidc_logout_prompt_enabled = false
  }

  error_page {
    html          = "<html></html>"
    show_log_link = false
    url           = "https://example.com/error"
  }
}
