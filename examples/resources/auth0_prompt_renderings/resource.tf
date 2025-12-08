# Bulk management of prompt rendering settings for multiple screens
resource "auth0_prompt_renderings" "bulk_config" {
  renderings {
    prompt         = "login-passwordless"
    screen         = "login-passwordless-email-code"
    rendering_mode = "advanced"
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
      "tenant.enabled_locales"
    ]
    default_head_tags_disabled = false
    use_page_template          = false
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

  renderings {
    prompt         = "signup-id"
    screen         = "signup-id"
    rendering_mode = "standard"
    context_configuration = [
      "branding.settings",
      "screen.texts",
      "tenant.name"
    ]
  }

  renderings {
    prompt         = "login-id"
    screen         = "login-id"
    rendering_mode = "advanced"
    context_configuration = [
      "branding.settings",
      "branding.themes.default",
      "screen.texts"
    ]
    default_head_tags_disabled = true
    use_page_template          = true
  }
}

# Bulk configuration with filters for specific clients and organizations
resource "auth0_prompt_renderings\" \"filtered_bulk\" {
  renderings {
    prompt         = "login-passwordless"
    screen         = "login-passwordless-email-code"
    rendering_mode = "advanced"
    context_configuration = [
      "branding.settings",
      "client.logo_uri",
      "screen.texts"
    ]
    filters {
      match_type = "includes_any"
      clients = jsonencode([
        {
          id = "client_id_1"
        },
        {
          id = "client_id_2"
        }
      ])
      organizations = jsonencode([
        {
          metadata = {
            key   = "org_type"
            value = "enterprise"
          }
        }
      ])
      domains = jsonencode(["example.com", "company.com"])
    }
  }
}

# Minimal configuration with defaults
resource "auth0_prompt_renderings\" \"minimal\" {
  renderings {
    prompt = "login-id"
    screen = "login-id"
    # rendering_mode defaults to "standard"
    # other fields use their defaults
  }
}
