# An Auth0 Role loaded using its name.
data "auth0_role" "some-role-by-name" {
  name = "my-role"
}

# An Auth0 Role loaded using its ID.
data "auth0_role" "some-role-by-id" {
  role_id = "abcdefghkijklmnopqrstuvwxyz0123456789"
}

# An organization-level Auth0 Role loaded using the role name. Several
# organizations can own a role of the same name, so the type and the ID of the
# organization owning it are needed to find the right one. (EA only)
data "auth0_role" "some-organization-role-by-name" {
  name     = "my-organization-role"
  type     = "organization"
  owner_id = "org_abcdefghkijklmn"
}
