resource "google_compute_global_network_endpoint_group" "proxy" {
  name                  = "auth0-proxy"
  network_endpoint_type = "INTERNET_FQDN_PORT"
}

resource "google_compute_global_network_endpoint" "proxy" {
  global_network_endpoint_group = google_compute_global_network_endpoint_group.proxy.name

  fqdn = auth0_custom_domain_verification.my_domain.origin_domain_name
  port = 443
}

resource "google_compute_backend_service" "proxy" {
  name        = "auth0-proxy"
  description = "Auth0 authentication proxy"

  backend {
    group = google_compute_global_network_endpoint_group.proxy.self_link
  }

  protocol   = "HTTPS"
  enable_cdn = false

  log_config {
    enable = true
  }

  custom_request_headers = [
    "host: ${auth0_custom_domain_verification.my_domain.origin_domain_name}",
    "cname-api-key: ${auth0_custom_domain_verification.my_domain.cname_api_key}",
  ]
}

resource "google_compute_url_map" "proxy_https" {
  name        = "auth0-proxy-https"
  description = "HTTPS endpoint for the Auth0 authentication proxy"

  default_service = google_compute_backend_service.proxy.self_link
}

resource "google_compute_managed_ssl_certificate" "proxy" {
  name = "auth0-proxy-https"

  managed {
    domains = [var.domain]
  }
}

resource "google_compute_target_https_proxy" "proxy" {
  name    = "auth0-proxy-https"
  url_map = google_compute_url_map.proxy_https.self_link
  ssl_certificates = [
    google_compute_managed_ssl_certificate.proxy.self_link,
  ]
}

resource "google_compute_global_forwarding_rule" "proxy_https_ipv4" {
  name                  = "auth0-proxy-https-ipv4"
  load_balancing_scheme = "EXTERNAL"
  port_range            = "443"

  ip_address = google_compute_global_address.lb_ipv4.address
  target     = google_compute_target_https_proxy.proxy.self_link
}

resource "google_compute_global_forwarding_rule" "proxy_https_ipv6" {
  name                  = "auth0-proxy-https-ipv6"
  load_balancing_scheme = "EXTERNAL"
  port_range            = "443"

  ip_address = google_compute_global_address.lb_ipv6.address
  target     = google_compute_target_https_proxy.proxy.self_link
}
