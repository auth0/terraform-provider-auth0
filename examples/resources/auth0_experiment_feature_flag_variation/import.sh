# This resource can be imported by specifying the
# feature flag ID and variation ID separated by "::" (note the double colon)
# <featureFlagID>::<variationID>
#
# Example:
terraform import auth0_experiment_feature_flag_variation.treatment "flg_XXXXXXXXXXXXXXX::var_XXXXXXXXXXXXXXX"
