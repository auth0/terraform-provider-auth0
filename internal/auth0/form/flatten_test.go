package form

import (
	"reflect"
	"testing"

	"github.com/auth0/go-auth0"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestFlattenForm(t *testing.T) {
	cases := []struct {
		form     *management.Form
		expected map[string]interface{}
	}{
		{
			form: &management.Form{
				Name: auth0.String("test-form"),
				Languages: &management.FormLanguages{
					Primary: auth0.String("en"),
					Default: auth0.String("en"),
				},
				Style: &map[string]interface{}{"styleKey": "styleValue"},
				Translations: &map[string]interface{}{
					"translationKey": "translationValue",
				},
				Start:  &map[string]interface{}{"startKey": "startValue"},
				Nodes:  []interface{}{"node1", "node2"},
				Ending: &map[string]interface{}{"endKey": "endValue"},
				Messages: &management.FormMessages{
					Errors: &map[string]interface{}{
						"errorKey": "errorValue",
					},
					Custom: &map[string]interface{}{
						"customKey": "customValue",
					},
				},
			},
			expected: map[string]interface{}{
				"name": "test-form",
				"languages": []interface{}{
					map[string]interface{}{
						"primary": "en",
						"default": "en",
					},
				},
				"style":        `{"styleKey":"styleValue"}`,
				"translations": `{"translationKey":"translationValue"}`,
				"start":        `{"startKey":"startValue"}`,
				"nodes":        `["node1","node2"]`,
				"ending":       `{"endKey":"endValue"}`,
				"messages": []interface{}{
					map[string]interface{}{
						"custom": "{\"customKey\":\"customValue\"}",
						"errors": "{\"errorKey\":\"errorValue\"}"},
				},
			},
		},
	}

	for _, c := range cases {
		data := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
			"name":         {Type: schema.TypeString},
			"languages":    {Type: schema.TypeList, Elem: &schema.Resource{Schema: map[string]*schema.Schema{"primary": {Type: schema.TypeString}, "default": {Type: schema.TypeString}}}},
			"style":        {Type: schema.TypeString},
			"translations": {Type: schema.TypeString},
			"messages":     {Type: schema.TypeList, Elem: &schema.Resource{Schema: map[string]*schema.Schema{"errors": {Type: schema.TypeString}, "custom": {Type: schema.TypeString}}}},
			"start":        {Type: schema.TypeString},
			"nodes":        {Type: schema.TypeString},
			"ending":       {Type: schema.TypeString},
		}, map[string]interface{}{})

		err := flattenForm(data, c.form)
		if err != nil {
			t.Fatalf("Error flattening form: %v", err)
		}

		for k, v := range c.expected {
			if !reflect.DeepEqual(data.Get(k), v) {
				t.Fatalf("Error matching output and expected for key %s: %#v vs %#v", k, data.Get(k), v)
			}
		}
	}
}
