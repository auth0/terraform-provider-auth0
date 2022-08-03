resource "auth0_role" "reader" {
  name = "Reader"
}

resource "auth0_role" "admin" {
  name = "Admin"
}

resource "auth0_user" "user" {
  email           = "test-user@auth0.com"
  connection_name = "Username-Password-Authentication"
  email_verified  = true
  password        = "MyPass123$"
}

resource "auth0_organization" "my_org" {
  name         = "org-admin"
  display_name = "Admin"
}

resource "auth0_organization_member" "my_org_member" {
  organization_id = auth0_organization.my_org.id
  user_id         = auth0_user.user.id
  roles           = [auth0_role.reader.id, auth0_role.admin.id]
}
