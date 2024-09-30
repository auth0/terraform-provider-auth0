package form_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testFormCreateInvalidStyle = `
resource "auth0_form" "my_form" {
    name = "test-form-{{.testName}}"
	style = "invalid-json"
	ending = jsonencode({resume_flow = true})
	languages { primary = "en" }
}
`

const testFormCreate = `
resource "auth0_form" "my_form" {
	name = "test-form-{{.testName}}"
	ending = jsonencode({resume_flow = true})
	languages { primary = "en" }
}
`

const testFormUpdate = `
resource "auth0_form" "my_form" {
	name = "updated-test-form-{{.testName}}"
	ending = jsonencode({resume_flow = true})
	languages { primary = "en" }
}
`

func TestAccForm(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      acctest.ParseTestName(testFormCreateInvalidStyle, t.Name()),
				ExpectError: regexp.MustCompile(`contains an invalid JSON`),
			},
			{
				Config: acctest.ParseTestName(testFormCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_form.my_form", "name", "test-form-"+t.Name()),
					resource.TestCheckResourceAttrSet("auth0_form.my_form", "id"),
				),
			},
			{
				Config: acctest.ParseTestName(testFormUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_form.my_form", "name", "updated-test-form-"+t.Name()),
					resource.TestCheckResourceAttrSet("auth0_form.my_form", "id"),
				),
			},
		},
	})
}
