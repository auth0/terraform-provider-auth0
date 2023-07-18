package rule

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandRule(data cty.Value) *management.Rule {
	return &management.Rule{
		Name:    value.String(data.GetAttr("name")),
		Script:  value.String(data.GetAttr("script")),
		Order:   value.Int(data.GetAttr("order")),
		Enabled: value.Bool(data.GetAttr("enabled")),
	}
}

func expandRuleConfig(data cty.Value) *management.RuleConfig {
	return &management.RuleConfig{
		Key:   value.String(data.GetAttr("key")),
		Value: value.String(data.GetAttr("value")),
	}
}
