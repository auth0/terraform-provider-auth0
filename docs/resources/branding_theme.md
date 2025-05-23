---
page_title: "Resource: auth0_branding_theme"
description: |-
  This resource allows you to manage branding themes for your Universal Login page within your Auth0 tenant.
---

# Resource: auth0_branding_theme

This resource allows you to manage branding themes for your Universal Login page within your Auth0 tenant.

## Example Usage

```terraform
# An example of the auth0_branding_theme using defaults for all the attributes.
resource "auth0_branding_theme" "my_theme" {
  borders {}
  colors {}
  fonts {
    title {}
    subtitle {}
    links {}
    input_labels {}
    buttons_text {}
    body_text {}
  }
  page_background {}
  widget {}
}

# An example of a fully configured auth0_branding_theme.
resource "auth0_branding_theme" "my_theme" {
  borders {
    button_border_radius = 1
    button_border_weight = 1
    buttons_style        = "pill"
    input_border_radius  = 3
    input_border_weight  = 1
    inputs_style         = "pill"
    show_widget_shadow   = false
    widget_border_weight = 1
    widget_corner_radius = 3
  }

  colors {
    body_text                 = "#FF00CC"
    error                     = "#FF00CC"
    header                    = "#FF00CC"
    icons                     = "#FF00CC"
    input_background          = "#FF00CC"
    input_border              = "#FF00CC"
    input_filled_text         = "#FF00CC"
    input_labels_placeholders = "#FF00CC"
    links_focused_components  = "#FF00CC"
    primary_button            = "#FF00CC"
    primary_button_label      = "#FF00CC"
    secondary_button_border   = "#FF00CC"
    secondary_button_label    = "#FF00CC"
    success                   = "#FF00CC"
    widget_background         = "#FF00CC"
    widget_border             = "#FF00CC"
  }

  fonts {
    font_url            = "https://google.com/font.woff"
    links_style         = "normal"
    reference_text_size = 12

    body_text {
      bold = false
      size = 100
    }

    buttons_text {
      bold = false
      size = 100
    }

    input_labels {
      bold = false
      size = 100
    }

    links {
      bold = false
      size = 100
    }

    title {
      bold = false
      size = 100
    }

    subtitle {
      bold = false
      size = 100
    }
  }

  page_background {
    background_color     = "#000000"
    background_image_url = "https://google.com/background.png"
    page_layout          = "center"
  }

  widget {
    header_text_alignment = "center"
    logo_height           = 55
    logo_position         = "center"
    logo_url              = "https://google.com/logo.png"
    social_buttons_layout = "top"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `borders` (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--borders))
- `colors` (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--colors))
- `fonts` (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--fonts))
- `page_background` (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--page_background))
- `widget` (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--widget))

### Optional

- `display_name` (String) The display name for the branding theme.

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--borders"></a>
### Nested Schema for `borders`

Optional:

- `button_border_radius` (Number) Button border radius. Value needs to be between `1` and `10`. Defaults to `3.0`.
- `button_border_weight` (Number) Button border weight. Value needs to be between `0` and `10`. Defaults to `1.0`.
- `buttons_style` (String) Buttons style. Available options: `pill`, `rounded`, `sharp`. Defaults to `rounded`.
- `input_border_radius` (Number) Input border radius. Value needs to be between `0` and `10`. Defaults to `3.0`.
- `input_border_weight` (Number) Input border weight. Value needs to be between `0` and `3`. Defaults to `1.0`.
- `inputs_style` (String) Inputs style. Available options: `pill`, `rounded`, `sharp`. Defaults to `rounded`.
- `show_widget_shadow` (Boolean) Show widget shadow. Defaults to `true`.
- `widget_border_weight` (Number) Widget border weight. Value needs to be between `0` and `10`. Defaults to `0.0`.
- `widget_corner_radius` (Number) Widget corner radius. Value needs to be between `0` and `50`. Defaults to `5.0`.


<a id="nestedblock--colors"></a>
### Nested Schema for `colors`

Optional:

- `base_focus_color` (String) Base focus color. Defaults to `#635dff`.
- `base_hover_color` (String) Base hover color. Defaults to `#000000`.
- `body_text` (String) Body text. Defaults to `#1e212a`.
- `captcha_widget_theme` (String) Captcha Widget Theme.
- `error` (String) Error. Defaults to `#d03c38`.
- `header` (String) Header. Defaults to `#1e212a`.
- `icons` (String) Icons. Defaults to `#65676e`.
- `input_background` (String) Input background. Defaults to `#ffffff`.
- `input_border` (String) Input border. Defaults to `#c9cace`.
- `input_filled_text` (String) Input filled text. Defaults to `#000000`.
- `input_labels_placeholders` (String) Input labels & placeholders. Defaults to `#65676e`.
- `links_focused_components` (String) Links & focused components. Defaults to `#635dff`.
- `primary_button` (String) Primary button. Defaults to `#635dff`.
- `primary_button_label` (String) Primary button label. Defaults to `#ffffff`.
- `secondary_button_border` (String) Secondary button border. Defaults to `#c9cace`.
- `secondary_button_label` (String) Secondary button label. Defaults to `#1e212a`.
- `success` (String) Success. Defaults to `#13a688`.
- `widget_background` (String) Widget background. Defaults to `#ffffff`.
- `widget_border` (String) Widget border. Defaults to `#c9cace`.


