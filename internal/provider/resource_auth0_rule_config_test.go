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
	resource.AddTestSweepers("auth0_rule_config", &resource.Sweeper{
		Name: "auth0_rule_config",
		F: func(_ string) error {
			api, err := Auth0()
			if err != nil {
				return err
			}

			configurations, err := api.RuleConfig.List()
			if err != nil {
				return err
			}

			var result *multierror.Error
			for _, c := range configurations {
				log.Printf("[DEBUG] ➝ %s", c.GetKey())
				if strings.Contains(c.GetKey(), "test") {
					result = multierror.Append(
						result,
						api.RuleConfig.Delete(c.GetKey()),
					)
					log.Printf("[DEBUG] ✗ %s", c.GetKey())
				}
			}

			return result.ErrorOrNil()
		},
	})
}

func TestAccRuleConfig(t *testing.T) {
	httpRecorder := recorder.New(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
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
		},
	})
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
