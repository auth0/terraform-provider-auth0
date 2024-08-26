# An Auth0 Self-Service- Profile loaded using it's ID.
data "auth0_self_service_profile" "auth0_self_service_profile" {
  id = "some-profile-id"
}
