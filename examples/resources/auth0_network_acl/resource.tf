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

# Example of auth0_network_acl using Auth0-curated blocklists (Early Access).
#
# The `auth0_managed` field requires the `advanced-breached-password-detection`
# entitlement and the `tenant_acl_curated_blocklists` feature flag on the tenant.
# It may be set on only one of `match` or `not_match` within a rule.
resource "auth0_network_acl" "block_icloud_relay" {
  description = "Block iCloud Private Relay egress proxies"
  active      = true
  priority    = 7
  rule {
    action {
      block = true
    }
    scope = "authentication"
    # Block requests matching the curated iCloud Private Relay proxy list.
    match {
      auth0_managed = ["auth0.icloud_relay_proxy"]
    }
  }
}

# Example using `not_match` to allow all traffic *unless* it comes from a
# low-reputation curated blocklist. `auth0_managed` may live on only one of
# `match` / `not_match`, so this demonstrates the mutual-exclusivity boundary.
resource "auth0_network_acl" "allow_unless_low_reputation" {
  description = "Allow traffic unless it is on the low-reputation blocklist"
  active      = true
  priority    = 8
  rule {
    action {
      allow = true
    }
    scope = "authentication"
    not_match {
      auth0_managed = ["auth0.low_reputation"]
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