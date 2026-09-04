package networkaclkey_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

// A 32-byte key encoded as standard base64 (256 bits — minimum valid length).
const testNetworkACLKeyValue = "dGhpcy1pcy1hLXZhbGlkLWtleS1mb3ItdGVzdGluZyE="

// A different valid key to exercise value-change → ForceNew.
const testNetworkACLKeyValueAlt = "YWx0ZXJuYXRlLXZhbGlkLWtleS1tYXRlcmlhbC0xISE="

const testAccNetworkACLKeyCreate = `
resource "auth0_network_acl_key" "my_key" {
	name  = "Acceptance Test Key - {{.testName}}"
	alg   = "hmac-sha256"
	value = "` + testNetworkACLKeyValue + `"
}
`

const testAccNetworkACLKeyUpdate = `
resource "auth0_network_acl_key" "my_key" {
	name  = "Acceptance Test Key - {{.testName}}"
	alg   = "hmac-sha256"
	value = "` + testNetworkACLKeyValueAlt + `"
}
`

const testAccNetworkACLKeyInvalidAlg = `
resource "auth0_network_acl_key" "my_key" {
	name  = "Bad Alg - {{.testName}}"
	alg   = "rsa-2048"
	value = "` + testNetworkACLKeyValue + `"
}
`

const testAccNetworkACLKeyInvalidValueShort = `
resource "auth0_network_acl_key" "my_key" {
	name  = "Short Value - {{.testName}}"
	alg   = "hmac-sha256"
	value = "dG9vc2hvcnQ="
}
`

func TestAccNetworkACLKey(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccNetworkACLKeyCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_network_acl_key.my_key", "name", fmt.Sprintf("Acceptance Test Key - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_network_acl_key.my_key", "alg", "hmac-sha256"),
					resource.TestCheckResourceAttrSet("auth0_network_acl_key.my_key", "id"),
					resource.TestCheckResourceAttrSet("auth0_network_acl_key.my_key", "fingerprint"),
					resource.TestCheckResourceAttrSet("auth0_network_acl_key.my_key", "created_at"),
					resource.TestCheckResourceAttrSet("auth0_network_acl_key.my_key", "updated_at"),
				),
			},
			{
				// Changing value triggers ForceNew (fingerprint mismatch).
				Config: acctest.ParseTestName(testAccNetworkACLKeyUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("auth0_network_acl_key.my_key", "id"),
					resource.TestCheckResourceAttrSet("auth0_network_acl_key.my_key", "fingerprint"),
				),
			},
			{
				// Verify that import leaves value absent from state.
				ResourceName:            "auth0_network_acl_key.my_key",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"value"},
			},
		},
	})
}

func TestAccNetworkACLKeyValidation(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      acctest.ParseTestName(testAccNetworkACLKeyInvalidAlg, t.Name()),
				ExpectError: regexp.MustCompile(`expected .* to be one of`),
			},
			{
				Config:      acctest.ParseTestName(testAccNetworkACLKeyInvalidValueShort, t.Name()),
				ExpectError: regexp.MustCompile(`decoded key material must be at least 32 bytes`),
			},
		},
	})
}
