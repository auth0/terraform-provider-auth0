package schema

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestTransformResourceToDataSource(t *testing.T) {
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

	dataSourceSchema := TransformResourceToDataSource(newMockResourceSchema)
	assertDataSourceSchemaDefinitionsAreCorrect(t, dataSourceSchema, newMockResourceSchema)
}

func assertDataSourceSchemaDefinitionsAreCorrect(
	t *testing.T,
	dataSourceSchema map[string]*schema.Schema,
	resourceSchema map[string]*schema.Schema,
) {
	assert.Equal(t, len(resourceSchema), len(dataSourceSchema))

	for key, definition := range dataSourceSchema {
		assert.Falsef(t, definition.Required, "Expected %s schema property to not be required", key)
		assert.Falsef(t, definition.Optional, "Expected %s schema property to not be optional", key)
		assert.Truef(t, definition.Computed, "Expected %s schema property to be computed", key)

		assert.Equalf(t,
			resourceSchema[key].Description,
			definition.Description,
			"Description for %s schema property does not match", key,
		)

		assert.Equalf(t,
			resourceSchema[key].Type,
			definition.Type,
			"Expected %s schema property to maintain the same type", key,
		)

		if definition.Type == schema.TypeList || definition.Type == schema.TypeSet {
			assert.NotNilf(t, definition.Elem, "Expected non-nil element for %s schema property", key)

			if elements, ok := definition.Elem.(*schema.Resource); ok {
				innerElem := resourceSchema[key].Elem.(*schema.Resource)
				assertDataSourceSchemaDefinitionsAreCorrect(t, elements.Schema, innerElem.Schema)
			}
		}
	}
}

func TestSetExistingAttributesAsOptional(t *testing.T) {
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
			Computed:    true,
			Description: "Some bool property.",
		},
	}

	expectedOptionalAttributes := []string{"string_prop", "map_prop", "bool_prop"}
	SetExistingAttributesAsOptional(newMockResourceSchema, expectedOptionalAttributes[1:]...)

	// It should not panic if we set a non-existent attribute as optional.
	SetExistingAttributesAsOptional(newMockResourceSchema, "non_existent", expectedOptionalAttributes[0])

	for _, attribute := range expectedOptionalAttributes {
		assert.True(t, newMockResourceSchema[attribute].Optional)
		assert.False(t, newMockResourceSchema[attribute].Computed)
		assert.False(t, newMockResourceSchema[attribute].Required)
	}
}

func TestSetExistingAttributesAsRequired(t *testing.T) {
	var newMockResourceSchema = map[string]*schema.Schema{
		"string_prop": {
			Type:        schema.TypeString,
			Description: "Some string property.",
			Optional:    true,
		},
		"map_prop": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "Some map property.",
		},
		"bool_prop": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Some bool property.",
		},
	}

	expectedRequiredAttributes := []string{"string_prop", "map_prop", "bool_prop"}
	SetExistingAttributesAsRequired(newMockResourceSchema, expectedRequiredAttributes[1:]...)

	// It should not panic if we set a non-existent attribute as required.
	SetExistingAttributesAsRequired(newMockResourceSchema, "non_existent", expectedRequiredAttributes[0])

	for _, attribute := range expectedRequiredAttributes {
		assert.True(t, newMockResourceSchema[attribute].Required)
		assert.False(t, newMockResourceSchema[attribute].Computed)
		assert.False(t, newMockResourceSchema[attribute].Optional)
	}
}

func TestClone(t *testing.T) {
	var m1 = map[int]int{1: 2, 2: 4, 4: 8, 8: 16}

	mc := Clone(m1)
	assert.Equal(t, m1, mc)

	mc[16] = 32
	assert.NotEqual(t, m1, mc)
}

func TestCloneNil(t *testing.T) {
	var m1 map[string]int
	mc := Clone(m1)
	assert.Nil(t, mc)
}