<a id="nestedblock--fonts"></a>
### Nested Schema for `fonts`

Required:

- `body_text` (Block List, Min: 1, Max: 1) Body text. (see [below for nested schema](#nestedblock--fonts--body_text))
- `buttons_text` (Block List, Min: 1, Max: 1) Buttons text. (see [below for nested schema](#nestedblock--fonts--buttons_text))
- `input_labels` (Block List, Min: 1, Max: 1) Input labels. (see [below for nested schema](#nestedblock--fonts--input_labels))
- `links` (Block List, Min: 1, Max: 1) Links. (see [below for nested schema](#nestedblock--fonts--links))
- `subtitle` (Block List, Min: 1, Max: 1) Subtitle. (see [below for nested schema](#nestedblock--fonts--subtitle))
- `title` (Block List, Min: 1, Max: 1) Title. (see [below for nested schema](#nestedblock--fonts--title))

Optional:

- `font_url` (String) Font URL. Defaults to an empty string.
- `links_style` (String) Links style. Defaults to `normal`.
- `reference_text_size` (Number) Reference text size. Value needs to be between `12` and `24`. Defaults to `16.0`.

<a id="nestedblock--fonts--body_text"></a>
### Nested Schema for `fonts.body_text`

Optional:

- `bold` (Boolean) Body text bold. Defaults to `false`.
- `size` (Number) Body text size. Value needs to be between `0` and `150`. Defaults to `87.5`.


<a id="nestedblock--fonts--buttons_text"></a>
### Nested Schema for `fonts.buttons_text`

Optional:

- `bold` (Boolean) Buttons text bold. Defaults to `false`.
- `size` (Number) Buttons text size. Value needs to be between `0` and `150`. Defaults to `100.0`.


<a id="nestedblock--fonts--input_labels"></a>
### Nested Schema for `fonts.input_labels`

Optional:

- `bold` (Boolean) Input labels bold. Defaults to `false`.
- `size` (Number) Input labels size. Value needs to be between `0` and `150`. Defaults to `100.0`.


<a id="nestedblock--fonts--links"></a>
### Nested Schema for `fonts.links`

Optional:

- `bold` (Boolean) Links bold. Defaults to `true`.
- `size` (Number) Links size. Value needs to be between `0` and `150`. Defaults to `87.5`.


<a id="nestedblock--fonts--subtitle"></a>
### Nested Schema for `fonts.subtitle`

Optional:

- `bold` (Boolean) Subtitle bold. Defaults to `false`.
- `size` (Number) Subtitle size. Value needs to be between `0` and `150`. Defaults to `87.5`.


<a id="nestedblock--fonts--title"></a>
### Nested Schema for `fonts.title`

Optional:

- `bold` (Boolean) Title bold. Defaults to `false`.
- `size` (Number) Title size. Value needs to be between `75` and `150`. Defaults to `150.0`.



<a id="nestedblock--page_background"></a>
### Nested Schema for `page_background`

Optional:

- `background_color` (String) Background color. Defaults to `#000000`.
- `background_image_url` (String) Background image url. Defaults to an empty string.
- `page_layout` (String) Page layout. Available options: `center`, `left`, `right`. Defaults to `center`.


<a id="nestedblock--widget"></a>
### Nested Schema for `widget`

Optional:

- `header_text_alignment` (String) Header text alignment. Available options: `center`, `left`, `right`. Defaults to `center`.
- `logo_height` (Number) Logo height. Value needs to be between `1` and `100`. Defaults to `52.0`.
- `logo_position` (String) Logo position. Available options: `center`, `left`, `right`, `none`. Defaults to `center`.
- `logo_url` (String) Logo url. Defaults to an empty string.
- `social_buttons_layout` (String) Social buttons layout. Available options: `bottom`, `top`. Defaults to `bottom`.

## Import

Import is supported using the following syntax:

```shell
# This resource can be imported by specifying the Branding Theme ID.
#
# Example:
terraform import auth0_branding_theme.my_theme "XXXXXXXXXXXXXXXXXXXX"
```
