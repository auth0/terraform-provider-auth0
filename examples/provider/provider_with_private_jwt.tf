provider "auth0" {
  domain                       = "<domain>"
  client_id                    = "<client-id>"
  client_assertion_private_key = file("<path-to-private-key>")
  client_assertion_signing_alg = "<signing-algorithm>"
  debug                        = "<debug>"
}
