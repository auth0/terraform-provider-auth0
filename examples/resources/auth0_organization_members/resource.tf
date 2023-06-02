resource "auth0_user" "user_1" {
  connection_name = "Username-Password-Authentication"
  email           = "{{.testName}}1@auth0.com"
  password        = "MyPass123$"
}

resource "auth0_user" "user_2" {
  connection_name = "Username-Password-Authentication"
  email           = "{{.testName}}2@auth0.com"
  password        = "MyPass123$"
}

resource "auth0_organization" "my_org" {
  name         = "some-org-{{.testName}}"
  display_name = "{{.testName}}"
}

resource "auth0_organization_members" "my_members" {
  organization_id = auth0_organization.my_org.id
  members         = [auth0_user.user_1.id, auth0_user.user_2.id]
}
