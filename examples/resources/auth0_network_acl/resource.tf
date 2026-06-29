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

# Example of auth0_network_acl with hostname and connecting IP restrictions
resource "auth0_network_acl" "block_canonical" {
  description = "Block canonical domain except from proxy"
  active      = true
  priority    = 5
  rule {
    action {
      block = true
    }
    scope = "tenant"
    match {
      hostnames             = ["mytenant1.us.auth0.com"]
      connecting_ipv6_cidrs = ["2001:db8::/32", "::1"]
    }
    not_match {
      hostnames             = ["mytenant2.us.auth0.com"]
      connecting_ipv4_cidrs = ["203.0.113.0/24"]
    }
  }
}