# This resource can be imported by specifying the
# role ID, resource identifier, and permission name separated by "::" (note the double colon)
# <roleID>::<resourceServerIdentifier>::<permission>
#
# Example:
terraform import auth0_role_permission.permission "rol_XXXXXXXXXXXXX::https://example.com::read:foo"
