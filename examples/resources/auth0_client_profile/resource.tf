# Early Access only. The Management API client needs the create, read, update and delete
# `client_profiles` scopes.
#
# A tenant can have only one client profile. Destroying it removes the governance it applies
# to all matched clients, so protect it from accidental deletion.

# Minimal example.
resource "auth0_client_profile" "minimal" {
  name = "Dynamic client profile"

  lifecycle {
    prevent_destroy = true
  }
}

# Full example.
resource "auth0_client_profile" "full" {
  name        = "Dynamic client profile"
  description = "Runtime policy for dynamically registered clients."

  apply_policy {
    external_metadata_type       = "cimd"
    external_metadata_created_by = "client"
  }

  client_configuration {
    refresh_token {
      token_lifetime {
        default_value = 2592000 # 30 days.
      }
      idle_token_lifetime {
        default_value = 1296000 # 15 days.
      }
      leeway {
        default_value = 0
      }
    }

    organization_usage {
      default_value = "require"
    }
    organization_require_behavior {
      default_value = "pre_login_prompt"
    }
    organization_discovery_methods {
      default_value = ["email", "organization_name"]
    }
    require_proof_of_possession {
      default_value = true
    }
  }

  lifecycle {
    prevent_destroy = true
  }
}

# If a client profile already exists on the tenant, adopt it instead of creating a new one
# (Terraform 1.5 or later). Replace the ID with the one of the existing profile.
#
# import {
#   to = auth0_client_profile.full
#   id = "cpr_XXXXXXXXXXXXXXXX"
# }
