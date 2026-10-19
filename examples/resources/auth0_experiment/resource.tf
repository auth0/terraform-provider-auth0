resource "auth0_experiment_feature_flag" "login_ui" {
  name        = "new-login-ui"
  description = "Controls whether the new login UI is shown."

  parameters {
    name  = "show_new_login"
    type  = "boolean"
    value = "false"
  }

  # Activate the flag in a later apply, after both variations exist.
  # status = "active"
}

# A control variation has no overrides, so it serves the flag defaults.
resource "auth0_experiment_feature_flag_variation" "control" {
  feature_flag_id = auth0_experiment_feature_flag.login_ui.id
  name            = "control"
}

resource "auth0_experiment_feature_flag_variation" "new_login" {
  feature_flag_id = auth0_experiment_feature_flag.login_ui.id
  name            = "new-login"

  overrides {
    name  = "show_new_login"
    type  = "boolean"
    value = "true"
  }
}

# Percentage strategy: split traffic between variations by weight.
resource "auth0_experiment" "percentage" {
  name                = "login-ui-percentage-test"
  description         = "Show the new login UI to half of the traffic."
  feature_flag_id     = auth0_experiment_feature_flag.login_ui.id
  authentication_flow = "authentication"
  allocation_strategy = "percentage"

  # Omit `status` or set it to "draft" on creation. Set it to "active" in a
  # later apply, after the feature flag is active. An active experiment
  # cannot be deleted: set it to "paused" or "completed" before destroying it.
  status = "draft"

  allocations {
    variation_id = auth0_experiment_feature_flag_variation.control.variation_id
    is_control   = true
    weight       = 50
  }

  allocations {
    variation_id = auth0_experiment_feature_flag_variation.new_login.variation_id
    is_control   = false
    weight       = 50
  }

  # To ramp exposure in steps, remove `weight` from every `allocations` block
  # and set the levels instead:
  # levels = [10, 50, 100]

  # Set `default_config` to "tenant" or "flag" only for an Auth0-managed flag.
  # default_config = "flag"
}

resource "auth0_experiment_segment" "us_traffic" {
  name = "us-traffic"

  rules {
    match {
      country {
        contains = ["US"]
      }
    }
  }
}

# Segment strategy: route each segment to a variation; other traffic gets the
# fallback allocation.
resource "auth0_experiment" "segment" {
  name                = "login-ui-segment-test"
  description         = "Keep US traffic on the current login UI and show the new UI elsewhere."
  feature_flag_id     = auth0_experiment_feature_flag.login_ui.id
  authentication_flow = "mfa_challenge"
  allocation_strategy = "segment"
  status              = "draft"

  # Each targeted allocation sets a segment and a unique priority, where 1 is
  # the highest. Segment allocations do not use `weight` or `levels`.
  allocations {
    variation_id = auth0_experiment_feature_flag_variation.control.variation_id
    is_control   = true
    segment_id   = auth0_experiment_segment.us_traffic.id
    priority     = 1
  }

  # The fallback allocation serves traffic that matches no segment. Omit
  # `segment_id` on the fallback.
  allocations {
    variation_id = auth0_experiment_feature_flag_variation.new_login.variation_id
    is_control   = false
    is_fallback  = true
  }
}
