resource "auth0_organization" "my_organization" {
  name         = "my-organization"
  display_name = "My Organization"
}

resource "auth0_client" "my_client" {
  name = "My Application"
}

# Entitle the application to the organization (EA only). This is distinct
# from `auth0_organization_client_grant`, which associates the organization
# with a client_grant instead of an application.
resource "auth0_organization_client" "my_org_client" {
  organization_id       = auth0_organization.my_organization.id
  client_id             = auth0_client.my_client.id
  use_for_member_access = true
}
