terraform {
  required_providers {
    auth0 = {
      source = "auth0/auth0"
    }
  }
}

provider "auth0" {}

resource auth0_role reader {
	name = "Reader"
}

resource auth0_role admin {
	name = "Admin"
}

resource auth0_user user {
	email = "test-user@auth0.com"
	connection_name = "Username-Password-Authentication"
	email_verified = true
	password = "MyPass123$"
}

resource auth0_organization some_org{
	name = "org-admin"
	display_name = "Admin"
}

resource auth0_organization_member member {
  depends_on = [ auth0_user.user, auth0_organization.some_org, auth0_role.reader, auth0_role.admin ]
  organization_id = auth0_organization.some_org.id
  user_id = auth0_user.user.id
  roles = [ auth0_role.reader.id, auth0_role.admin.id ]
}
