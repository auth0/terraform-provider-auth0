package provider

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/assert"

	"github.com/auth0/terraform-provider-auth0/internal/recorder"
)

func init() {
	resource.AddTestSweepers("auth0_email", &resource.Sweeper{
		Name: "auth0_email",
		F: func(_ string) error {
			api, err := Auth0()
			if err != nil {
				return err
			}
			return api.Email.Delete()
		},
	})
}

func TestAccEmail(t *testing.T) {
	httpRecorder := recorder.New(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: `
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
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "name", "ses"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "enabled", "true"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "default_from_address", "accounts@example.com"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.0.access_key_id", "AKIAXXXXXXXXXXXXXXXX"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.0.secret_access_key", "7e8c2148xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.0.region", "us-east-1"),
				),
			},
			{
				Config: `
				resource "auth0_email" "my_email_provider" {
					name = "ses"
					enabled = true
					default_from_address = "accounts@example.com"
					credentials {
						access_key_id = "AKIAXXXXXXXXXXXXXXXY"
						secret_access_key = "7e8c2148xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
						region = "us-east-1"
					}
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "name", "ses"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "enabled", "true"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "default_from_address", "accounts@example.com"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.0.access_key_id", "AKIAXXXXXXXXXXXXXXXY"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.0.secret_access_key", "7e8c2148xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.0.region", "us-east-1"),
				),
			},
			{
				Config: `
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
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "name", "mailgun"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "enabled", "true"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "default_from_address", "accounts@example.com"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.0.domain", "example.com"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.0.region", "eu"),
				),
			},
			{
				Config: `
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
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "name", "mailgun"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "enabled", "false"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "default_from_address", ""),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.0.domain", "example.com"),
					resource.TestCheckResourceAttr("auth0_email.my_email_provider", "credentials.0.region", "eu"),
				),
			},
			{
				Config: `
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
				`,
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

func TestEmailProviderIsConfigured(t *testing.T) {
	t.Run("it returns true if the provider is configured", func(t *testing.T) {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v2/emails/provider" {
				w.WriteHeader(http.StatusOK)
				return
			}
			http.NotFound(w, r)
		})
		testServer := httptest.NewServer(testHandler)

		api, err := management.New(testServer.URL, management.WithInsecure())
		assert.NoError(t, err)

		actual := emailProviderIsConfigured(api)
		assert.True(t, actual)
	})

	t.Run("it returns false if the provider is not configured", func(t *testing.T) {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v2/emails/provider" {
				http.NotFound(w, r)
				return
			}
			http.NotFound(w, r)
		})
		testServer := httptest.NewServer(testHandler)

		api, err := management.New(testServer.URL, management.WithInsecure())
		assert.NoError(t, err)

		actual := emailProviderIsConfigured(api)
		assert.False(t, actual)
	})
}
