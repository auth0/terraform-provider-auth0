resource "auth0_self_service_profile_custom_text" "sso_custom_text" {
  sso_id   = "some-sso-id"
  language = "en"
  page     = "get-started"
  body     = "{}"
}

