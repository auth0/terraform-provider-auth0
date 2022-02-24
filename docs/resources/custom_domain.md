---
layout: "auth0"
page_title: "Auth0: auth0_custom_domain"
description: |-
  With this resource, you can create and manage a custom domain within your Auth0 tenant in order to maintain a
  consistent user experience.
---

# auth0_custom_domain

With Auth0, you can use a custom domain to maintain a consistent user experience. This resource allows you to create and
manage a custom domain within your Auth0 tenant.

## Example Usage

```hcl
resource "auth0_custom_domain" "my_custom_domain" {
  domain = "auth.example.com"
  type = "auth0_managed_certs"
}
```

## Argument Reference

Arguments accepted by this resource include:

* `domain` - (Required) String. Name of the custom domain. 
* `type` - (Required) String. Provisioning type for the custom domain. Options include `auth0_managed_certs` and `self_managed_certs`.
* `verification_method` - (Deprecated) String. Domain verification method. The method is chosen according to the type of
the custom domain. `CNAME` for `auth0_managed_certs`, `TXT` for `self_managed_certs`.

## Attribute Reference

Attributes exported by this resource include:

* `primary` - Boolean. Indicates whether this is a primary domain.
* `status` - String. Configuration status for the custom domain. Options include `disabled`, `pending`, `pending_verification`, and `ready`.
* `verification` - List(Resource). Configuration settings for verification. For details, see [Verification](#verification).

### Verification

`verification` exports the following attributes:

* `methods` - List(Map). Verification methods for the domain.

## Import

Custom Domains can be imported using the id, e.g.

```
$ terraform import auth0_custom_domain.my_custom_domain cd_XXXXXXXXXXXXXXXX
```
