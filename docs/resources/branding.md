---
layout: "auth0"
page_title: "Auth0: auth0_branding"
description: |-
  With this resource, you can manage branding, including logo, color.
---

# auth0_branding

This resource allows you to manage branding within your Auth0 tenant.

Auth0 can be customized with a look and feel that aligns with your organization's brand requirements and user
expectations. 


## Example Usage

```hcl
resource "auth0_branding" "my_brand" {
	logo_url = "https://mycompany.org/logo.png"
	colors {
		primary = "#0059d6"
		page_background = "#000000"
	}
	universal_login {
		body = "<!DOCTYPE html><html><head>{%- auth0:head -%}</head><body>{%- auth0:widget -%}</body></html>"
	}
}
```

## Argument Reference

The following arguments are supported:

* `colors` - (Optional) List(Resource). Configuration settings for colors for branding. See [Colors](#colors).
* `favicon_url` - (Optional) String. URL for the favicon.
* `logo_url` - (Optional) String. URL of logo for branding.
* `font` - (Optional) List(Resource). Configuration settings to customize the font. See [Font](#font).
* `universal_login` - (Optional) List(Resource). Configuration settings for Universal Login.
See [Universal Login](#universal-login). This capability can only be used if the tenant has
[Custom Domains](https://auth0.com/docs/custom-domains) enabled.

### `Colors`

`colors` supports the following arguments:

* `page_background` - (Optional) String, Hexadecimal. Background color of login pages.
* `primary` - (Optional) String, Hexadecimal. Primary button background color.

### `Font`

`font` supports the following arguments:

* `url` - (Required) String. URL for the custom font.

### `Universal Login`

`universal_login` supports the following arguments:

* `body` - (Optional) String, body of login pages.

## Attributes Reference

No additional attributes are exported by this resource.

## Import 

As this is not a resource identifiable by an ID within the Auth0 Management API, branding can be imported using a random 
string. We recommend [Version 4 UUID](https://www.uuidgenerator.net/version4) e.g.

```
$ terraform import auth0_branding.my_brand 22f4f21b-017a-319d-92e7-2291c1ca36c4
```
