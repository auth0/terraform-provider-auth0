resource "auth0_custom_domain" "my_custom_domain" {
  domain = "auth.example.com"
  type   = "auth0_managed_certs"
}

resource "auth0_custom_domain_default" "default" {
  domain = auth0_custom_domain.my_custom_domain.domain
}

