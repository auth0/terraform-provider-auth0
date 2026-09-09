# Create an HMAC-SHA256 key for Network ACL HTTP Message Signature verification.
# The key material (value) is write-only and not stored in Terraform state.
resource "auth0_network_acl_key" "my_key" {
  name  = "my-hmac-key"
  alg   = "hmac-sha256"
  value = var.hmac_key_base64 # Base64-encoded, 32–512 bytes when decoded
}

# Reference the key in a Network ACL rule to allow only signed requests.
resource "auth0_network_acl" "allow_signed_only" {
  description = "Block requests without a valid HMAC signature"
  active      = true
  priority    = 99
  rule {
    action {
      block = true
    }
    scope = "tenant"
    not_match {
      http_message_signature {
        keys {
          id = auth0_network_acl_key.my_key.id
        }
      }
    }
  }
}
