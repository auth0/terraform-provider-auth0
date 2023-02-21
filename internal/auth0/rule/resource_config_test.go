package rule_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
	"github.com/auth0/terraform-provider-auth0/internal/sweep"
	"github.com/auth0/terraform-provider-auth0/internal/template"
)

func init() {
	sweep.RuleConfigs()
}

// This is needed so that the test
// sweepers get registered.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

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
				Config: template.ParseTestName(testAccRuleConfigCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_rule_config.foo", "id", fmt.Sprintf("acc_test_%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_rule_config.foo", "key", fmt.Sprintf("acc_test_%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_rule_config.foo", "value", "bar"),
				),
			},
			{
				Config: template.ParseTestName(testAccRuleConfigUpdateValue, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_rule_config.foo", "id", fmt.Sprintf("acc_test_%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_rule_config.foo", "key", fmt.Sprintf("acc_test_%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_rule_config.foo", "value", "foo"),
				),
			},
			{
				Config: template.ParseTestName(testAccRuleConfigUpdateKey, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_rule_config.foo", "id", fmt.Sprintf("acc_test_key_%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_rule_config.foo", "key", fmt.Sprintf("acc_test_key_%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_rule_config.foo", "value", "foo"),
				),
			},
			{
				Config: template.ParseTestName(testAccRuleConfigEmptyValue, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_rule_config.foo", "id", fmt.Sprintf("acc_test_key_%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_rule_config.foo", "key", fmt.Sprintf("acc_test_key_%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_rule_config.foo", "value", ""),
				),
			},
		},
	})
}
