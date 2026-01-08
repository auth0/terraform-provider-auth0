package prompt_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccPromptRenderingsBulkResource = `
resource "auth0_prompt_screen_renderers" "bulk_update" {
  renderings {
    prompt         = "login-passwordless"
    screen         = "login-passwordless-email-code"
    rendering_mode = "advanced"
    context_configuration = [
      "branding.settings",
      "branding.themes.default",
      "client.logo_uri",
      "client.description",
      "organization.display_name",
      "screen.texts",
      "tenant.name",
      "tenant.friendly_name"
    ]
    default_head_tags_disabled = false
    use_page_template          = false
    head_tags = jsonencode([
      {
        attributes : {
          "async" : true,
          "defer" : true,
          "src" : "https://cdnjs.cloudflare.com/ajax/libs/jquery/3.7.1/jquery.min.js"
        },
        tag : "script"
      }
    ])
  }

  renderings {
    prompt         = "signup-id"
    screen         = "signup-id"
    rendering_mode = "standard"
    context_configuration = [
      "branding.settings",
      "screen.texts",
      "tenant.name"
    ]
    default_head_tags_disabled = false
    use_page_template          = false
  }
}
`

const testAccPromptRenderingsBulkResourceUpdate = `
resource "auth0_prompt_screen_renderers" "bulk_update" {
  renderings {
    prompt         = "login-passwordless"
    screen         = "login-passwordless-email-code"
    rendering_mode = "advanced"
    context_configuration = [
      "branding.settings",
      "branding.themes.default",
      "client.logo_uri",
      "screen.texts",
      "tenant.name"
    ]
    default_head_tags_disabled = true
    use_page_template          = true
  }

  renderings {
    prompt         = "signup-id"
    screen         = "signup-id"
    rendering_mode = "advanced"
    context_configuration = [
      "branding.settings",
      "branding.themes.default",
      "screen.texts",
      "tenant.name"
    ]
    default_head_tags_disabled = false
    use_page_template          = false
  }

  renderings {
    prompt         = "login-id"
    screen         = "login-id"
    rendering_mode = "standard"
  }
}
`

func TestAccPromptRenderingsBulkResource(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccPromptRenderingsBulkResource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("auth0_prompt_screen_renderers.bulk_update", "id"),
					resource.TestMatchResourceAttr("auth0_prompt_screen_renderers.bulk_update", "id", regexp.MustCompile(`^prompt-renderings-bulk-`)),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.bulk_update", "renderings.#", "2"),

					// Check first rendering.
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.bulk_update", "renderings.0.prompt", "login-passwordless"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.bulk_update", "renderings.0.screen", "login-passwordless-email-code"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.bulk_update", "renderings.0.rendering_mode", "advanced"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.bulk_update", "renderings.0.context_configuration.#", "8"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.bulk_update", "renderings.0.default_head_tags_disabled", "false"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.bulk_update", "renderings.0.use_page_template", "false"),
					resource.TestCheckResourceAttrSet("auth0_prompt_screen_renderers.bulk_update", "renderings.0.head_tags"),
					// resource.TestCheckResourceAttrSet("auth0_prompt_screen_renderers.bulk_update", "renderings.0.tenant"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.bulk_update", "renderings.0.filters.#", "0"),

					// Check second rendering.
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.bulk_update", "renderings.1.prompt", "signup-id"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.bulk_update", "renderings.1.screen", "signup-id"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.bulk_update", "renderings.1.rendering_mode", "standard"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.bulk_update", "renderings.1.context_configuration.#", "3"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.bulk_update", "renderings.1.default_head_tags_disabled", "false"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.bulk_update", "renderings.1.use_page_template", "false"),
					// resource.TestCheckResourceAttrSet("auth0_prompt_screen_renderers.bulk_update", "renderings.1.tenant"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.bulk_update", "renderings.1.filters.#", "0"),
				),
			},
			{
				Config: testAccPromptRenderingsBulkResourceUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("auth0_prompt_screen_renderers.bulk_update", "id"),
					resource.TestMatchResourceAttr("auth0_prompt_screen_renderers.bulk_update", "id", regexp.MustCompile(`^prompt-renderings-bulk-`)),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.bulk_update", "renderings.#", "3"),

					// Check updated first rendering.
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.bulk_update", "renderings.0.prompt", "login-passwordless"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.bulk_update", "renderings.0.screen", "login-passwordless-email-code"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.bulk_update", "renderings.0.rendering_mode", "advanced"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.bulk_update", "renderings.0.context_configuration.#", "5"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.bulk_update", "renderings.0.default_head_tags_disabled", "true"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.bulk_update", "renderings.0.use_page_template", "true"),

					// Check updated second rendering.
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.bulk_update", "renderings.1.prompt", "signup-id"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.bulk_update", "renderings.1.screen", "signup-id"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.bulk_update", "renderings.1.rendering_mode", "advanced"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.bulk_update", "renderings.1.context_configuration.#", "4"),

					// Check new third rendering.
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.bulk_update", "renderings.2.prompt", "login-id"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.bulk_update", "renderings.2.screen", "login-id"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.bulk_update", "renderings.2.rendering_mode", "standard"),
					// resource.TestCheckResourceAttrSet("auth0_prompt_screen_renderers.bulk_update", "renderings.2.tenant"),
				),
			},
		},
	})
}

