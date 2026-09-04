---
page_title: "Data Source: auth0_network_acl_key"
description: |-
  Retrieve a Network ACL key by its ID. (EA Only)
---

# Data Source: auth0_network_acl_key

Retrieve a Network ACL key by its ID. The `value` field (key material) is not available via this data source.

> **Early Access (EA):** This data source is backed by the `/api/v2/keys/network-acls/*` endpoints, which are in Early Access. The API surface may change before GA.

## Example Usage

```terraform
data "auth0_network_acl_key" "my_key" {
  id = "key_<base58>"
}

output "key_fingerprint" {
  value = data.auth0_network_acl_key.my_key.fingerprint
}
```

## Argument Reference

- `id` (Required) - The ID of the Network ACL key to retrieve.

## Attributes Reference

- `name` - User-supplied label for the key.
- `alg` - Algorithm used for the key.
- `fingerprint` - SHA-256 fingerprint of the decoded key material (lowercase hex).
- `created_at` - Time when the key was created.
- `updated_at` - Time when the key was last updated.
