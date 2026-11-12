package portal

import (
	"maps"
	"slices"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaIsValid(t *testing.T) {
	require.NoError(t, NewResource().InternalValidate(nil, true))
	require.NoError(t, NewDataSource().InternalValidate(nil, false))
}

func TestDataSourceSchema(t *testing.T) {
	secretPostSchema := func(resourceSchema map[string]*schema.Schema) map[string]*schema.Schema {
		client := resourceSchema["client"].Elem.(*schema.Resource)
		return client.Schema["client_secret_post"].Elem.(*schema.Resource).Schema
	}

	t.Run("it has no secret attributes", func(t *testing.T) {
		assert.ElementsMatch(t, []string{"client_id"}, slices.Collect(maps.Keys(secretPostSchema(dataSourceSchema()))))
	})

	t.Run("it does not change the resource schema", func(t *testing.T) {
		assert.Contains(t, secretPostSchema(NewResource().Schema), "client_secret_wo")
	})
}

// TestAllComponentTypes ensures the test data has one entry of every component type in the schema. The expand
// and flatten tests use this data both on a page and in a section, and the SDK models these two places with
// separate types. So a new component type that is not mapped in both places makes those tests fail.
func TestAllComponentTypes(t *testing.T) {
	var names []string
	for _, component := range allComponentTypes() {
		names = append(names, slices.Collect(maps.Keys(component.(map[string]interface{})))...)
	}

	assert.ElementsMatch(t, slices.Collect(maps.Keys(componentTypes(false))), names)
}

func sidebarConfig(components ...interface{}) map[string]interface{} {
	return withConfig(minimalConfig(), "navigation", block(map[string]interface{}{
		"sidebar": block(map[string]interface{}{"component": components}),
	}))
}

func pagesConfig(components ...interface{}) map[string]interface{} {
	return withConfig(minimalConfig(), "pages", block(map[string]interface{}{
		"standard": []interface{}{
			map[string]interface{}{"title": "Account", "slug": "account"},
			map[string]interface{}{"title": "Settings", "slug": "settings", "component": components},
		},
	}))
}

func sectionConfig(children ...interface{}) map[string]interface{} {
	return pagesConfig(map[string]interface{}{"structure_section": block(map[string]interface{}{
		"variant": "card",
		"child":   children,
	})})
}

var (
	internalLink = map[string]interface{}{"internal_link_v2": block(map[string]interface{}{"label": "Account", "page_slug": "account"})}
	externalLink = map[string]interface{}{"external_link_v2": block(map[string]interface{}{"label": "Help", "href": "https://help.example.com"})}
	heading      = map[string]interface{}{"typography_heading": block(map[string]interface{}{"title": "Heading"})}
	mfa          = map[string]interface{}{"my_account_mfa_management": emptyBlock()}
)

func TestValidateSidebarComponents(t *testing.T) {
	t.Run("it accepts components that set one link each", func(t *testing.T) {
		assert.NoError(t, validateSidebarComponents(rawConfig(t, sidebarConfig(internalLink, externalLink))))
	})

	t.Run("it accepts a portal without navigation", func(t *testing.T) {
		assert.NoError(t, validateSidebarComponents(rawConfig(t, minimalConfig())))
	})

	t.Run("it rejects a component that sets no link", func(t *testing.T) {
		err := validateSidebarComponents(rawConfig(t, sidebarConfig(internalLink, map[string]interface{}{})))

		assert.ErrorContains(t, err, "`navigation.sidebar.component[1]` must set exactly one of "+
			"`external_link_v2`, `internal_link_v2`, but sets none")
	})

	t.Run("it rejects a component that sets two links", func(t *testing.T) {
		both := map[string]interface{}{
			"internal_link_v2": internalLink["internal_link_v2"],
			"external_link_v2": externalLink["external_link_v2"],
		}

		err := validateSidebarComponents(rawConfig(t, sidebarConfig(both)))

		assert.ErrorContains(t, err, "`navigation.sidebar.component[0]` must set exactly one of "+
			"`external_link_v2`, `internal_link_v2`, but sets `external_link_v2`, `internal_link_v2`")
	})
}

func TestValidatePages(t *testing.T) {
	t.Run("it accepts every component type and section child type", func(t *testing.T) {
		assert.NoError(t, validatePages(rawConfig(t, fullConfig())))
	})

	t.Run("it rejects a page component that sets two component types", func(t *testing.T) {
		both := map[string]interface{}{
			"typography_heading":        heading["typography_heading"],
			"my_account_mfa_management": mfa["my_account_mfa_management"],
		}

		err := validatePages(rawConfig(t, pagesConfig(heading, both)))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "`pages.standard[1].component[1]` must set exactly one of `form`, ")
		assert.Contains(t, err.Error(), "but sets `my_account_mfa_management`, `typography_heading`")
	})

	t.Run("it reports every invalid page component and section child", func(t *testing.T) {
		err := validatePages(rawConfig(t, pagesConfig(
			map[string]interface{}{},
			map[string]interface{}{"structure_section": block(map[string]interface{}{
				"variant": "card",
				"child":   []interface{}{map[string]interface{}{}},
			})},
			map[string]interface{}{},
		)))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "3 errors occurred")
		assert.Contains(t, err.Error(), "`pages.standard[1].component[0]` must set exactly one of ")
		assert.Contains(t, err.Error(), "`pages.standard[1].component[1].structure_section.child[0]` must set exactly one of ")
		assert.Contains(t, err.Error(), "`pages.standard[1].component[2]` must set exactly one of ")
	})

	t.Run("it rejects a section child that sets no component type", func(t *testing.T) {
		err := validatePages(rawConfig(t, sectionConfig(mfa, map[string]interface{}{})))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "`pages.standard[1].component[0].structure_section.child[1]` must set exactly one of ")
		assert.Contains(t, err.Error(), "but sets none")
		assert.NotContains(t, err.Error(), "structure_section`")
	})
}

