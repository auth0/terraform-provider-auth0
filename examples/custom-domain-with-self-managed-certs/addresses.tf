resource "google_compute_global_address" "lb_ipv4" {
  name       = "auth0-ipv4"
  ip_version = "IPV4"
}

resource "google_compute_global_address" "lb_ipv6" {
  name       = "auth0-ipv6"
  ip_version = "IPV6"
}

resource "google_dns_record_set" "lb_a" {
  name         = "${var.domain}."
  managed_zone = var.managed_zone_name
  type         = "A"
  ttl          = 300
  rrdatas = [
    google_compute_global_address.lb_ipv4.address,
  ]
}

resource "google_dns_record_set" "lb_aaaa" {
  name         = "${var.domain}."
  managed_zone = var.managed_zone_name
  type         = "AAAA"
  ttl          = 300
  rrdatas = [
    google_compute_global_address.lb_ipv6.address,
  ]
}
