resource "auth0_tenant" "my_tenant" {
  friendly_name           = "Tenant Name"
  picture_url             = "http://example.com/logo.png"
  support_email           = "support@example.com"
  support_url             = "http://example.com/support"
  allowed_logout_urls     = ["http://example.com/logout"]
  session_lifetime        = 8760
  sandbox_version         = "12"
  enabled_locales         = ["en"]
  default_redirection_uri = "https://example.com/login"

  change_password {
    enabled = true
    html    = "<html>Change Password</html>"
  }

  guardian_mfa_page {
    enabled = true
    html    = "<html>MFA</html>"
  }

  error_page {
    html          = "<html>Error Page</html>"
    show_log_link = true
    url           = "https://example.com/errors"
  }

  session_cookie {
    mode = "non-persistent"
  }

  universal_login {
    colors {
      primary         = "#0059d6"
      page_background = "#000000"
    }
  }

  flags {
    universal_login                        = true
    disable_clickjack_protection_headers   = true
    enable_public_signup_user_exists_error = true
    use_scope_descriptions_for_consent     = true
    no_disclose_enterprise_connections     = false
    disable_management_api_sms_obfuscation = false
    disable_fields_map_fix                 = false
  }
}
