package page_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccDataPagesConfig = `
resource "auth0_pages" "my_pages" {
	login {
		enabled = true
		html    = "<html><body>My Custom Login Page</body></html>"
	}

	change_password {
		enabled = true
		html    = "<html><body>My Custom Reset Password Page</body></html>"
	}

	guardian_mfa {
		enabled = true
		html    = "<html><body>My Custom MFA Page</body></html>"
	}

	error {
		show_log_link = true
		html          = "<html><body>My Custom Error Page</body></html>"
		url           = "https://example.com"
	}
}

data "auth0_pages" "my_pages" {
	depends_on = [ auth0_pages.my_pages ]
}
`

func TestAccDataSourcePages(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDataPagesConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "login.#", "1"),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "login.0.enabled", "true"),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "login.0.html", "<html><body>My Custom Login Page</body></html>"),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "change_password.#", "1"),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "change_password.0.enabled", "true"),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "change_password.0.html", "<html><body>My Custom Reset Password Page</body></html>"),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "guardian_mfa.#", "1"),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "guardian_mfa.0.enabled", "true"),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "guardian_mfa.0.html", "<html><body>My Custom MFA Page</body></html>"),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "error.#", "1"),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "error.0.show_log_link", "true"),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "error.0.html", "<html><body>My Custom Error Page</body></html>"),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "error.0.url", "https://example.com"),
				),
			},
		},
	})
}
