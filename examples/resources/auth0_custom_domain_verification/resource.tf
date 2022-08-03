# Example of a custom domain managed through DigitalOcean and verified using this resource.

resource "auth0_custom_domain" "my_custom_domain" {
  domain = "login.example.com"
  type   = "auth0_managed_certs"
}

resource "auth0_custom_domain_verification" "my_custom_domain_verification" {
  depends_on = [digitalocean_record.my_domain_name_record]

  custom_domain_id = auth0_custom_domain.my_custom_domain.id

  timeouts { create = "15m" }
}

resource "digitalocean_record" "my_domain_name_record" {
  domain = "example.com"
  type   = upper(auth0_custom_domain.my_custom_domain.verification[0].methods[0].name)
  name   = "${auth0_custom_domain.my_custom_domain.domain}."
  value  = "${auth0_custom_domain.my_custom_domain.verification[0].methods[0].record}."
}
