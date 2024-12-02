resource "auth0_self_service_profile_custom_text" "sso_custom_text" {
  sso_id   = "some-sso-id"
  language = "en"
  page     = "get-started"
  body = jsonencode(
    {
      "introduction" : "Welcome! With only a few steps you'll be able to setup your new custom text."
    }
  )
}

