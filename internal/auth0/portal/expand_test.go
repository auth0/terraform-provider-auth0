package portal

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	ctyjson "github.com/hashicorp/go-cty/cty/json"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// rawConfig converts a configuration map to the raw configuration value that Terraform sends.
// Attributes and blocks that are not in the map are null.
func rawConfig(t *testing.T, config map[string]interface{}) cty.Value {
	t.Helper()

	encoded, err := json.Marshal(config)
	require.NoError(t, err)

	impliedType := schema.InternalMap(NewResource().SchemaMap()).CoreConfigSchema().ImpliedType()
	raw, err := ctyjson.Unmarshal(encoded, impliedType)
	require.NoError(t, err)

	return raw
}

// resourceData returns the resource data of a plan from the prior state to the configuration.
// TestResourceDataRaw is not used, because it does not set the raw configuration.
func resourceData(t *testing.T, prior, config map[string]interface{}) *schema.ResourceData {
	t.Helper()

	resource := NewResource()

	var priorState *terraform.InstanceState
	if prior != nil {
		writer := resource.Data(nil)
		writer.SetId("up_1")
		for key, value := range prior {
			require.NoError(t, writer.Set(key, value))
		}
		priorState = writer.State()
	}

	resourceSchema := schema.InternalMap(resource.SchemaMap())
	diff, err := resourceSchema.Diff(context.Background(), priorState, terraform.NewResourceConfigRaw(config), nil, nil, true)
	require.NoError(t, err)
	if diff == nil {
		diff = &terraform.InstanceDiff{}
	}
	diff.RawConfig = rawConfig(t, config)

	data, err := resourceSchema.Data(priorState, diff)
	require.NoError(t, err)

	return data
}

func marshal(t *testing.T, v interface{}) string {
	t.Helper()

	encoded, err := json.Marshal(v)
	require.NoError(t, err)

	return string(encoded)
}

func block(fields map[string]interface{}) []interface{} {
	return []interface{}{fields}
}

func emptyBlock() []interface{} {
	return block(map[string]interface{}{})
}

func clientBlock(secretPost map[string]interface{}) []interface{} {
	return block(map[string]interface{}{"client_secret_post": block(secretPost)})
}

func minimalConfig() map[string]interface{} {
	return map[string]interface{}{
		"name": "My Portal",
		"slug": "my-portal",
		"client": clientBlock(map[string]interface{}{
			"client_id":     "portal-client",
			"client_secret": "portal-secret",
		}),
	}
}

func withConfig(config map[string]interface{}, key string, value interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(config)+1)
	for k, v := range config {
		result[k] = v
	}
	result[key] = value

	return result
}

// allComponentTypes returns one entry of each component type that is allowed both on a page and in a section.
func allComponentTypes() []interface{} {
	return []interface{}{
		map[string]interface{}{"form": block(map[string]interface{}{
			"form_id":            "ap_form",
			"completion_message": "Saved.",
		})},
		map[string]interface{}{"typography_heading": block(map[string]interface{}{
			"title":       "Heading",
			"description": "Description.",
		})},
		map[string]interface{}{"typography_rich_text": block(map[string]interface{}{
			"content": "<p>Hello</p>",
		})},
		map[string]interface{}{"media_image": block(map[string]interface{}{
			"src":      "https://example.com/a.png",
			"position": "center",
			"height":   120,
		})},
		map[string]interface{}{"structure_separator": block(map[string]interface{}{
			"variant": "dashed",
			"text":    "or",
		})},
		map[string]interface{}{"my_organization_domain_table": emptyBlock()},
		map[string]interface{}{"my_organization_member_management": emptyBlock()},
		map[string]interface{}{"my_organization_details_edit": emptyBlock()},
		map[string]interface{}{"my_account_passkey_management": emptyBlock()},
		map[string]interface{}{"my_account_mfa_management": emptyBlock()},
	}
}

const allComponentTypesJSON = `
	{"type": "page:component:auth0:form", "config": {"form_id": "ap_form", "completion_message": "Saved."}},
	{"type": "page:component:auth0:typography:heading", "config": {"title": "Heading", "description": "Description."}},
	{"type": "page:component:auth0:typography:rich_text", "config": {"content": "<p>Hello</p>"}},
	{"type": "page:component:auth0:media:image", "config": {"src": "https://example.com/a.png", "position": "center", "height": 120}},
	{"type": "page:component:auth0:structure:separator", "config": {"variant": "dashed", "text": "or"}},
	{"type": "page:component:auth0:my_organization:domain_table"},
	{"type": "page:component:auth0:my_organization:member_management"},
	{"type": "page:component:auth0:my_organization:details_edit"},
	{"type": "page:component:auth0:my_account:passkey_management"},
	{"type": "page:component:auth0:my_account:mfa_management"}`

