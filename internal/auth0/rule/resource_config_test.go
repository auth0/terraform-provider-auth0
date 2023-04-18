package rule_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccRuleConfigCreate = `
resource "auth0_rule_config" "foo" {
  key = "acc_test_{{.testName}}"
  value = "bar"
}
`

const testAccRuleConfigUpdateValue = `
resource "auth0_rule_config" "foo" {
  key = "acc_test_{{.testName}}"
  value = "foo"
}
`

const testAccRuleConfigUpdateKey = `
resource "auth0_rule_config" "foo" {
  key = "acc_test_key_{{.testName}}"
  value = "foo"
}
`

const testAccRuleConfigEmptyValue = `
resource "auth0_rule_config" "foo" {
  key = "acc_test_key_{{.testName}}"
  value = ""
}
`

func TestAccRuleConfig(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccRuleConfigCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_rule_config.foo", "id", fmt.Sprintf("acc_test_%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_rule_config.foo", "key", fmt.Sprintf("acc_test_%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_rule_config.foo", "value", "bar"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRuleConfigUpdateValue, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_rule_config.foo", "id", fmt.Sprintf("acc_test_%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_rule_config.foo", "key", fmt.Sprintf("acc_test_%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_rule_config.foo", "value", "foo"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRuleConfigUpdateKey, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_rule_config.foo", "id", fmt.Sprintf("acc_test_key_%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_rule_config.foo", "key", fmt.Sprintf("acc_test_key_%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_rule_config.foo", "value", "foo"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRuleConfigEmptyValue, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_rule_config.foo", "id", fmt.Sprintf("acc_test_key_%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_rule_config.foo", "key", fmt.Sprintf("acc_test_key_%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_rule_config.foo", "value", ""),
				),
			},
		},
	})
}
