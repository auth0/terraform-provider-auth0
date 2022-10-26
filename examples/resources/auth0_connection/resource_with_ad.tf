resource "auth0_connection" "ad" {
  name           = "connection-active-directory"
  display_name = "Active Directory Connection"
  strategy       = "ad"
  show_as_button = true

  options {
    brute_force_protection = true
    tenant_domain          = "example.com"
    icon_url = "https://example.com/assets/logo.png"
    domain_aliases = [
      "example.com",
      "api.example.com"
    ]
    ips                      = ["192.168.1.1", "192.168.1.2"]
    set_user_root_attributes = "on_each_login"
    non_persistent_attrs     = ["ethnicity", "gender"]
    upstream_params = jsonencode({
      "screen_name" : {
        "alias" : "login_hint"
      }
    })
    use_cert_auth = false
    use_kerberos = false
    disable_cache = false
  }
}