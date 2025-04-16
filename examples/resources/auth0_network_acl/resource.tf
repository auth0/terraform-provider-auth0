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
      geo_country_codes     = ["US", "CA"]
      geo_subdivision_codes = ["US-NY", "CA-ON"]
      ipv4_cidrs            = ["192.168.1.0/24", "10.0.0.0/8"]
      ipv6_cidrs            = ["2001:db8::/32"]
    }
  }
}
