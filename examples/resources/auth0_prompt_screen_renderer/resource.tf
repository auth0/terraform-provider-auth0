resource "auth0_prompt_screen_renderer" "prompt_screen_renderer" {
  prompt_type                = "login-id"
  screen_name                = "login-id"
  rendering_mode             = "advanced"
  default_head_tags_disabled = false
  context_configuration = [
    "branding.settings",
    "branding.themes.default",
    "client.logo_uri",
    "client.description",
    "organization.display_name",
    "organization.branding",
    "screen.texts",
    "tenant.name",
    "tenant.friendly_name",
    "tenant.enabled_locales",
    "untrusted_data.submitted_form_data",
    "untrusted_data.authorization_params.ui_locales",
    "untrusted_data.authorization_params.login_hint",
    "untrusted_data.authorization_params.screen_hint"
  ]
  head_tags = jsonencode([
    {
      attributes : {
        "async" : true,
        "defer" : true,
        "integrity" : [
          "sha512-v2CJ7UaYy4JwqLDIrZUI/4hqeoQieOmAZNXBeQyjo21dadnwR+8ZaIJVT8EE2iyI61OV8e6M8PP2/4hpQINQ/g=="
        ],
        "src" : "https://cdnjs.cloudflare.com/ajax/libs/jquery/3.7.1/jquery.min.js"
      },
      tag : "script"
    }
  ])
}
