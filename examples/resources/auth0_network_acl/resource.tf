# Example:
resource "auth0_network_acl" "my_network_acl" {
  description = "My network ACL"
  active      = true
  priority    = 1
  rule {
    action {
      allow = true
    }
    scope = "management"
    match {
      geo_country_codes = ["US", "CA"]
    }
  }
}
