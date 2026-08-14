# The groups assigned to an organization-level role.
data "auth0_organization_role_groups" "my_role_groups" {
  organization_id = "org_abcdefghkijklmn"
  role_id         = "rol_abcdefghkijklmn"
}
