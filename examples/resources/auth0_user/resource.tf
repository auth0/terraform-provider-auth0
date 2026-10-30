resource "auth0_user" "user" {
  connection_name = "Username-Password-Authentication"
  user_id         = "12345"
  username        = "unique_username"
  name            = "Firstname Lastname"
  nickname        = "some.nickname"
  email           = "test@test.com"
  email_verified  = true
  password        = "passpass$12$12"
  picture         = "https://www.example.com/a-valid-picture-url.jpg"
}


# Create a user with custom_domain_header
resource "auth0_user" "auth0_user_with_custom_domain" {
  connection_name      = "Username-Password-Authentication"
  username             = "your_new_user_"
  email                = "change.username@acceptance.test.com"
  email_verified       = true
  password             = "MyPass123$"
  custom_domain_header = "my-custom.domain.org"
}

# The password can be set as a write-only argument so it is never persisted to Terraform state.
# It can be sourced from an ephemeral value (e.g. a secrets manager) and is mutually exclusive
# with `password`. Bump `password_wo_version` to send a new password.
#
# NOTE: Write-only arguments require Terraform 1.11 or later.
resource "auth0_user" "user_write_only_password" {
  connection_name     = "Username-Password-Authentication"
  username            = "write_only_user"
  email               = "write.only@test.com"
  password_wo         = var.user_password # can be an ephemeral value
  password_wo_version = 1
}
