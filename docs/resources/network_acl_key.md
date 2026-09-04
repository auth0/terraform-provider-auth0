---
page_title: "Resource: auth0_network_acl_key"
description: |-
  Manages HMAC-SHA256 signing keys used to verify HTTP Message Signatures on Network ACL rules. (EA Only)
---

# Resource: auth0_network_acl_key

Manages HMAC-SHA256 signing keys used to verify HTTP Message Signatures on Network ACL rules.

> **Early Access (EA):** This resource is backed by the `/api/v2/keys/network-acls/*` endpoints, which are in Early Access. The API surface may change before GA.

## Example Usage

```terraform
resource "auth0_network_acl_key" "my_key" {
  name  = "my-hmac-key"
  alg   = "hmac-sha256"
  value = var.hmac_key_base64
}

resource "auth0_network_acl" "my_acl" {
  description = "Block unsigned requests"
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
```

## Argument Reference

- `name` (Required, ForceNew) - User-supplied label for the key. Must be unique across all Network ACL keys for the tenant. Max 255 characters.
- `alg` (Required, ForceNew) - Algorithm used for the key. Currently only `hmac-sha256` is supported.
- `value` (Required, Sensitive) - Base64-encoded raw key material. The decoded value must be between 32 and 512 bytes. This field is **write-only**: it is not returned by the API and not stored in Terraform state. Changes are detected by comparing SHA-256 fingerprints.

## Attributes Reference

- `id` - The ID of the Network ACL key (format: `key_<base58>`).
- `fingerprint` - SHA-256 fingerprint of the decoded key material (lowercase hex). Used for drift detection.
- `created_at` - Time when the key was created.
- `updated_at` - Time when the key was last updated.

## Write-Only `value`

The `value` field is write-only: Auth0 does not return key material in API responses. Terraform detects changes by comparing the SHA-256 fingerprint of the configured `value` against the fingerprint stored from the last apply.

If you rotate the key material in your secret store without updating the Terraform config, the next `terraform plan` will show the fingerprint has drifted and prompt for replacement.

## Tenant Key Cap

Auth0 allows at most **10 Network ACL keys per tenant**. Avoid setting `create_before_destroy = true` on this resource — the default `destroy-before-create` lifecycle is safe and avoids hitting the cap during replacement.

## Import

Import using the key ID:

```bash
terraform import auth0_network_acl_key.my_key key_<base58>
```

After import, the `value` field will be absent from state (the API does not return key material). Supply `value` in your config; if the fingerprint matches, the next `terraform plan` will be clean.

## Gateway Enforcement

Creating this resource configures HMAC signing keys for ACL rules. Whether requests are actually blocked depends on the gateway enforcement rollout for your tenant. Contact Auth0 support for enrollment status.
