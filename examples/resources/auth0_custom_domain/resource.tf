resource "auth0_custom_domain" "my_custom_domain" {
  domain     = "auth.example.com"
  type       = "auth0_managed_certs"
  tls_policy = "recommended"
  domain_metadata = {
    key1 : "value1"
    key2 : "value2"
  }
}
