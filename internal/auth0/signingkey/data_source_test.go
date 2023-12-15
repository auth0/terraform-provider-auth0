package signingkey_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccDataSourceSigningKeys = `
data "auth0_signing_keys" "my_keys" { }
`

func TestAccDataSourceSigningKeys(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSigningKeys,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_signing_keys.my_keys", "signing_keys.#", "2"),
					resource.TestCheckResourceAttrSet("data.auth0_signing_keys.my_keys", "signing_keys.0.kid"),
					resource.TestCheckResourceAttrSet("data.auth0_signing_keys.my_keys", "signing_keys.0.cert"),
					resource.TestCheckResourceAttrSet("data.auth0_signing_keys.my_keys", "signing_keys.0.pkcs7"),
					resource.TestCheckResourceAttrSet("data.auth0_signing_keys.my_keys", "signing_keys.0.current"),
					resource.TestCheckResourceAttrSet("data.auth0_signing_keys.my_keys", "signing_keys.0.next"),
					resource.TestCheckResourceAttrSet("data.auth0_signing_keys.my_keys", "signing_keys.0.previous"),
					resource.TestCheckResourceAttrSet("data.auth0_signing_keys.my_keys", "signing_keys.0.revoked"),
					resource.TestCheckResourceAttrSet("data.auth0_signing_keys.my_keys", "signing_keys.0.fingerprint"),
					resource.TestCheckResourceAttrSet("data.auth0_signing_keys.my_keys", "signing_keys.0.thumbprint"),
					resource.TestCheckResourceAttrSet("data.auth0_signing_keys.my_keys", "signing_keys.1.kid"),
					resource.TestCheckResourceAttrSet("data.auth0_signing_keys.my_keys", "signing_keys.1.cert"),
					resource.TestCheckResourceAttrSet("data.auth0_signing_keys.my_keys", "signing_keys.1.pkcs7"),
					resource.TestCheckResourceAttrSet("data.auth0_signing_keys.my_keys", "signing_keys.1.current"),
					resource.TestCheckResourceAttrSet("data.auth0_signing_keys.my_keys", "signing_keys.1.next"),
					resource.TestCheckResourceAttrSet("data.auth0_signing_keys.my_keys", "signing_keys.1.previous"),
					resource.TestCheckResourceAttrSet("data.auth0_signing_keys.my_keys", "signing_keys.1.revoked"),
					resource.TestCheckResourceAttrSet("data.auth0_signing_keys.my_keys", "signing_keys.1.fingerprint"),
					resource.TestCheckResourceAttrSet("data.auth0_signing_keys.my_keys", "signing_keys.1.thumbprint"),
				),
			},
		},
	})
}