func fullConfig() map[string]interface{} {
	config := minimalConfig()
	config["custom_code"] = block(map[string]interface{}{"head": "<style></style>"})
	config["navigation"] = block(map[string]interface{}{
		"sidebar": block(map[string]interface{}{
			"component": []interface{}{
				map[string]interface{}{"internal_link_v2": block(map[string]interface{}{
					"label":     "Account",
					"page_slug": "account",
					"icon":      "user",
				})},
				map[string]interface{}{"external_link_v2": block(map[string]interface{}{
					"label": "Help",
					"href":  "https://help.example.com",
				})},
			},
		}),
	})
	config["pages"] = block(map[string]interface{}{
		"default": "account",
		"standard": []interface{}{
			map[string]interface{}{
				"title": "Account",
				"slug":  "account",
				"component": append(allComponentTypes(), map[string]interface{}{
					"structure_section": block(map[string]interface{}{
						"variant":     "card",
						"title":       "More",
						"description": "More settings.",
						"child":       allComponentTypes(),
					}),
				}),
			},
		},
	})

	return config
}

const fullConfigJSON = `{
	"name": "My Portal",
	"slug": "my-portal",
	"client": {"token_endpoint_auth_method": "client_secret_post", "client_id": "portal-client", "client_secret": "portal-secret"},
	"custom_code": {"head": "<style></style>"},
	"navigation": {"sidebar": {"components": [
		{"type": "sidebar:component:auth0:internal_link:v2", "config": {"label": "Account", "page_slug": "account", "icon": "user"}},
		{"type": "sidebar:component:auth0:external_link:v2", "config": {"label": "Help", "href": "https://help.example.com"}}
	]}},
	"pages": {"default": "account", "standard": [{"title": "Account", "slug": "account", "components": [` +
	allComponentTypesJSON + `,
		{"type": "page:component:auth0:structure:section", "config": {"variant": "card", "title": "More", "description": "More settings.", "children": [` +
	allComponentTypesJSON + `
		]}}
	]}]}
}`

func TestExpandPortalCreate(t *testing.T) {
	t.Run("it sends only the client when no optional block is set", func(t *testing.T) {
		request := expandPortalCreate(resourceData(t, nil, minimalConfig()))

		assert.JSONEq(t, `{
			"name": "My Portal",
			"slug": "my-portal",
			"client": {"token_endpoint_auth_method": "client_secret_post", "client_id": "portal-client", "client_secret": "portal-secret"}
		}`, marshal(t, request))
	})

	t.Run("it sends every block and component type", func(t *testing.T) {
		request := expandPortalCreate(resourceData(t, nil, fullConfig()))

		assert.JSONEq(t, fullConfigJSON, marshal(t, request))
	})

	t.Run("it sends the write-only client secret", func(t *testing.T) {
		config := withConfig(minimalConfig(), "client", clientBlock(map[string]interface{}{
			"client_id":                "portal-client",
			"client_secret_wo":         "write-only-secret",
			"client_secret_wo_version": 1,
		}))

		request := expandPortalCreate(resourceData(t, nil, config))

		assert.Equal(t, "write-only-secret", request.Client.ClientSecret)
	})

	t.Run("it sends an empty client secret when no secret is set", func(t *testing.T) {
		config := withConfig(minimalConfig(), "client", clientBlock(map[string]interface{}{
			"client_id": "portal-client",
		}))

		request := expandPortalCreate(resourceData(t, nil, config))

		assert.Equal(t, "", request.Client.ClientSecret)
	})

	t.Run("it sends empty blocks as empty objects", func(t *testing.T) {
		config := withConfig(minimalConfig(), "custom_code", emptyBlock())
		config["navigation"] = emptyBlock()
		config["pages"] = emptyBlock()

		request := expandPortalCreate(resourceData(t, nil, config))

		assert.JSONEq(t, `{
			"name": "My Portal",
			"slug": "my-portal",
			"client": {"token_endpoint_auth_method": "client_secret_post", "client_id": "portal-client", "client_secret": "portal-secret"},
			"custom_code": {},
			"navigation": {},
			"pages": {}
		}`, marshal(t, request))
	})

	t.Run("it sends an empty sidebar as an empty object", func(t *testing.T) {
		config := withConfig(minimalConfig(), "navigation", block(map[string]interface{}{"sidebar": emptyBlock()}))

		request := expandPortalCreate(resourceData(t, nil, config))

		assert.JSONEq(t, `{"sidebar": {}}`, marshal(t, request.Navigation))
	})

	t.Run("it sends a section without children with an empty children list", func(t *testing.T) {
		config := withConfig(minimalConfig(), "pages", block(map[string]interface{}{
			"standard": []interface{}{map[string]interface{}{
				"title": "Account",
				"slug":  "account",
				"component": []interface{}{map[string]interface{}{
					"structure_section": block(map[string]interface{}{"variant": "none"}),
				}},
			}},
		}))

		request := expandPortalCreate(resourceData(t, nil, config))

		assert.JSONEq(t, `{"standard": [{"title": "Account", "slug": "account", "components": [
			{"type": "page:component:auth0:structure:section", "config": {"variant": "none", "children": []}}
		]}]}`, marshal(t, request.Pages))
	})

	t.Run("it sends a separator without attributes without a config", func(t *testing.T) {
		config := withConfig(minimalConfig(), "pages", block(map[string]interface{}{
			"standard": []interface{}{map[string]interface{}{
				"title":     "Account",
				"slug":      "account",
				"component": []interface{}{map[string]interface{}{"structure_separator": emptyBlock()}},
			}},
		}))

		request := expandPortalCreate(resourceData(t, nil, config))

		assert.JSONEq(t, `{"standard": [{"title": "Account", "slug": "account", "components": [
			{"type": "page:component:auth0:structure:separator"}
		]}]}`, marshal(t, request.Pages))
	})

	t.Run("it sends empty strings as they are", func(t *testing.T) {
		config := withConfig(minimalConfig(), "custom_code", block(map[string]interface{}{"head": ""}))
		config["navigation"] = block(map[string]interface{}{
			"sidebar": block(map[string]interface{}{
				"component": []interface{}{map[string]interface{}{"external_link_v2": block(map[string]interface{}{
					"label": "Help",
					"href":  "https://help.example.com",
					"icon":  "",
				})}},
			}),
		})

		request := expandPortalCreate(resourceData(t, nil, config))

		assert.JSONEq(t, `{"head": ""}`, marshal(t, request.CustomCode))
		assert.JSONEq(t, `{"sidebar": {"components": [
			{"type": "sidebar:component:auth0:external_link:v2", "config": {"label": "Help", "href": "https://help.example.com", "icon": ""}}
		]}}`, marshal(t, request.Navigation))
	})
}

