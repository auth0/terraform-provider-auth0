package experimentsegment

import (
	"strings"
	"testing"

	"github.com/hashicorp/go-cty/cty"
)

func TestSegmentMetadataOperatorRequired(t *testing.T) {
	contains := cty.ListVal([]cty.Value{cty.StringVal("gold")})
	empty := cty.ListValEmpty(cty.String)

	for _, test := range []struct {
		name      string
		overrides map[string]cty.Value
		valid     bool
	}{
		{name: "missing operator"},
		{name: "empty list only", overrides: map[string]cty.Value{"one_of": empty}},
		{name: "two lists", overrides: map[string]cty.Value{"contains": contains, "one_of": contains}},
		{name: "list and exists true", overrides: map[string]cty.Value{"contains": contains, "exists": cty.StringVal("true")}},
		{name: "list and exists false", overrides: map[string]cty.Value{"contains": contains, "exists": cty.StringVal("false")}},
		{name: "contains only", overrides: map[string]cty.Value{"contains": contains}, valid: true},
		{name: "exists false only", overrides: map[string]cty.Value{"exists": cty.StringVal("false")}, valid: true},
		{
			name: "empty lists and contains",
			overrides: map[string]cty.Value{
				"one_of": empty, "contains": contains, "starts_with": empty, "ends_with": empty,
			},
			valid: true,
		},
		{
			name: "generated exists shape",
			overrides: map[string]cty.Value{
				"one_of": empty, "contains": empty, "starts_with": empty, "ends_with": empty, "exists": cty.StringVal("true"),
			},
			valid: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			expression := map[string]cty.Value{
				"one_of":      cty.NullVal(cty.List(cty.String)),
				"contains":    cty.NullVal(cty.List(cty.String)),
				"starts_with": cty.NullVal(cty.List(cty.String)),
				"ends_with":   cty.NullVal(cty.List(cty.String)),
				"exists":      cty.NullVal(cty.String),
			}
			for key, value := range test.overrides {
				expression[key] = value
			}

			err := validateExpression(cty.ObjectVal(expression))
			if test.valid && err != nil {
				t.Fatalf("expected valid expression, got %v", err)
			}
			if !test.valid && (err == nil || !strings.Contains(err.Error(), "exactly one")) {
				t.Fatalf("expected exactly one operator error, got %v", err)
			}
		})
	}
}
