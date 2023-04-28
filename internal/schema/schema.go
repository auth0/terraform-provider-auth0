package schema

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TransformResourceToDataSource is a recursive function that
// converts an existing Resource schema to a DataSource schema.
func TransformResourceToDataSource(resourceSchema map[string]*schema.Schema) map[string]*schema.Schema {
	dataSourceSchema := make(map[string]*schema.Schema, len(resourceSchema))

	for key, definition := range resourceSchema {
		elementType := definition.Elem
		isListOrSet := definition.Type == schema.TypeList || definition.Type == schema.TypeSet

		resource, ok := elementType.(*schema.Resource)
		if ok && isListOrSet {
			elementType = &schema.Resource{
				Schema: TransformResourceToDataSource(resource.Schema),
			}
		}

		dataSourceSchema[key] = &schema.Schema{
			Computed:    true,
			ForceNew:    false,
			Required:    false,
			Optional:    false,
			Description: definition.Description,
			Type:        definition.Type,
			Set:         definition.Set,
			Elem:        elementType,
		}
	}

	return dataSourceSchema
}

// SetExistingAttributesAsOptional updates the schema of existing top level attributes by
// ensuring they are optional by setting Computed and Required to false and Optional to true.
func SetExistingAttributesAsOptional(schema map[string]*schema.Schema, keys ...string) {
	for _, attribute := range keys {
		if _, ok := schema[attribute]; !ok {
			continue
		}

		schema[attribute].Computed = false
		schema[attribute].Optional = true
		schema[attribute].Required = false
	}
}

// Clone returns a shallow clone of m.
func Clone[M ~map[K]V, K comparable, V any](m M) M {
	if m == nil {
		return nil
	}

	result := make(M, len(m))
	for key, value := range m {
		result[key] = value
	}

	return result
}
