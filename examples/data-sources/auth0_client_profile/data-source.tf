# An Auth0 client profile (application profile) can be retrieved by its ID.
data "auth0_client_profile" "my_profile" {
  id = "cpr_XXXXXXXXXXXXXXXXXXXXXX"
}
