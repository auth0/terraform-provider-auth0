resource "auth0_custom_domain" "my_custom_domain" {
  domain     = "{{.testName}}.auth.tempdomain.com"
  type       = "auth0_managed_certs"
  tls_policy = "recommended"
  domain_metadata = {
    key1 : "value1"
    key2 : "value2"
  }
}

data "auth0_custom_domain" "test" {
  custom_domain_id = auth0_custom_domain.my_custom_domain.id
}
