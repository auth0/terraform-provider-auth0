resource "auth0_experiment_feature_flag" "my_flag" {
  name        = "checkout-passkey-rollout"
  description = "Feature flag driving the passkey A/B test on the login screen."

  # At least one parameter is required. `value` is always a string;
  # use a JSON-encoded string when `type = "object"`.
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

  # Activating the flag (status = "active") requires at least two variations.
  variation {
    name        = "control"
    description = "The current login screen."
    overrides = {
      cta_label = "Sign up"
    }
  }

  variation {
    name = "treatment"
    overrides = {
      show_passkey = "true"
      cta_label    = "Continue with passkey"
    }
  }

  status = "active"
}
