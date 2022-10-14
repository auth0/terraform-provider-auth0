package provider

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/recorder"
	"github.com/auth0/terraform-provider-auth0/internal/template"
)

func init() {
	resource.AddTestSweepers("auth0_custom_domain", &resource.Sweeper{
		Name: "auth0_custom_domain",
		F: func(_ string) error {
			api, err := Auth0()
			if err != nil {
				return err
			}

			domains, err := api.CustomDomain.List()
			if err != nil {
				return err
			}

			var result *multierror.Error
			for _, domain := range domains {
				log.Printf("[DEBUG] ➝ %s", domain.GetDomain())

				if strings.Contains(domain.GetDomain(), "auth.uat.terraform-provider-auth0.com") {
					result = multierror.Append(
						result,
						api.CustomDomain.Delete(domain.GetID()),
					)

					log.Printf("[DEBUG] ✗ %s", domain.GetDomain())
				}
			}

			return result.ErrorOrNil()
		},
	})
}

const testAccCreateSelfManagedCustomDomain = `
resource "auth0_custom_domain" "my_custom_domain" {
	domain = "{{.testName}}.auth.terraform-provider-auth0.com"
	type = "self_managed_certs"
}
`

const testAccUpdateSelfManagedCustomDomain = `
resource "auth0_custom_domain" "my_custom_domain" {
	domain = "{{.testName}}.auth.terraform-provider-auth0.com"
	type = "self_managed_certs"
	custom_client_ip_header = "true-client-ip"
}
`

const testAccUpdateSelfManagedCustomDomainWithEmptyClientIPHeader = `
resource "auth0_custom_domain" "my_custom_domain" {
	domain = "{{.testName}}.auth.terraform-provider-auth0.com"
	type = "self_managed_certs"
	custom_client_ip_header = ""
}
`

const testAccCreateAuth0ManagedCustomDomain = `
resource "auth0_custom_domain" "my_custom_domain" {
	domain = "{{.testName}}.auth.terraform-provider-auth0.com"
	type = "auth0_managed_certs"
}
`

const testAccUpdateAuth0ManagedCustomDomain = `
resource "auth0_custom_domain" "my_custom_domain" {
	domain = "{{.testName}}.auth.terraform-provider-auth0.com"
	type = "auth0_managed_certs"
	tls_policy = "recommended"
}
`

func TestAccCustomDomain(t *testing.T) {
	httpRecorder := recorder.New(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccCreateSelfManagedCustomDomain, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"auth0_custom_domain.my_custom_domain",
						"domain",
						fmt.Sprintf("%s.auth.terraform-provider-auth0.com", strings.ToLower(t.Name())),
					),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "type", "self_managed_certs"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "status", "pending_verification"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "origin_domain_name", ""),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "primary", "true"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "verification.#", "1"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "custom_client_ip_header", ""),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "tls_policy", ""),
				),
			},
			{
				Config: template.ParseTestName(testAccUpdateSelfManagedCustomDomain, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"auth0_custom_domain.my_custom_domain",
						"domain",
						fmt.Sprintf("%s.auth.terraform-provider-auth0.com", strings.ToLower(t.Name())),
					),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "type", "self_managed_certs"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "status", "pending_verification"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "origin_domain_name", ""),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "primary", "true"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "verification.#", "1"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "custom_client_ip_header", "true-client-ip"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "tls_policy", ""),
				),
			},
			{
				Config: template.ParseTestName(testAccUpdateSelfManagedCustomDomainWithEmptyClientIPHeader, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"auth0_custom_domain.my_custom_domain",
						"domain",
						fmt.Sprintf("%s.auth.terraform-provider-auth0.com", strings.ToLower(t.Name())),
					),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "type", "self_managed_certs"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "status", "pending_verification"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "origin_domain_name", ""),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "primary", "true"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "verification.#", "1"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "custom_client_ip_header", ""),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "tls_policy", ""),
				),
			},
			{
				Config: template.ParseTestName(testAccCreateAuth0ManagedCustomDomain, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"auth0_custom_domain.my_custom_domain",
						"domain",
						fmt.Sprintf("%s.auth.terraform-provider-auth0.com", strings.ToLower(t.Name())),
					),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "type", "auth0_managed_certs"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "status", "pending_verification"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "origin_domain_name", ""),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "primary", "true"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "verification.#", "1"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "custom_client_ip_header", ""),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "tls_policy", "recommended"),
				),
			},
			{
				Config: template.ParseTestName(testAccUpdateAuth0ManagedCustomDomain, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"auth0_custom_domain.my_custom_domain",
						"domain",
						fmt.Sprintf("%s.auth.terraform-provider-auth0.com", strings.ToLower(t.Name())),
					),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "type", "auth0_managed_certs"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "status", "pending_verification"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "origin_domain_name", ""),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "primary", "true"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "verification.#", "1"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "custom_client_ip_header", ""),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "tls_policy", "recommended"),
				),
			},
		},
	})
}
