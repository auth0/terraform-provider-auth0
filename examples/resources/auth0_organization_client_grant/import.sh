# This resource can be imported by specifying the
# organization ID and client grant ID separated by "::" (note the double colon)
# <organizationID>::<clientGrantID>
#
# Example:
terraform import auth0_organization_client_grant.my_org_client_grant "org_XXXXX::cgr_XXXXX"
