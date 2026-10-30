# An Auth0 Experiment Center feature flag loaded using its ID.
data "auth0_experiment_feature_flag" "my_flag" {
  id = "flg_XXXXXXXXXXXXXXX"
}

# The ID of a variation of the feature flag, looked up by its name.
output "treatment_variation_id" {
  value = data.auth0_experiment_feature_flag.my_flag.variation_ids["treatment"]
}
