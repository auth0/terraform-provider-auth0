package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// dataSourceSchemaFromResourceSchema is a recursive function that converts an existing
// Resource schema to a Datasource schema.
//
// All schema properties are copied, but some are ignored or changed:
// - All fields have Computed = true.
// - All fields have ForceNew, Required = false.
// - Validation and attributes (e.g. MaxItems) are not copied.
func dataSourceSchemaFromResourceSchema(resourceSchema map[string]*schema.Schema) map[string]*schema.Schema {
	dataSourceSchema := make(map[string]*schema.Schema, len(resourceSchema))

	for key, definition := range resourceSchema {
		dataSourceKeyDefinition := &schema.Schema{
			Computed:    true,
			ForceNew:    false,
			Required:    false,
			Description: definition.Description,
			Type:        definition.Type,
		}

		switch definition.Type {
		case schema.TypeSet:
			dataSourceKeyDefinition.Set = definition.Set
			fallthrough
		case schema.TypeList:
			// List & Set types are generally used for 2 cases:
			// - a list/set of simple primitive values (e.g. list of strings)
			// - a sub resource
			if elem, ok := definition.Elem.(*schema.Resource); ok {
				// Handle the case where the Element is a sub-resource.
				dataSourceKeyDefinition.Elem = &schema.Resource{
					Schema: dataSourceSchemaFromResourceSchema(elem.Schema),
				}
			} else {
				// Handle simple primitive case.
				dataSourceKeyDefinition.Elem = definition.Elem
			}
		default:
			// Elem of all other types are copied as-is.
			dataSourceKeyDefinition.Elem = definition.Elem
		}

		dataSourceSchema[key] = dataSourceKeyDefinition
	}

	return dataSourceSchema
}

// fixDatasourceSchemaFlags is a convenience function that toggles the Computed,
// Optional + Required flags on a schema element. This is useful when the schema
// has been generated (using `dataSourceSchemaFromResourceSchema` above for
// example) and therefore the attribute flags were not set appropriately when
// first added to the schema definition. Currently only supports top-level
// schema elements.
func fixDatasourceSchemaFlags(schema map[string]*schema.Schema, required bool, keys ...string) {
	for _, v := range keys {
		schema[v].Computed = false
		schema[v].Optional = !required
		schema[v].Required = required
	}
}

func addOptionalFieldsToSchema(schema map[string]*schema.Schema, keys ...string) {
	fixDatasourceSchemaFlags(schema, false, keys...)
}
