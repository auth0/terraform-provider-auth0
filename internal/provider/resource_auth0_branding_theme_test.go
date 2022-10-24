package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/recorder"
)

const testAccBrandingThemeCreate = `
resource "auth0_branding_theme" "my_theme" {
	borders {
		button_border_radius = 1.1
		button_border_weight = 1.34
		buttons_style = "pill"
		input_border_radius = 3.2
		input_border_weight = 1.99
		inputs_style = "pill"
		show_widget_shadow = false
		widget_border_weight = 1.11
		widget_corner_radius = 3.57
	}

	colors {
		body_text = "#FF00CC"
		error = "#FF00CC"
		header = "#FF00CC"
		icons = "#FF00CC"
		input_background = "#FF00CC"
		input_border = "#FF00CC"
		input_filled_text = "#FF00CC"
		input_labels_placeholders = "#FF00CC"
		links_focused_components = "#FF00CC"
		primary_button = "#FF00CC"
		primary_button_label = "#FF00CC"
		secondary_button_border = "#FF00CC"
		secondary_button_label = "#FF00CC"
		success = "#FF00CC"
		widget_background = "#FF00CC"
		widget_border = "#FF00CC"
	}

	fonts {
		font_url = "https://google.com/font.woff"
		links_style = "normal"
		reference_text_size = 12.5

		body_text {
			bold = false
			size = 99.5
		}

		buttons_text {
			bold = false
			size = 99.5
		}

		input_labels {
			bold = false
			size = 99.5
		}

		links {
			bold = false
			size = 99.5
		}

		title {
			bold = false
			size = 99.5
		}

		subtitle {
			bold = false
			size = 99.5
		}
	}

	page_background {
		background_color = "#000000"
		background_image_url = "https://google.com/background.png"
		page_layout = "center"
	}

	widget {
		header_text_alignment = "center"
		logo_height = 55.5
		logo_position = "center"
		logo_url = "https://google.com/logo.png"
		social_buttons_layout = "top"
	}
}
`

const testAccBrandingThemeUpdate = `
resource "auth0_branding_theme" "my_theme" {
	display_name = "My branding"

	borders {
		button_border_radius = 2
		button_border_weight = 2
		buttons_style = "pill"
		input_border_radius = 2
		input_border_weight = 2
		inputs_style = "pill"
		show_widget_shadow = true
		widget_border_weight = 2
		widget_corner_radius = 2
	}

	colors {
		base_focus_color = "#BBBBBB"
		base_hover_color = "#CCCCCC"

		body_text = "#00FF00"
		error = "#00FF00"
		header = "#00FF00"
		icons = "#00FF00"
		input_background = "#00FF00"
		input_border = "#00FF00"
		input_filled_text = "#00FF00"
		input_labels_placeholders = "#00FF00"
		links_focused_components = "#00FF00"
		primary_button = "#00FF00"
		primary_button_label = "#00FF00"
		secondary_button_border = "#00FF00"
		secondary_button_label = "#00FF00"
		success = "#00FF00"
		widget_background = "#00FF00"
		widget_border = "#00FF00"
	}

	fonts {
		font_url = "https://google.com/font.woff"
		links_style = "normal"
		reference_text_size = 12

		body_text {
			bold = true
			size = 100
		}

		buttons_text {
			bold = true
			size = 100
		}

		input_labels {
			bold = true
			size = 100
		}

		links {
			bold = true
			size = 100
		}

		title {
			bold = true
			size = 100
		}

		subtitle {
			bold = true
			size = 100
		}
	}

	page_background {
		background_color = "#000000"
		background_image_url = "https://google.com/background.png"
		page_layout = "center"
	}

	widget {
		header_text_alignment = "center"
		logo_height = 55
		logo_position = "center"
		logo_url = "https://google.com/logo.png"
		social_buttons_layout = "top"
	}
}
`

