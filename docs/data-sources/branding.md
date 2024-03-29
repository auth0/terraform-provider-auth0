---
page_title: "Data Source: auth0_branding"
description: |-
  Use this data source to access information about the tenant's branding settings.
---

# Data Source: auth0_branding

Use this data source to access information about the tenant's branding settings.

## Example Usage

```terraform
data "auth0_branding" "my_branding" {}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Read-Only

- `colors` (List of Object) Configuration settings for colors for branding. (see [below for nested schema](#nestedatt--colors))
- `favicon_url` (String) URL for the favicon.
- `font` (List of Object) Configuration settings to customize the font. (see [below for nested schema](#nestedatt--font))
- `id` (String) The ID of this resource.
- `logo_url` (String) URL of logo for branding.
- `universal_login` (List of Object) Configuration settings for Universal Login. (see [below for nested schema](#nestedatt--universal_login))

<a id="nestedatt--colors"></a>
### Nested Schema for `colors`

Read-Only:

- `page_background` (String)
- `primary` (String)


<a id="nestedatt--font"></a>
### Nested Schema for `font`

Read-Only:

- `url` (String)


<a id="nestedatt--universal_login"></a>
### Nested Schema for `universal_login`

Read-Only:

- `body` (String)


