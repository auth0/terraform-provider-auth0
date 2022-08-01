package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/template"
)

func TestAccRule(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccRule, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_rule.my_rule", "name", fmt.Sprintf("acceptance-test-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_rule.my_rule", "script", "function (user, context, callback) { callback(null, user, context); }"),
					resource.TestCheckResourceAttr("auth0_rule.my_rule", "enabled", "true"),
				),
			},
		},
	})
}

const testAccRule = `
resource "auth0_rule" "my_rule" {
  name = "acceptance-test-{{.testName}}"
  script = "function (user, context, callback) { callback(null, user, context); }"
  enabled = true
}
`

func TestRuleNameRegexp(t *testing.T) {
	vf := validation.StringMatch(ruleNameRegexp, "invalid name")

	for name, valid := range map[string]bool{
		"my-rule-1":                 true,
		"1-my-rule":                 true,
		"rule 2 name with spaces":   true,
		" rule with a space prefix": false,
		"rule with a space suffix ": false,
		" ":                         false,
		"   ":                       false,
	} {
		_, errs := vf(name, "name")
		if errs != nil && valid {
			t.Fatalf("Expected %q to be valid, but got validation errors %v", name, errs)
		}
		if errs == nil && !valid {
			t.Fatalf("Expected %q to be invalid, but got no validation errors.", name)
		}
	}
}
