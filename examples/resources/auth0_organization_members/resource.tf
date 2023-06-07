resource "auth0_user" "user_1" {
  connection_name = "Username-Password-Authentication"
  email           = "myuser1@auth0.com"
  password        = "MyPass123$"
}

resource "auth0_user" "user_2" {
  connection_name = "Username-Password-Authentication"
  email           = "myuser2@auth0.com"
  password        = "MyPass123$"
}

resource "auth0_organization" "my_org" {
  name         = "some-org"
  display_name = "Some Organization"
}

resource "auth0_organization_members" "my_members" {
  organization_id = auth0_organization.my_org.id
  members         = [auth0_user.user_1.id, auth0_user.user_2.id]
}
