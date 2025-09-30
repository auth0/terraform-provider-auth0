# Retrieve Auth0's outbound IP ranges
data "auth0_outbound_ips" "outbound_ips" {}

locals {
  # To access IPs by region code, create a local map first.
  auth0_regions_map = { for r in data.auth0_outbound_ips.test.regions : r.region => {
    ipv4_cidrs = r.ipv4_cidrs
  } }
}

# Example: Output the results for verification
output "last_updated_at" {
  value = data.auth0_outbound_ips.test.last_updated_at
}

# Example: Output US region IPs
output "us" {
  value = local.auth0_regions_map["US"]["ipv4_cidrs"]
}

# Example: concatenate all Auth0 Outbound IPs
output "all" {
  value = concat([for r in data.auth0_outbound_ips.test.regions : r.ipv4_cidrs]...)
}

# Example: Create AWS security group for specific region
resource "aws_security_group_rule" "auth0_webhook_us" {
  count             = length(local.auth0_regions_map["US"]["ipvd4_cidrs"])
  type              = "ingress"
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  cidr_blocks       = [local.auth0_regions_map["US"]["ipvd4_cidrs"]]
  security_group_id = aws_security_group.app.id
  description       = "Auth0 outbound IPs - US region"
}
