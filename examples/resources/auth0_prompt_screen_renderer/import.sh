# This resource can be imported using the prompt name and screen_name.
#
# As this is not a resource identifiable by an ID within the Auth0 Management API,
# login can be imported using the prompt name and screen name using the format:
# prompt_name:screen_name
#
# Example:
terraform import auth0_prompt_screen_renderer "login-id:login-id"
