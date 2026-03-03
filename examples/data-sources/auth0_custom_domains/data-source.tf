resource "auth0_custom_domain" "my_custom_domain_1" {
  domain     = "example1.auth.tempdomain.com"
  type       = "auth0_managed_certs"
  tls_policy = "recommended"
  domain_metadata = {
    key1 : "foo1"
    key2 : "bar1"
  }
}

resource "auth0_custom_domain" "my_custom_domain_2" {
  domain     = "example2.auth.tempdomain.com"
  type       = "auth0_managed_certs"
  tls_policy = "recommended"
  domain_metadata = {
    key1 : "foo2"
    key2 : "bar2"
  }
}

data "auth0_custom_domains" "test" {
  q = "domain:example1* AND status:pending_verification"
}
