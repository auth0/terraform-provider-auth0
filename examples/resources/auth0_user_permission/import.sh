# This resource can be imported by specifying the
# user ID, resource identifier and permission name separated by "::" (note the double colon)
# <userID>::<resourceServerIdentifier>::<permission>
#
# Example:
terraform import auth0_user_permission.permission "auth0|111111111111111111111111::https://api.travel0.com/v1::read:posts"
