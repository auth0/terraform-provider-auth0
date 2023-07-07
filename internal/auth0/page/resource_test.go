package page_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccPagesCreate = `
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
`

const testAccPagesSetHTMLToEmptyAndDisabled = `
resource "auth0_pages" "my_pages" {
	login {
		enabled = false
		html    = ""
	}

	change_password {
		enabled = false
		html    = ""
	}

	guardian_mfa {
		enabled = false
		html    = ""
	}

	error {
		show_log_link = false
		html          = ""
		url           = ""
	}
}
`

const testAccPagesWithNoOptionalBlocksWillNotModifyPreExistingChanges = `
resource "auth0_pages" "my_pages" { }
`

const testAccPagesUpdateOnlyWithCustomLoginAndErrorPages = `
resource "auth0_pages" "my_pages" {
	login {
		enabled = true
		html    = "<html><body>My Custom Login Page</body></html>"
	}

	error {
		show_log_link = true
		html          = "<html><body>My Custom Error Page</body></html>"
		url           = "https://example.com"
	}
}
`

func TestAccPages(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccPagesCreate, t.Name()),
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
			{
				Config: acctest.ParseTestName(testAccPagesSetHTMLToEmptyAndDisabled, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "login.#", "1"),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "login.0.enabled", "false"),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "login.0.html", ""),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "change_password.#", "1"),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "change_password.0.enabled", "false"),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "change_password.0.html", ""),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "guardian_mfa.#", "1"),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "guardian_mfa.0.enabled", "false"),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "guardian_mfa.0.html", ""),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "error.#", "1"),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "error.0.show_log_link", "false"),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "error.0.html", ""),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "error.0.url", ""),
				),
			},
			{
				Config: acctest.ParseTestName(testAccPagesUpdateOnlyWithCustomLoginAndErrorPages, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "login.#", "1"),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "login.0.enabled", "true"),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "login.0.html", "<html><body>My Custom Login Page</body></html>"),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "change_password.#", "1"),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "change_password.0.enabled", "false"),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "change_password.0.html", ""),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "guardian_mfa.#", "1"),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "guardian_mfa.0.enabled", "false"),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "guardian_mfa.0.html", ""),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "error.#", "1"),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "error.0.show_log_link", "true"),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "error.0.html", "<html><body>My Custom Error Page</body></html>"),
					resource.TestCheckResourceAttr("auth0_pages.my_pages", "error.0.url", "https://example.com"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccPagesCreate, t.Name()),
			},
			{
				Config: acctest.ParseTestName(testAccPagesWithNoOptionalBlocksWillNotModifyPreExistingChanges, t.Name()),
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
