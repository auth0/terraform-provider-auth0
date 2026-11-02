resource "auth0_experiment_feature_flag" "my_flag" {
  name        = "checkout-passkey-rollout"
  description = "Feature flag driving the passkey A/B test on the login screen."

  # Activation needs at least two variations. Omit status to keep the flag in draft.
  status = "active"

  # At least one parameter is required. `value` is always a string;
  # use a JSON-encoded string when `type` is "object" or "array".
  parameters {
    name        = "show_passkey"
    type        = "boolean"
    value       = "false"
    description = "Whether the passkey enrollment screen is shown."
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

  parameters {
    name  = "arr1"
    type  = "array"
    value = jsonencode(["a", "b"])
  }

  # A control variation has no overrides, so it uses the parameter defaults.
  variations {
    name        = "control"
    description = "Current login screen."
  }

  # Each override must use a parameter name and type, and a value different from the default.
  variations {
    name        = "treatment"
    description = "Login screen with passkey enrollment."

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
}

# Use variation_ids to reference a variation by ID, for example in an experiment allocation.
output "treatment_variation_id" {
  value = auth0_experiment_feature_flag.my_flag.variation_ids["treatment"]
}
