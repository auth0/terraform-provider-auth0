resource "auth0_role" "reader" {
  name = "Reader"
}

resource "auth0_role" "writer" {
  name = "Writer"
}

resource "auth0_user" "user" {
  connection_name = "Username-Password-Authentication"
  email           = "test-user@auth0.com"
  password        = "MyPass123$"
}

resource "auth0_organization" "my_org" {
  name         = "some-org"
  display_name = "Some Org"
}

resource "auth0_organization_member" "my_org_member" {
  organization_id = auth0_organization.my_org.id
  user_id         = auth0_user.user.id
}

resource "auth0_organization_member_role" "role1" {
  organization_id = auth0_organization.my_org.id
  user_id         = auth0_user.user.id
  role_id         = auth0_role.reader.id
}

resource "auth0_organization_member_role" "role2" {
  organization_id = auth0_organization.my_org.id
  user_id         = auth0_user.user.id
  role_id         = auth0_role.writer.id
}
