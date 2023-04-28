package rule

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

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
