provider "auth0" {}

resource "auth0_branding" "my_brand" {
  logo_url = "https://mycompany.org/logo.png"
  colors {
    primary         = "#0059d6"
    page_background = "#000000"
  }
  universal_login {
    body = "<!DOCTYPE html><html><head>{%- auth0:head -%}</head><body>{%- auth0:widget -%}</body></html>"
  }
}
