package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

var newMockResourceSchema = map[string]*schema.Schema{
	"string_prop": {
		Type:        schema.TypeString,
		Description: "Some string property.",
		Required:    true,
	},
	"map_prop": {
		Type:        schema.TypeMap,
		Optional:    true,
		Description: "Some map property.",
	},
	"bool_prop": {
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    false,
		Description: "Some bool property.",
	},
	"list_prop": {
		Type:        schema.TypeList,
		Description: "Some list property.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"inner_list": {
					Type:        schema.TypeList,
					Description: "Description for list_prop.inner_list.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"inner_list_element": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Description for list_prop.inner_list.inner_list_element.",
							},
						},
					},
				},
			},
		},
		Optional: true,
	},
	"float_prop": {
		Type:        schema.TypeFloat,
		Optional:    true,
		Computed:    false,
		Description: "Some float property.",
	},
	"set_prop": {
		Type:        schema.TypeSet,
		Optional:    true,
		Description: "Some set property passed into mock schema.",
		Elem:        &schema.Schema{Type: schema.TypeString},
	},
}

func TestDatasourceSchemaFromResourceSchema(t *testing.T) {
	dataSourceSchema := dataSourceSchemaFromResourceSchema(newMockResourceSchema)
	assertDataSourceSchemaDefinitionsAreCorrect(t, dataSourceSchema, newMockResourceSchema)
}

func assertDataSourceSchemaDefinitionsAreCorrect(
	t *testing.T,
	dataSourceSchema map[string]*schema.Schema,
	resourceSchema map[string]*schema.Schema,
) {
	assert.Equal(t, len(resourceSchema), len(dataSourceSchema))

	for key, definition := range dataSourceSchema {
		assert.Falsef(t, definition.Optional, "Expected %s schema property to be required", key)
		assert.Truef(t, definition.Computed, "Expected %s schema property to be computed", key)

		assert.Equalf(
			t,
			resourceSchema[key].Description,
			definition.Description,
			"Description for %s schema property does not match",
			key,
		)

		assert.Equalf(
			t,
			resourceSchema[key].Type,
			definition.Type,
			"Expected %s schema property to maintain the same type",
			key,
		)

		if definition.Type == schema.TypeList || definition.Type == schema.TypeSet {
			assert.NotNil(t, definition.Elem)
		}

		if definition.Type == schema.TypeList {
			if elements, ok := definition.Elem.(*schema.Resource); ok {
				innerElem := resourceSchema[key].Elem.(*schema.Resource)
				assertDataSourceSchemaDefinitionsAreCorrect(t, elements.Schema, innerElem.Schema)
			}
		}
	}
}
