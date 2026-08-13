resource "auth0_role" "my_role" {
  name        = "My Role - (Managed by Terraform)"
  description = "Role Description..."
}

resource "auth0_organization" "my_organization" {
  name         = "my-organization"
  display_name = "My Organization"
}

# A role scoped to a single organization. Changing type or owner_id afterwards
# replaces the role, as neither can be updated. (EA only)
resource "auth0_role" "my_organization_role" {
  name        = "My Organization Role - (Managed by Terraform)"
  description = "Organization Role Description..."
  type        = "organization"
  owner_id    = auth0_organization.my_organization.id
}
