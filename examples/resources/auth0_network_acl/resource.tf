# Example of auth0_network_acl with match criteria

resource "auth0_network_acl" "my_network_acl_match" {
  description = "Example with match network ACL"
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

    }
  }
}

# Example of auth0_network_acl with not-match criteria
resource "auth0_network_acl" "my_network_acl_not_match" {
  description = "Example with not match network ACL"
  active      = true
  priority    = 3
  rule {
    action {
      log = true
    }
    scope = "authentication"
    not_match {
      asns       = [9876]
      ipv4_cidrs = ["192.168.1.0/24", "10.0.0.0/8"]
      ipv6_cidrs = ["2001:db8::/32"]
    }
  }
}
