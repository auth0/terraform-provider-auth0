package branding_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccDataSourceBrandingTheme = `
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
		base_focus_color = "#635dff"
		base_hover_color = "#000000"
		body_text = "#FF00CC"
		captcha_widget_theme = "auto"
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

data "auth0_branding_theme" "test" {
	depends_on = [ auth0_branding_theme.my_theme ]
}
`

func TestAccDataSourceBrandingTheme(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceBrandingTheme,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "borders.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "borders.0.button_border_radius", "1.1"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "borders.0.button_border_weight", "1.34"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "borders.0.buttons_style", "pill"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "borders.0.input_border_radius", "3.2"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "borders.0.input_border_weight", "1.99"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "borders.0.inputs_style", "pill"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "borders.0.show_widget_shadow", "false"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "borders.0.widget_border_weight", "1.11"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "borders.0.widget_corner_radius", "3.57"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "colors.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "colors.0.base_focus_color", "#635dff"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "colors.0.base_hover_color", "#000000"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "colors.0.body_text", "#FF00CC"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "colors.0.captcha_widget_theme", "auto"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "colors.0.error", "#FF00CC"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "colors.0.header", "#FF00CC"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "colors.0.icons", "#FF00CC"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "colors.0.input_background", "#FF00CC"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "colors.0.input_border", "#FF00CC"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "colors.0.input_filled_text", "#FF00CC"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "colors.0.input_labels_placeholders", "#FF00CC"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "colors.0.links_focused_components", "#FF00CC"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "colors.0.primary_button", "#FF00CC"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "colors.0.primary_button_label", "#FF00CC"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "colors.0.secondary_button_border", "#FF00CC"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "colors.0.secondary_button_label", "#FF00CC"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "colors.0.success", "#FF00CC"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "colors.0.widget_background", "#FF00CC"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "colors.0.widget_border", "#FF00CC"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "fonts.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "fonts.0.font_url", "https://google.com/font.woff"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "fonts.0.links_style", "normal"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "fonts.0.reference_text_size", "12.5"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "fonts.0.body_text.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "fonts.0.body_text.0.bold", "false"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "fonts.0.body_text.0.size", "99.5"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "fonts.0.buttons_text.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "fonts.0.buttons_text.0.bold", "false"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "fonts.0.buttons_text.0.size", "99.5"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "fonts.0.input_labels.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "fonts.0.input_labels.0.bold", "false"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "fonts.0.input_labels.0.size", "99.5"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "fonts.0.links.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "fonts.0.links.0.bold", "false"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "fonts.0.links.0.size", "99.5"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "fonts.0.title.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "fonts.0.title.0.bold", "false"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "fonts.0.title.0.size", "99.5"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "fonts.0.subtitle.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "fonts.0.subtitle.0.bold", "false"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "fonts.0.subtitle.0.size", "99.5"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "page_background.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "page_background.0.background_color", "#000000"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "page_background.0.background_image_url", "https://google.com/background.png"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "page_background.0.page_layout", "center"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "widget.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "widget.0.header_text_alignment", "center"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "widget.0.logo_height", "55.5"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "widget.0.logo_position", "center"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "widget.0.logo_url", "https://google.com/logo.png"),
					resource.TestCheckResourceAttr("data.auth0_branding_theme.test", "widget.0.social_buttons_layout", "top"),
				),
			},
		},
	})
}
