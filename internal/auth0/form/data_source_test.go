package form_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAGivenForm = `
resource "auth0_form" "my_form" {
    name = "test-form-{{.testName}}"
	ending = jsonencode({resume_flow = true})
	languages { primary = "en" }
}
`

const testDataResourceWithoutID = testAGivenForm + `
data "auth0_form" "my_form" {
	depends_on = [resource.auth0_form.my_form]
}`

const testDataResourceWithInvalidID = testAGivenForm + `
data "auth0_form" "my_form" {
	depends_on = [resource.auth0_form.my_form]
	id = "ap_9AXxPP59pJx5ZtA471cSBx"
}`

const testDataResourceWithValidID = testAGivenForm + `
data "auth0_form" "my_form" {
	depends_on = [resource.auth0_form.my_form]
	id = resource.auth0_form.my_form.id
}`

func TestAccFormDataSource(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      acctest.ParseTestName(testDataResourceWithoutID, t.Name()),
				ExpectError: regexp.MustCompile("The argument \"id\" is required, but no definition was found."),
			},
			{
				Config: acctest.ParseTestName(testDataResourceWithInvalidID, t.Name()),
				ExpectError: regexp.MustCompile(
					`Error: 404 Not Found`,
				),
			},
			{
				Config: acctest.ParseTestName(testDataResourceWithValidID, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_form.my_form", "name"),
				),
			},
		},
	})
}
