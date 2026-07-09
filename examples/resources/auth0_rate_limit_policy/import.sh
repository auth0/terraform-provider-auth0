# Rate limit policies can be imported using their ID.
#
# You can find existing rate limit policy IDs using the Auth0 Management API.
# https://auth0.com/docs/api/management/v2#!/Rate_Limit_Policies/get_rate_limit_policies
#
# Example:
terraform import auth0_rate_limit_policy.noisy_app "rlp_XXXXXXXXXXXXXXXX"