func TestAccBrandingTheme(t *testing.T) {
	httpRecorder := recorder.New(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: testAccBrandingThemeCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "borders.#", "1"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "borders.0.button_border_radius", "1.1"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "borders.0.button_border_weight", "1.34"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "borders.0.buttons_style", "pill"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "borders.0.input_border_radius", "3.2"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "borders.0.input_border_weight", "1.99"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "borders.0.inputs_style", "pill"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "borders.0.show_widget_shadow", "false"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "borders.0.widget_border_weight", "1.11"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "borders.0.widget_corner_radius", "3.57"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.#", "1"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.body_text", "#FF00CC"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.error", "#FF00CC"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.header", "#FF00CC"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.icons", "#FF00CC"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.input_background", "#FF00CC"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.input_border", "#FF00CC"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.input_filled_text", "#FF00CC"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.input_labels_placeholders", "#FF00CC"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.links_focused_components", "#FF00CC"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.primary_button", "#FF00CC"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.primary_button_label", "#FF00CC"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.secondary_button_border", "#FF00CC"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.secondary_button_label", "#FF00CC"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.success", "#FF00CC"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.widget_background", "#FF00CC"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.widget_border", "#FF00CC"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.#", "1"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.font_url", "https://google.com/font.woff"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.links_style", "normal"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.reference_text_size", "12.5"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.body_text.#", "1"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.body_text.0.bold", "false"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.body_text.0.size", "99.5"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.buttons_text.#", "1"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.buttons_text.0.bold", "false"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.buttons_text.0.size", "99.5"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.input_labels.#", "1"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.input_labels.0.bold", "false"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.input_labels.0.size", "99.5"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.links.#", "1"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.links.0.bold", "false"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.links.0.size", "99.5"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.title.#", "1"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.title.0.bold", "false"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.title.0.size", "99.5"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.subtitle.#", "1"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.subtitle.0.bold", "false"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.subtitle.0.size", "99.5"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "page_background.#", "1"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "page_background.0.background_color", "#000000"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "page_background.0.background_image_url", "https://google.com/background.png"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "page_background.0.page_layout", "center"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "widget.#", "1"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "widget.0.header_text_alignment", "center"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "widget.0.logo_height", "55.5"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "widget.0.logo_position", "center"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "widget.0.logo_url", "https://google.com/logo.png"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "widget.0.social_buttons_layout", "top"),
				),
			},
			{
				Config: testAccBrandingThemeUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "display_name", "My branding"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "borders.#", "1"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "borders.0.button_border_radius", "2"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "borders.0.button_border_weight", "2"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "borders.0.buttons_style", "pill"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "borders.0.input_border_radius", "2"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "borders.0.input_border_weight", "2"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "borders.0.inputs_style", "pill"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "borders.0.show_widget_shadow", "true"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "borders.0.widget_border_weight", "2"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "borders.0.widget_corner_radius", "2"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.#", "1"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.base_focus_color", "#BBBBBB"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.base_hover_color", "#CCCCCC"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.body_text", "#00FF00"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.error", "#00FF00"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.header", "#00FF00"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.icons", "#00FF00"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.input_background", "#00FF00"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.input_border", "#00FF00"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.input_filled_text", "#00FF00"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.input_labels_placeholders", "#00FF00"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.links_focused_components", "#00FF00"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.primary_button", "#00FF00"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.primary_button_label", "#00FF00"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.secondary_button_border", "#00FF00"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.secondary_button_label", "#00FF00"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.success", "#00FF00"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.widget_background", "#00FF00"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "colors.0.widget_border", "#00FF00"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.#", "1"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.font_url", "https://google.com/font.woff"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.links_style", "normal"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.body_text.#", "1"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.body_text.0.bold", "true"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.body_text.0.size", "100"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.buttons_text.#", "1"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.buttons_text.0.bold", "true"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.buttons_text.0.size", "100"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.input_labels.#", "1"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.input_labels.0.bold", "true"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.input_labels.0.size", "100"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.links.#", "1"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.links.0.bold", "true"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.links.0.size", "100"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.title.#", "1"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.title.0.bold", "true"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.title.0.size", "100"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.subtitle.#", "1"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.subtitle.0.bold", "true"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "fonts.0.subtitle.0.size", "100"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "page_background.#", "1"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "page_background.0.background_color", "#000000"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "page_background.0.background_image_url", "https://google.com/background.png"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "page_background.0.page_layout", "center"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "widget.#", "1"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "widget.0.header_text_alignment", "center"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "widget.0.logo_height", "55"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "widget.0.logo_position", "center"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "widget.0.logo_url", "https://google.com/logo.png"),
					resource.TestCheckResourceAttr("auth0_branding_theme.my_theme", "widget.0.social_buttons_layout", "top"),
				),
			},
		},
	})
}
