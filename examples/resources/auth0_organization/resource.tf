resource "auth0_organization" "my_organization" {
  name         = "auth0-inc"
  display_name = "Auth0 Inc."

  branding {
    logo_url = "https://example.com/assets/icons/icon.png"
    colors = {
      primary         = "#f2f2f2"
      page_background = "#e1e1e1"
    }
  }
}
