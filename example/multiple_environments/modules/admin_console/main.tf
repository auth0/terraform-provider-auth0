resource "auth0_client" "admin" {
  name     = "Admin Console"
  app_type = "non_interactive"
}

data "auth0_tenant" "current" {}

resource "auth0_client_grant" "admin_management_api" {
  client_id = auth0_client.admin.client_id
  audience  = data.auth0_tenant.current.management_api_identifier
  scope     = ["read:users", "create:users", "update:users", "delete:users"]
}
