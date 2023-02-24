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
