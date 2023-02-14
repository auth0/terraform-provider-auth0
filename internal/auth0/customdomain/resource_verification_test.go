package customdomain_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/provider"
	"github.com/auth0/terraform-provider-auth0/internal/recorder"
)

func TestAccCustomDomainVerificationWithAuth0ManagedCerts(t *testing.T) {
	if os.Getenv("AUTH0_DOMAIN") != recorder.RecordingsDomain {
		t.Skip()
	}

	httpRecorder := recorder.New(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: provider.TestFactories(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDomainVerificationWithAuth0ManagedCerts,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "domain", "terraform-provider.auth0.com"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "type", "auth0_managed_certs"),
					// The status attribute is set to "pending_verification"
					// here because Terraform has settled its state before
					// attempting the custom domain verification. We need to
					// refresh the state to move it along.
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "status", "pending_verification"),
					resource.TestCheckResourceAttrPair(
						"auth0_custom_domain.my_custom_domain", "id",
						"auth0_custom_domain_verification.my_custom_domain_verification", "custom_domain_id",
					),
					resource.TestCheckResourceAttrSet("auth0_custom_domain_verification.my_custom_domain_verification", "origin_domain_name"),
				),
			},
			{
				Config: testAccCustomDomainVerificationWithAuth0ManagedCerts,
				Check: resource.ComposeTestCheckFunc(
					// By applying an identical plan, we can reconcile the
					// status attribute.
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "status", "ready"),
					resource.TestCheckResourceAttrPair(
						"auth0_custom_domain.my_custom_domain", "origin_domain_name",
						"auth0_custom_domain_verification.my_custom_domain_verification", "origin_domain_name",
					),
				),
			},
		},
	})
}

const testAccCustomDomainVerificationWithAuth0ManagedCerts = `
resource "auth0_custom_domain" "my_custom_domain" {
	domain = "terraform-provider.auth0.com"
	type = "auth0_managed_certs"
}

resource "auth0_custom_domain_verification" "my_custom_domain_verification" {
	custom_domain_id = auth0_custom_domain.my_custom_domain.id
	timeouts { create = "15m" }
}
`

func TestAccCustomDomainVerificationWithSelfManagedCerts(t *testing.T) {
	if os.Getenv("AUTH0_DOMAIN") != recorder.RecordingsDomain {
		t.Skip()
	}

	httpRecorder := recorder.New(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: provider.TestFactories(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDomainVerificationWithSelfManagedCerts,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "domain", "terraform-provider.auth0.com"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "type", "self_managed_certs"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "status", "pending_verification"),
					resource.TestCheckResourceAttrPair(
						"auth0_custom_domain.my_custom_domain", "id",
						"auth0_custom_domain_verification.my_custom_domain_verification", "custom_domain_id",
					),
					resource.TestCheckResourceAttrSet("auth0_custom_domain_verification.my_custom_domain_verification", "origin_domain_name"),
					resource.TestCheckResourceAttrSet("auth0_custom_domain_verification.my_custom_domain_verification", "cname_api_key"),
				),
			},
			{
				Config: testAccCustomDomainVerificationWithSelfManagedCerts,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "status", "ready"),
					resource.TestCheckResourceAttrPair(
						"auth0_custom_domain.my_custom_domain", "origin_domain_name",
						"auth0_custom_domain_verification.my_custom_domain_verification", "origin_domain_name",
					),
					// Even though we can no longer read this from the API, it
					// should remain set after refresh as we won't clear it out
					// in the read operation.
					resource.TestCheckResourceAttrSet("auth0_custom_domain_verification.my_custom_domain_verification", "cname_api_key"),
				),
			},
		},
	})
}

const testAccCustomDomainVerificationWithSelfManagedCerts = `
resource "auth0_custom_domain" "my_custom_domain" {
	domain = "terraform-provider.auth0.com"
	type = "self_managed_certs"
}

resource "auth0_custom_domain_verification" "my_custom_domain_verification" {
	custom_domain_id = auth0_custom_domain.my_custom_domain.id
	timeouts { create = "15m" }
}
`