func TestValidateClientSecret(t *testing.T) {
	t.Run("it accepts a client secret", func(t *testing.T) {
		assert.NoError(t, validateClientSecret(rawConfig(t, minimalConfig())))
	})

	t.Run("it accepts a write-only client secret", func(t *testing.T) {
		config := withConfig(minimalConfig(), "client", clientBlock(map[string]interface{}{
			"client_id":                "portal-client",
			"client_secret_wo":         "portal-secret",
			"client_secret_wo_version": 1,
		}))

		assert.NoError(t, validateClientSecret(rawConfig(t, config)))
	})

	t.Run("it accepts an unknown client secret", func(t *testing.T) {
		raw := cty.ObjectVal(map[string]cty.Value{
			"client": cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
				"client_secret_post": cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
					"client_id":        cty.StringVal("portal-client"),
					"client_secret":    cty.UnknownVal(cty.String),
					"client_secret_wo": cty.NullVal(cty.String),
				})}),
			})}),
		})

		assert.NoError(t, validateClientSecret(raw))
	})

	t.Run("it rejects a client without a secret", func(t *testing.T) {
		config := withConfig(minimalConfig(), "client", clientBlock(map[string]interface{}{"client_id": "portal-client"}))

		assert.EqualError(t, validateClientSecret(rawConfig(t, config)),
			"`client.client_secret_post` must set one of `client_secret`, `client_secret_wo`, but sets none")
	})
}

func TestValidateExactlyOneOf(t *testing.T) {
	setBlock := cty.ListVal([]cty.Value{cty.EmptyObjectVal})
	nullBlock := cty.NullVal(cty.List(cty.EmptyObject))
	unknownBlock := cty.UnknownVal(cty.List(cty.EmptyObject))
	entry := func(first, second cty.Value) cty.Value {
		return cty.ObjectVal(map[string]cty.Value{"first": first, "second": second})
	}

	t.Run("it skips an entry with an unknown type block and no set type block", func(t *testing.T) {
		list := cty.ListVal([]cty.Value{entry(unknownBlock, nullBlock)})

		assert.NoError(t, validateExactlyOneOf(list, "component"))
	})

	t.Run("it rejects an entry with an unknown type block and two set type blocks", func(t *testing.T) {
		list := cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
			"first": setBlock, "second": setBlock, "third": unknownBlock,
		})})

		assert.ErrorContains(t, validateExactlyOneOf(list, "component"),
			"`component[0]` must set exactly one of `first`, `second`, `third`, but sets `first`, `second`")
	})

	t.Run("it skips an unknown entry and keeps the indexes of the configuration", func(t *testing.T) {
		list := cty.ListVal([]cty.Value{
			cty.UnknownVal(cty.Object(map[string]cty.Type{"first": cty.List(cty.EmptyObject), "second": cty.List(cty.EmptyObject)})),
			entry(nullBlock, nullBlock),
		})

		assert.ErrorContains(t, validateExactlyOneOf(list, "component"),
			"`component[1]` must set exactly one of `first`, `second`, but sets none")
	})

	t.Run("it reports every invalid entry", func(t *testing.T) {
		list := cty.ListVal([]cty.Value{entry(nullBlock, nullBlock), entry(setBlock, nullBlock), entry(setBlock, setBlock)})

		err := validateExactlyOneOf(list, "component")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "2 errors occurred")
		assert.Contains(t, err.Error(), "`component[0]` must set exactly one of `first`, `second`, but sets none")
		assert.Contains(t, err.Error(), "`component[2]` must set exactly one of `first`, `second`, but sets `first`, `second`")
	})

	t.Run("it skips an unknown list", func(t *testing.T) {
		list := cty.UnknownVal(cty.List(cty.Object(map[string]cty.Type{"first": cty.List(cty.EmptyObject)})))

		assert.NoError(t, validateExactlyOneOf(list, "component"))
	})
}
