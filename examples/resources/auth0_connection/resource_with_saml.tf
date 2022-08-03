# This is an example of a SAML connection.

resource "auth0_connection" "samlp" {
  name     = "SAML-Connection"
  strategy = "samlp"

  options {
    signing_cert        = "<signing-certificate>"
    sign_in_endpoint    = "https://saml.provider/sign_in"
    sign_out_endpoint   = "https://saml.provider/sign_out"
    disable_sign_out    = true
    tenant_domain       = "example.com"
    domain_aliases      = ["example.com", "alias.example.com"]
    protocol_binding    = "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
    request_template    = "<samlp:AuthnRequest xmlns:samlp=\"urn:oasis:names:tc:SAML:2.0:protocol\"\n@@AssertServiceURLAndDestination@@\n    ID=\"@@ID@@\"\n    IssueInstant=\"@@IssueInstant@@\"\n    ProtocolBinding=\"@@ProtocolBinding@@\" Version=\"2.0\">\n    <saml:Issuer xmlns:saml=\"urn:oasis:names:tc:SAML:2.0:assertion\">@@Issuer@@</saml:Issuer>\n</samlp:AuthnRequest>"
    user_id_attribute   = "https://saml.provider/imi/ns/identity-200810"
    signature_algorithm = "rsa-sha256"
    digest_algorithm    = "sha256"
    metadata_url        = "https://saml.provider/imi/ns/FederationMetadata.xml"
    fields_map = jsonencode({
      "name" : ["name", "nameidentifier"]
      "email" : ["emailaddress", "nameidentifier"]
      "family_name" : "surname"
    })

    signing_key {
      key  = "-----BEGIN PRIVATE KEY-----\n...{your private key here}...\n-----END PRIVATE KEY-----"
      cert = "-----BEGIN CERTIFICATE-----\n...{your public key cert here}...\n-----END CERTIFICATE-----"
    }
  }
}
