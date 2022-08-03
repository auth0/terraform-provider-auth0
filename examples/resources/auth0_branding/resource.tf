resource "auth0_branding" "my_brand" {
  logo_url = "https://mycompany.org/logo.png"

  colors {
    primary         = "#0059d6"
    page_background = "#000000"
  }

  universal_login {
    # Ensure that "{%- auth0:head -%}" and "{%- auth0:widget -%}"
    # are present in the body.
    body = file("universal_login_body.html")
  }
}
