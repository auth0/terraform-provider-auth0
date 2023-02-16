package rule

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandRule(d cty.Value) *management.Rule {
	return &management.Rule{
		Name:    value.String(d.GetAttr("name")),
		Script:  value.String(d.GetAttr("script")),
		Order:   value.Int(d.GetAttr("order")),
		Enabled: value.Bool(d.GetAttr("enabled")),
	}
}

func expandRuleConfig(d cty.Value) *management.RuleConfig {
	return &management.RuleConfig{
		Key:   value.String(d.GetAttr("key")),
		Value: value.String(d.GetAttr("value")),
	}
}
