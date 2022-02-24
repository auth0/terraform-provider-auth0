---
layout: "auth0"
page_title: "Auth0: auth0_prompt"
description: |-
  With this resource, you can manage your Auth0 prompts, including choosing the login experience version.
---

# auth0_prompt

With this resource, you can manage your Auth0 prompts, including choosing the login experience version.

## Example Usage

```
resource "auth0_prompt" "example" {
  universal_login_experience = "classic"
  identifier_first           = false
}
```

## Argument Reference

The following arguments are supported:

- `universal_login_experience` - (Optional) Which login experience to use. Options include `classic` and `new`.
- `identifier_first` - (Optional) Boolean. Indicates whether the identifier first is used when using the new universal 
login experience.

## Attributes Reference

No additional attributes are exported by this resource.

## Import

As this is not a resource identifiable by an ID within the Auth0 Management API, prompt can be imported using a random
string. We recommend [Version 4 UUID](https://www.uuidgenerator.net/version4) e.g.

```
$ terraform import auth0_prompt.example 22f4f21b-017a-319d-92e7-2291c1ca36c4
```
