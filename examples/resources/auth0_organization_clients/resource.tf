resource "auth0_organization" "my_organization" {
  name         = "my-organization"
  display_name = "My Organization"
}

resource "auth0_client" "my_client" {
  name = "My Application"
}

resource "auth0_client" "my_other_client" {
  name = "My Other Application"
}

# Entitle both applications to the organization in one authoritative resource (EA only).
# Use this instead of auth0_organization_client when you want Terraform to own the full
# set of associations; do not use both resources on the same organization.
resource "auth0_organization_clients" "my_org_clients" {
  organization_id = auth0_organization.my_organization.id

  clients {
    client_id             = auth0_client.my_client.id
    use_for_member_access = true
  }

  clients {
    client_id = auth0_client.my_other_client.id
  }
}
