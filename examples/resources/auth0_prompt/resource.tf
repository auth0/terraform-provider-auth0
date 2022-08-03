resource "auth0_prompt" "my_prompt" {
  universal_login_experience     = "new"
  identifier_first               = false
  webauthn_platform_first_factor = true
}
