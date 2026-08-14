# The organization members with a direct assignment of an organization-level role.
data "auth0_organization_role_members" "my_role_members" {
  organization_id = "org_abcdefghkijklmn"
  role_id         = "rol_abcdefghkijklmn"
}