func TestExpandPortalUpdate(t *testing.T) {
	prior := map[string]interface{}{
		"name":        "My Portal",
		"slug":        "my-portal",
		"client":      clientBlock(map[string]interface{}{"client_id": "portal-client", "client_secret": "portal-secret"}),
		"custom_code": block(map[string]interface{}{"head": "<style></style>"}),
	}
	config := withConfig(minimalConfig(), "custom_code", block(map[string]interface{}{"head": "<style></style>"}))

	t.Run("it sends an empty object when nothing changed", func(t *testing.T) {
		request := expandPortalUpdate(resourceData(t, prior, config))

		assert.JSONEq(t, `{}`, marshal(t, request))
	})

	t.Run("it sends only the changed top-level attributes", func(t *testing.T) {
		changed := withConfig(config, "name", "My Renamed Portal")
		changed["slug"] = "my-renamed-portal"

		request := expandPortalUpdate(resourceData(t, prior, changed))

		assert.JSONEq(t, `{"name": "My Renamed Portal", "slug": "my-renamed-portal"}`, marshal(t, request))
	})

	t.Run("it sends the whole client when one client attribute changed", func(t *testing.T) {
		changed := withConfig(config, "client", clientBlock(map[string]interface{}{
			"client_id":     "other-client",
			"client_secret": "portal-secret",
		}))

		request := expandPortalUpdate(resourceData(t, prior, changed))

		assert.JSONEq(t, `{"client": {
			"token_endpoint_auth_method": "client_secret_post", "client_id": "other-client", "client_secret": "portal-secret"
		}}`, marshal(t, request))
	})

	t.Run("it sends the write-only client secret when its version changed", func(t *testing.T) {
		priorWriteOnly := withConfig(prior, "client", clientBlock(map[string]interface{}{
			"client_id":                "portal-client",
			"client_secret_wo_version": 1,
		}))
		changed := withConfig(config, "client", clientBlock(map[string]interface{}{
			"client_id":                "portal-client",
			"client_secret_wo":         "rotated-secret",
			"client_secret_wo_version": 2,
		}))

		request := expandPortalUpdate(resourceData(t, priorWriteOnly, changed))

		assert.JSONEq(t, `{"client": {
			"token_endpoint_auth_method": "client_secret_post", "client_id": "portal-client", "client_secret": "rotated-secret"
		}}`, marshal(t, request))
	})

	t.Run("it sends a changed block whole", func(t *testing.T) {
		changed := fullConfig()

		request := expandPortalUpdate(resourceData(t, prior, changed))

		expected := map[string]json.RawMessage{}
		require.NoError(t, json.Unmarshal([]byte(fullConfigJSON), &expected))
		assert.JSONEq(t, marshal(t, map[string]json.RawMessage{
			"navigation": expected["navigation"],
			"pages":      expected["pages"],
		}), marshal(t, request))
	})

	t.Run("it sends null for a removed block", func(t *testing.T) {
		request := expandPortalUpdate(resourceData(t, prior, minimalConfig()))

		assert.JSONEq(t, `{"custom_code": null}`, marshal(t, request))
	})
}