const testAccPromptRenderingsBulkResourceWithFilters = testClientCreate + testClientCreate2 + `
resource "auth0_prompt_screen_renderers" "with_filters" {
  renderings {
    prompt         = "login-passwordless"
    screen         = "login-passwordless-email-code"
    rendering_mode = "advanced"
    context_configuration = [
      "branding.settings",
      "screen.texts"
    ]
    filters {
      match_type = "includes_any"
      clients = jsonencode([
       {
		   id = auth0_client.my_client-1.id
		},
		{
		   id = auth0_client.my_client-2.id
		}
      ])
      organizations = jsonencode([
        {
          metadata = {
             some_key = "some_value"
          }
        }
      ])
    }
  }
}
`

const testAccPromptRenderingsBulkResourceWithFiltersUpdate = `
resource "auth0_prompt_screen_renderers" "with_filters" {
  renderings {
    prompt         = "login-passwordless"
    screen         = "login-passwordless-email-code"
    rendering_mode = "advanced"
    context_configuration = [
      "branding.settings",
      "screen.texts"
    ]
  }
}
`

func TestAccPromptRenderingsBulkResourceWithFilters(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccPromptRenderingsBulkResourceWithFilters,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("auth0_prompt_screen_renderers.with_filters", "id"),
					resource.TestMatchResourceAttr("auth0_prompt_screen_renderers.with_filters", "id", regexp.MustCompile(`^prompt-renderings-bulk-`)),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.with_filters", "renderings.#", "1"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.with_filters", "renderings.0.prompt", "login-passwordless"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.with_filters", "renderings.0.screen", "login-passwordless-email-code"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.with_filters", "renderings.0.rendering_mode", "advanced"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.with_filters", "renderings.0.context_configuration.#", "2"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.with_filters", "renderings.0.filters.#", "1"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.with_filters", "renderings.0.filters.0.match_type", "includes_any"),
					resource.TestCheckResourceAttrSet("auth0_prompt_screen_renderers.with_filters", "renderings.0.filters.0.clients"),
					resource.TestCheckResourceAttrSet("auth0_prompt_screen_renderers.with_filters", "renderings.0.filters.0.organizations"),
				),
			},
			{
				Config: testAccPromptRenderingsBulkResourceWithFiltersUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("auth0_prompt_screen_renderers.with_filters", "id"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.with_filters", "renderings.#", "1"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.with_filters", "renderings.0.prompt", "login-passwordless"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.with_filters", "renderings.0.rendering_mode", "advanced"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.with_filters", "renderings.0.filters.#", "0"),
				),
			},
		},
	})
}

const testAccPromptRenderingsBulkResourceMultipleWithFilters = testClientCreate + testClientCreate2 + `
resource "auth0_prompt_screen_renderers" "multiple_filters" {
  renderings {
    prompt         = "login-passwordless"
    screen         = "login-passwordless-email-code"
    rendering_mode = "advanced"
    context_configuration = [
      "branding.settings",
      "screen.texts"
    ]
    filters {
      match_type = "includes_any"
      clients = jsonencode([
        {
          id = auth0_client.my_client-1.id
        }
      ])
    }
  }

  renderings {
    prompt         = "signup-id"
    screen         = "signup-id"
    rendering_mode = "advanced"
    context_configuration = [
      "branding.settings",
      "tenant.name"
    ]
    filters {
      match_type = "includes_any"
      organizations = jsonencode([
        {
          metadata = {
            some_key = "some_value"
          }
        }
      ])
    }
  }
}
`

func TestAccPromptRenderingsBulkResourceMultipleWithFilters(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccPromptRenderingsBulkResourceMultipleWithFilters,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.multiple_filters", "renderings.#", "2"),

					// First rendering with client filters.
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.multiple_filters", "renderings.0.prompt", "login-passwordless"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.multiple_filters", "renderings.0.filters.#", "1"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.multiple_filters", "renderings.0.filters.0.match_type", "includes_any"),
					resource.TestCheckResourceAttrSet("auth0_prompt_screen_renderers.multiple_filters", "renderings.0.filters.0.clients"),

					// Second rendering with organization and domain filters.
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.multiple_filters", "renderings.1.prompt", "signup-id"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.multiple_filters", "renderings.1.filters.#", "1"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.multiple_filters", "renderings.1.filters.0.match_type", "includes_any"),
					resource.TestCheckResourceAttrSet("auth0_prompt_screen_renderers.multiple_filters", "renderings.1.filters.0.organizations"),
				),
			},
		},
	})
}

