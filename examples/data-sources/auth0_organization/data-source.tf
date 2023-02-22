# An Auth0 Organization loaded using its name.
data "auth0_organization" "some-organization-by-name" {
  name = "my-org"
}

# An Auth0 Organization loaded using its ID.
data "auth0_organization" "some-organization-by-id" {
  organization_id = "org_abcdefghkijklmnopqrstuvwxyz0123456789"
}
