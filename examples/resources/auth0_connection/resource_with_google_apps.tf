resource "auth0_connection" "google_apps" {
  name                 = "connection-google-apps"
  is_domain_connection = false
  strategy             = "google-apps"
  show_as_button       = false
  options {
    client_id        = ""
    client_secret    = ""
    domain           = "example.com"
    tenant_domain    = "example.com"
    domain_aliases   = ["example.com", "api.example.com"]
    api_enable_users = true
    scopes           = ["ext_profile", "ext_groups"]
    icon_url         = "http://example.com/assets/logo.png"
    upstream_params = jsonencode({
      "screen_name" : {
        "alias" : "login_hint"
      }
    })
    set_user_root_attributes = "on_each_login"
    non_persistent_attrs     = ["ethnicity", "gender"]
  }
}