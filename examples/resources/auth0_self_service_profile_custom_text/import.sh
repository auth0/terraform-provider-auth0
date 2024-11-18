# This resource can be imported by specifying the
# sso-profile-id, language and page separated by "::" (note the double colon)
# <sso-profile-id>::<language>::<page>
#
# Example
terraform import auth0_self_service_profile_custom_text.example "some-sso-id::en::get-started"
