provider "auth0" {}

provider "google" {}

resource "auth0_custom_domain" "my_domain" {
  domain = var.domain
  type   = "self_managed_certs"
}

resource "google_dns_record_set" "my_domain_verification" {
  name         = "${auth0_custom_domain.my_domain.verification[0].methods[0].domain}."
  managed_zone = var.managed_zone_name
  type         = upper(auth0_custom_domain.my_domain.verification[0].methods[0].name)
  ttl          = 300
  rrdatas = [
    "${auth0_custom_domain.my_domain.verification[0].methods[0].record}.",
  ]
}

resource "auth0_custom_domain_verification" "my_domain" {
  custom_domain_id = auth0_custom_domain.my_domain.id

  depends_on = [
    google_dns_record_set.my_domain_verification,
  ]

  timeouts {
    create = "15m"
  }
}
