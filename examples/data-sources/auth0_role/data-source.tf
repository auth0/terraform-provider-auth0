# An Auth0 Role loaded using its name.
data "auth0_role" "some-role-by-name" {
  name = "my-role"
}

# An Auth0 Role loaded using its ID.
data "auth0_role" "some-role-by-id" {
  role_id = "abcdefghkijklmnopqrstuvwxyz0123456789"
}
