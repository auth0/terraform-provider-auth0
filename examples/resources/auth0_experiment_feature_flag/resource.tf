resource "auth0_experiment_feature_flag" "my_flag" {
  name        = "checkout-passkey-rollout"
  description = "Feature flag driving the passkey A/B test on the login screen."

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
  # The flag starts in draft. After creating at least two variations separately,
  # add status = "active" in a later apply to activate it.
}
