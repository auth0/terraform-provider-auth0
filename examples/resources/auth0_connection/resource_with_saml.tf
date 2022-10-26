# This is an example of a SAML connection.

resource "auth0_connection" "samlp" {
  name     = "SAML-Connection"
  strategy = "samlp"

  options {
    debug               = false
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
    icon_url            = "https://saml.provider/assets/logo.png"
    entity_id           = "<entity_id>"
    metadata_xml        = <<EOF 
    <?xml version="1.0"?>
    <md:EntityDescriptor xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata" xmlns:ds="http://www.w3.org/2000/09/xmldsig#" entityID="https://example.com">
      <md:IDPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
        <md:SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="https://saml.provider/sign_out"/>
        <md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="https://saml.provider/sign_in"/>
      </md:IDPSSODescriptor>
    </md:EntityDescriptor>
    EOF
    metadata_url        = "https://saml.provider/imi/ns/FederationMetadata.xml" # Use either metadata_url or metadata_xml but not simultaneously
    fields_map = jsonencode({
      "name" : ["name", "nameidentifier"]
      "email" : ["emailaddress", "nameidentifier"]
      "family_name" : "surname"
    })
    signing_key {
      key  = "-----BEGIN PRIVATE KEY-----\n...{your private key here}...\n-----END PRIVATE KEY-----"
      cert = "-----BEGIN CERTIFICATE-----\n...{your public key cert here}...\n-----END CERTIFICATE-----"
    }
    idp_initiated {
      client_id              = "client_id"
      client_protocol        = "samlp"
      client_authorize_query = "type=code&timeout=30"
    }
  }
}
