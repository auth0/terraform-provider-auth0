provider "auth0" {}

resource "auth0_branding" "my_brand" {
  logo_url = "https://mycompany.org/logo.png"

  colors {
    primary         = "#0059d6"
    page_background = "#000000"
  }

  universal_login {
    body = file("universal_login_body.html")
  }
}