const testAccPromptRenderingsInvalidConfig = `
resource "auth0_prompt_screen_renderers" "invalid" {
  renderings {
    prompt         = "invalid-prompt-type"
    screen         = "login-id"
    rendering_mode = "advanced"
  }
}
`

const testAccPromptRenderingsInvalidScreenConfig = `
resource "auth0_prompt_screen_renderers" "invalid" {
  renderings {
    prompt         = "login-id"
    screen         = "invalid-screen-name"
    rendering_mode = "advanced"
  }
}
`

const testAccPromptRenderingsInvalidRenderingMode = `
resource "auth0_prompt_screen_renderers" "invalid" {
  renderings {
    prompt         = "login-id"
    screen         = "login-id"
    rendering_mode = "invalid-mode"
  }
}
`

func TestAccPromptRenderingsInvalidConfig(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      testAccPromptRenderingsInvalidConfig,
				ExpectError: regexp.MustCompile("expected renderings.0.prompt to be one of"),
			},
			{
				Config:      testAccPromptRenderingsInvalidScreenConfig,
				ExpectError: regexp.MustCompile("expected renderings.0.screen to be one of"),
			},
			{
				Config:      testAccPromptRenderingsInvalidRenderingMode,
				ExpectError: regexp.MustCompile("expected renderings.0.rendering_mode to be one of"),
			},
			{
				Config:      testAccPromptRenderingsEmptyConfig,
				ExpectError: regexp.MustCompile("Insufficient renderings blocks"),
			},
		},
	})
}

const testAccPromptRenderingsEmptyConfig = `
resource "auth0_prompt_screen_renderers" "empty" {
}
`

const testAccPromptRenderingsImportConfig = `
resource "auth0_prompt_screen_renderers" "import_test" {
  renderings {
    prompt         = "login-id"
    screen         = "login-id"
    rendering_mode = "advanced"
    context_configuration = [
      "branding.settings",
      "screen.texts"
    ]
  }
}
`

func TestAccPromptRenderingsImport(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccPromptRenderingsImportConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("auth0_prompt_screen_renderers.import_test", "id", regexp.MustCompile(`^prompt-renderings-bulk-`)),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.import_test", "renderings.#", "1"),
				),
			},
		},
	})
}

const testAccPromptRenderingsWithAllFields = `
resource "auth0_prompt_screen_renderers" "all_fields" {
  renderings {
    prompt         = "login-id"
    screen         = "login-id"
    rendering_mode = "advanced"
    context_configuration = [
      "branding.settings",
      "branding.themes.default",
      "client.logo_uri",
      "client.description",
      "organization.display_name",
      "organization.branding",
      "screen.texts",
      "tenant.name",
      "tenant.friendly_name",
      "tenant.enabled_locales"
    ]
    default_head_tags_disabled = true
    use_page_template          = true
    head_tags = jsonencode([
      {
        attributes : {
          "async" : true,
          "defer" : true,
          "integrity" : [
            "sha512-v2CJ7UaYy4JwqLDIrZUI/4hqeoQieOmAZNXBeQyjo21dadnwR+8ZaIJVT8EE2iyI61OV8e6M8PP2/4hpQINQ/g=="
          ],
          "src" : "https://cdnjs.cloudflare.com/ajax/libs/jquery/3.7.1/jquery.min.js"
        },
        tag : "script"
      },
      {
        attributes : {
          "href" : "https://cdn.example.com/styles.css",
          "rel" : "stylesheet"
        },
        tag : "link"
      }
    ])
  }
}
`

func TestAccPromptRenderingsWithAllFields(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccPromptRenderingsWithAllFields,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.all_fields", "renderings.#", "1"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.all_fields", "renderings.0.prompt", "login-id"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.all_fields", "renderings.0.screen", "login-id"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.all_fields", "renderings.0.rendering_mode", "advanced"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.all_fields", "renderings.0.context_configuration.#", "10"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.all_fields", "renderings.0.default_head_tags_disabled", "true"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderers.all_fields", "renderings.0.use_page_template", "true"),
					resource.TestCheckResourceAttrSet("auth0_prompt_screen_renderers.all_fields", "renderings.0.head_tags"),
					// resource.TestCheckResourceAttrSet("auth0_prompt_screen_renderers.all_fields", "renderings.0.tenant"),
				),
			},
		},
	})
}
