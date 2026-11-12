# Early Access only. The Management API client needs the create, read, update and delete
# `portals` scopes.

# The portal signs users in with a regular web application.
resource "auth0_client" "portal" {
  name        = "My Portal Client"
  app_type    = "regular_web"
  callbacks   = ["https://{yourDomain}/portals/auth/callback"]
  grant_types = ["authorization_code", "refresh_token"]
}

resource "auth0_client_credentials" "portal" {
  client_id             = auth0_client.portal.id
  authentication_method = "client_secret_post"
}

# Full example.
resource "auth0_portal" "full" {
  name = "My Portal"
  slug = "my-portal"

  client {
    client_secret_post {
      client_id = auth0_client.portal.client_id

      # The write-only secret is not stored in state. Increment the version to send a new secret.
      # NOTE: Write-only arguments are supported in Terraform 1.11 and later.
      client_secret_wo         = auth0_client_credentials.portal.client_secret
      client_secret_wo_version = 1
    }
  }

  custom_code {
    head = "<style>body { font-family: sans-serif; }</style>"
  }

  navigation {
    sidebar {
      component {
        internal_link_v2 {
          label     = "Account"
          page_slug = "account"
          icon      = "user"
        }
      }
      component {
        external_link_v2 {
          label = "Help"
          href  = "https://help.example.com"
          icon  = "circle-help"
        }
      }
    }
  }

  pages {
    default = "account"

    standard {
      title = "Account"
      slug  = "account"

      component {
        typography_heading {
          title       = "Your account"
          description = "Manage how you sign in."
        }
      }
      component {
        media_image {
          src      = "https://example.com/banner.png"
          position = "center"
          height   = 120
        }
      }
      component {
        my_account_mfa_management {}
      }
      # A section groups up to 20 components. Sections cannot be nested.
      component {
        structure_section {
          variant     = "card"
          title       = "Profile"
          description = "Update your details."

          child {
            form {
              form_id            = "ap_XXXXXXXXXXXXXXXXXXXXXX"
              completion_message = "Your profile is saved."
            }
          }
        }
      }
    }

    standard {
      title = "Organization"
      slug  = "organization"

      component {
        my_organization_member_management {}
      }
    }
  }
}
