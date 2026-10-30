# Each rule sets at least one of `match` or `not_match`. Each attribute block
# sets exactly one operator: `one_of`, `contains`, `starts_with`, `ends_with`
# or `exists`. Each `match` or `not_match` block sets between 1 and 10
# attributes, and each `anonymous_session_metadata` block counts as one.
resource "auth0_experiment_segment" "campaign_traffic" {
  name        = "email-campaign-traffic"
  description = "Gold-plan visitors from an email campaign, excluding internal traffic."

  # Omit `status` or set it to "active" on creation. To archive the segment,
  # set it to "archived" in a later apply. Archiving cannot be undone, and
  # `name`, `description` and `rules` cannot change after that.
  # status = "archived"

  rules {
    match {
      country {
        one_of = ["US", "CA"]
      }

      domain {
        starts_with = ["www"]
      }

      device_type {
        contains = ["mobile"]
      }

      client_id {
        exists = true
      }

      # Set one block per metadata key. Use the key without the
      # `anonymous_session_metadata.` prefix.
      anonymous_session_metadata {
        key    = "plan"
        one_of = ["gold"]
      }

      anonymous_session_metadata {
        key      = "utm_source"
        contains = ["email"]
      }
    }

    not_match {
      region {
        ends_with = ["-west"]
      }

      browser {
        ends_with = ["bot"]
      }

      anonymous_session_metadata {
        key    = "internal"
        exists = true
      }
    }
  }

  rules {
    match {
      organization_id {
        exists = true
      }
    }
  }
}
