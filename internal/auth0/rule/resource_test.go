package rule_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

func TestAccRule(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccRuleCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_rule.my_rule", "name", fmt.Sprintf("acceptance-test-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_rule.my_rule", "script", "function (user, context, callback) { callback(null, user, context); }"),
					resource.TestCheckResourceAttrSet("auth0_rule.my_rule", "enabled"),
					resource.TestCheckResourceAttrSet("auth0_rule.my_rule", "order"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRuleUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_rule.my_rule", "name", fmt.Sprintf("acceptance-test-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_rule.my_rule", "script", "function (user, context, callback) { console.log(\"here!\"); callback(null, user, context); }"),
					resource.TestCheckResourceAttr("auth0_rule.my_rule", "enabled", "true"),
					resource.TestCheckResourceAttr("auth0_rule.my_rule", "order", "1"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRuleUpdateAgain, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_rule.my_rule", "name", fmt.Sprintf("acceptance-test-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_rule.my_rule", "script", "function (user, context, callback) { console.log(\"here!\"); callback(null, user, context); }"),
					resource.TestCheckResourceAttr("auth0_rule.my_rule", "enabled", "false"),
					resource.TestCheckResourceAttr("auth0_rule.my_rule", "order", "1"),
				),
			},
		},
	})
}

const testAccRuleCreate = `
resource "auth0_rule" "my_rule" {
  name = "acceptance-test-{{.testName}}"
  script = "function (user, context, callback) { callback(null, user, context); }"
}
`

const testAccRuleUpdate = `
resource "auth0_rule" "my_rule" {
  name = "acceptance-test-{{.testName}}"
  script = "function (user, context, callback) { console.log(\"here!\"); callback(null, user, context); }"
  order = 1
  enabled = true
}
`

const testAccRuleUpdateAgain = `
resource "auth0_rule" "my_rule" {
  name = "acceptance-test-{{.testName}}"
  script = "function (user, context, callback) { console.log(\"here!\"); callback(null, user, context); }"
  enabled = false
}
`
