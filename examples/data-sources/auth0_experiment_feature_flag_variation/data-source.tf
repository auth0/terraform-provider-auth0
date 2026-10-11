# An Auth0 Experiment Center feature flag variation loaded using its feature flag ID and variation ID.
data "auth0_experiment_feature_flag_variation" "treatment" {
  feature_flag_id = "flg_XXXXXXXXXXXXXXX"
  variation_id    = "var_XXXXXXXXXXXXXXX"
}
