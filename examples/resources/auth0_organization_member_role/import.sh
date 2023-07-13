# This resource can be imported by specifying the
# organization ID, user ID and role ID separated by "::" (note the double colon)
# <organizationID>::<userID>::<roleID>
#
# Example:
terraform import auth0_organization_member_role.my_org_member_role "org_XXXXX::auth0|XXXXX::role_XXXX"
