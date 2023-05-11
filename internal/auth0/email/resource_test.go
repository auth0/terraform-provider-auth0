package email_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccCreateSESEmailProvider = `
resource "auth0_email" "my_email_provider" {
	name = "ses"
	enabled = true
	default_from_address = "accounts@example.com"
	credentials {
		access_key_id = "AKIAXXXXXXXXXXXXXXXX"
		secret_access_key = "7e8c2148xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
		region = "us-east-1"
	}
}
`

const testAccUpdateSESEmailProvider = `
resource "auth0_email" "my_email_provider" {
	name = "ses"
	enabled = true
	default_from_address = "accounts@example.com"
	credentials {
		access_key_id = "AKIAXXXXXXXXXXXXXXXX"
		secret_access_key = "7e8c2148xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
		region = "us-east-1"
	}
	settings {
		message {
			configuration_set_name = "example"
		}
	}
}
`

const testAccCreateMandrillEmailProvider = `
resource "auth0_email" "my_email_provider" {
	name = "mandrill"
	enabled = true
	default_from_address = "accounts@example.com"
	credentials {
		api_key = "7e8c2148xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	}
}
`

const testAccUpdateMandrillEmailProvider = `
resource "auth0_email" "my_email_provider" {
	name = "mandrill"
	enabled = true
	default_from_address = "accounts@example.com"
	credentials {
		api_key = "7e8c2148xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	}
	settings {
		message {
			view_content_link = true
		}
	}
}
`

const testAccCreateSMTPEmailProvider = `
resource "auth0_email" "my_email_provider" {
	name = "smtp"
	enabled = true
	default_from_address = "accounts@example.com"
	credentials {
		smtp_host = "example.com"
		smtp_port = 984
		smtp_user = "bob"
		smtp_pass = "secret"
	}
}
`

const testAccUpdateSMTPEmailProvider = `
resource "auth0_email" "my_email_provider" {
	name = "smtp"
	enabled = true
	default_from_address = "accounts@example.com"
	credentials {
		smtp_host = "example.com"
		smtp_port = 984
		smtp_user = "bob"
		smtp_pass = "secret"
	}
	settings {
		headers {
			x_mc_view_content_link = "true"
			x_ses_configuration_set = "example"
		}
	}
}
`

const testAccCreateMailgunEmailProvider = `
resource "auth0_email" "my_email_provider" {
	name = "mailgun"
	enabled = true
	default_from_address = "accounts@example.com"
	credentials {
		api_key = "MAILGUNXXXXXXXXXXXXXXX"
		domain = "example.com"
		region = "eu"
	}
}
`

const testAccUpdateMailgunEmailProvider = `
resource "auth0_email" "my_email_provider" {
	name = "mailgun"
	enabled = false
	default_from_address = ""
	credentials {
		api_key = "MAILGUNXXXXXXXXXXXXXXX"
		domain = "example.com"
		region = "eu"
	}
}
`

const testAccAlreadyConfiguredEmailProviderWillNotConflict = `
resource "auth0_email" "my_email_provider" {
	name = "mailgun"
	enabled = false
	default_from_address = ""
	credentials {
		api_key = "MAILGUNXXXXXXXXXXXXXXX"
		domain = "example.com"
		region = "eu"
	}
}

resource "auth0_email" "no_conflict_email_provider" {
	depends_on = [ auth0_email.my_email_provider ]

	name = "mailgun"
	enabled = false
	default_from_address = ""
	credentials {
		api_key = "MAILGUNXXXXXXXXXXXXXXX"
		domain = "example.com"
		region = "eu"
	}
}
`

func TestAccEmail(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccCreateSESEmailProvider,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "name", "ses"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "enabled", "true"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "default_from_address", "accounts@example.com"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.#", "1"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.0.access_key_id", "AKIAXXXXXXXXXXXXXXXX"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.0.secret_access_key", "7e8c2148xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.0.region", "us-east-1"),
				),
			},
			{
				Config: testAccUpdateSESEmailProvider,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "name", "ses"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "enabled", "true"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "default_from_address", "accounts@example.com"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.#", "1"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.0.access_key_id", "AKIAXXXXXXXXXXXXXXXX"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.0.secret_access_key", "7e8c2148xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.0.region", "us-east-1"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "settings.#", "1"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "settings.0.message.#", "1"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "settings.0.message.0.configuration_set_name", "example"),
				),
			},
			{
				Config: testAccCreateMandrillEmailProvider,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "name", "mandrill"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "enabled", "true"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "default_from_address", "accounts@example.com"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.#", "1"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.0.api_key", "7e8c2148xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"),
				),
			},
			{
				Config: testAccUpdateMandrillEmailProvider,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "name", "mandrill"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "enabled", "true"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "default_from_address", "accounts@example.com"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.#", "1"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.0.api_key", "7e8c2148xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "settings.#", "1"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "settings.0.message.#", "1"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "settings.0.message.0.view_content_link", "true"),
				),
			},
			{
				Config: testAccCreateSMTPEmailProvider,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "name", "smtp"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "enabled", "true"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "default_from_address", "accounts@example.com"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.#", "1"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.0.smtp_host", "example.com"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.0.smtp_port", "984"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.0.smtp_user", "bob"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.0.smtp_pass", "secret"),
				),
			},
			{
				Config: testAccUpdateSMTPEmailProvider,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "name", "smtp"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "enabled", "true"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "default_from_address", "accounts@example.com"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.#", "1"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.0.smtp_host", "example.com"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.0.smtp_port", "984"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.0.smtp_user", "bob"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.0.smtp_pass", "secret"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "settings.#", "1"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "settings.0.headers.#", "1"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "settings.0.headers.0.x_mc_view_content_link", "true"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "settings.0.headers.0.x_ses_configuration_set", "example"),
				),
			},
			{
				Config: testAccCreateMailgunEmailProvider,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "name", "mailgun"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "enabled", "true"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "default_from_address", "accounts@example.com"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.0.domain", "example.com"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.0.region", "eu"),
				),
			},
			{
				Config: testAccUpdateMailgunEmailProvider,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "name", "mailgun"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "enabled", "false"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "default_from_address", ""),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.0.domain", "example.com"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.0.region", "eu"),
				),
			},
			{
				Config: testAccAlreadyConfiguredEmailProviderWillNotConflict,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "name", "mailgun"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "enabled", "false"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "default_from_address", ""),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.0.domain", "example.com"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.0.region", "eu"),
					resource.TestCheckResourceAttr("auth0_email.no_conflict_email_provider", "name", "mailgun"),
					resource.TestCheckResourceAttr("auth0_email.no_conflict_email_provider", "enabled", "false"),
					resource.TestCheckResourceAttr("auth0_email.no_conflict_email_provider", "default_from_address", ""),
					resource.TestCheckResourceAttr("auth0_email.no_conflict_email_provider", "credentials.0.domain", "example.com"),
					resource.TestCheckResourceAttr("auth0_email.no_conflict_email_provider", "credentials.0.region", "eu"),
				),
			},
		},
	})
}
