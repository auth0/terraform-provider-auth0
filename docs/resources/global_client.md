---
layout: "auth0"
page_title: "Auth0: auth0_global_client"
description: |-
Use a tenant's global Auth0 Application client.
---

# auth0_global_client

Use a tenant's global Auth0 Application client.

## Example Usage

```hcl
resource "auth0_global_client" "global" {
    // Auth0 Universal Login - Custom Login Page
    custom_login_page_on = true
    custom_login_page    = <<PAGE
<html>
    <head><title>My Custom Login Page</title></head>
    <body>
        I should probably have a login form here
    </body>
</html>
PAGE
    callbacks            = ["http://somehostname.com/a/callback"]
}
```

## Argument Reference

Arguments accepted by this resource include the same ones as for an [auth0_client resource](client.md) with the
difference that all of them are optional.

## Attribute Reference

Attributes exported by this resource include the same ones as for an [auth0_client resource](client.md).

## Import

The auth0_global_client can be imported using the client's ID. You can find the ID of the global client by going to the
[API Explorer](https://auth0.com/docs/api/management/v2#!/Clients/get_clients) and fetching the clients that have
`"global": true`.

```
$ terraform import auth0_global_client.global XaiyAXXXYdXXXXnqjj8HXXXXXT5titww
```
