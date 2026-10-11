resource "auth0_experiment_feature_flag" "my_flag" {
  name        = "checkout-passkey-rollout"
  description = "Feature flag driving the passkey A/B test on the login screen."

  parameters {
    name  = "show_passkey"
    type  = "boolean"
    value = "false"
  }

  parameters {
    name  = "cta_label"
    type  = "string"
    value = "Sign in"
  }

  parameters {
    name = "obj1"
    type = "object"
    value = jsonencode({
      label   = "A"
      enabled = true
    })
  }
}

# A control variation has no overrides, so it serves the flag defaults.
resource "auth0_experiment_feature_flag_variation" "control" {
  feature_flag_id = auth0_experiment_feature_flag.my_flag.id
  name            = "control"
  description     = "Serves the feature flag defaults."
}

# Each override must name a parameter of the feature flag, match its type,
# and differ from its default value. `value` is always a string; use a
# JSON-encoded string when `type` is "object" or "array".
resource "auth0_experiment_feature_flag_variation" "treatment" {
  feature_flag_id = auth0_experiment_feature_flag.my_flag.id
  name            = "passkey-treatment"
  description     = "Shows the passkey enrollment screen."

  overrides {
    name  = "show_passkey"
    type  = "boolean"
    value = "true"
  }

  overrides {
    name  = "cta_label"
    type  = "string"
    value = "Continue with passkey"
  }

  overrides {
    name = "obj1"
    type = "object"
    value = jsonencode({
      label   = "B"
      enabled = true
    })
  }
}
