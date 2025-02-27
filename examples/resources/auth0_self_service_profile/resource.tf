resource "auth0_self_service_profile" "my_self_service_profile" {
  user_attributes {
    name        = "sample-name"
    description = "sample-description"
    is_optional = true
  }
  branding {
    logo_url = "https://mycompany.org/v2/logo.png"
    colors {
      primary = "#0059d6"
    }
  }
}

