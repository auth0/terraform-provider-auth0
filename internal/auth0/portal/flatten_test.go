package portal

import (
	"encoding/json"
	"testing"

	"github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stateData returns the resource data of a refresh, with the given attributes in the prior state.
func stateData(t *testing.T, attributes map[string]interface{}) *schema.ResourceData {
	t.Helper()

	resource := NewResource()
	writer := resource.Data(nil)
	writer.SetId("up_1")
	for key, value := range attributes {
		require.NoError(t, writer.Set(key, value))
	}

	return resource.Data(writer.State())
}

func unmarshal[T any](t *testing.T, payload string) *T {
	t.Helper()

	var result T
	require.NoError(t, json.Unmarshal([]byte(payload), &result))

	return &result
}

func TestFlattenPortal(t *testing.T) {
	t.Run("it flattens a portal", func(t *testing.T) {
		data := stateData(t, map[string]interface{}{
			"client": clientBlock(map[string]interface{}{"client_id": "portal-client", "client_secret": "portal-secret"}),
		})
		portal := unmarshal[management.Portal](t, `{
			"id": "up_1",
			"name": "My Portal",
			"slug": "my-portal",
			"client": {"token_endpoint_auth_method": "client_secret_post", "client_id": "portal-client"},
			"custom_code": {"head": "<style></style>"},
			"navigation": {},
			"pages": {"default": "account", "standard": []},
			"created_at": "2026-10-29T06:30:38.329Z",
			"updated_at": "2026-10-30T06:30:38.329Z"
		}`)

		require.False(t, flattenPortal(data, portal).HasError())

		assert.Equal(t, "My Portal", data.Get("name"))
		assert.Equal(t, "my-portal", data.Get("slug"))
		assert.Equal(t, "portal-client", data.Get("client.0.client_secret_post.0.client_id"))
		assert.Equal(t, "portal-secret", data.Get("client.0.client_secret_post.0.client_secret"))
		assert.Equal(t, "<style></style>", data.Get("custom_code.0.head"))
		assert.Equal(t, 1, data.Get("navigation.#"))
		assert.Equal(t, 0, data.Get("navigation.0.sidebar.#"))
		assert.Equal(t, "account", data.Get("pages.0.default"))
		assert.Equal(t, "2026-10-29T06:30:38Z", data.Get("created_at"))
		assert.Equal(t, "2026-10-30T06:30:38Z", data.Get("updated_at"))
	})

	t.Run("it removes the blocks that the API omits", func(t *testing.T) {
		data := stateData(t, map[string]interface{}{
			"custom_code": block(map[string]interface{}{"head": "<style></style>"}),
			"navigation":  emptyBlock(),
			"pages":       emptyBlock(),
		})
		portal := unmarshal[management.Portal](t, `{"id": "up_1", "name": "My Portal", "slug": "my-portal"}`)

		require.False(t, flattenPortal(data, portal).HasError())

		assert.Equal(t, 0, data.Get("client.#"))
		assert.Equal(t, 0, data.Get("custom_code.#"))
		assert.Equal(t, 0, data.Get("navigation.#"))
		assert.Equal(t, 0, data.Get("pages.#"))
	})

	t.Run("it sets no attribute when a component type is not supported", func(t *testing.T) {
		data := stateData(t, map[string]interface{}{"name": "My Portal"})
		portal := unmarshal[management.Portal](t, `{
			"id": "up_1",
			"name": "My Renamed Portal",
			"slug": "my-portal",
			"pages": {"standard": [{"title": "Account", "slug": "account", "components": [
				{"type": "page:component:auth0:unknown"}
			]}]}
		}`)

		diagnostics := flattenPortal(data, portal)

		require.True(t, diagnostics.HasError())
		assert.Equal(t, "My Portal", data.Get("name"))
	})
}

func TestFlattenClient(t *testing.T) {
	client := &management.PortalClientSecretPost{ClientID: "portal-client"}

	t.Run("it keeps the client secret from the prior state", func(t *testing.T) {
		data := stateData(t, map[string]interface{}{
			"client": clientBlock(map[string]interface{}{"client_id": "old-client", "client_secret": "portal-secret"}),
		})

		assert.Equal(t, clientBlock(map[string]interface{}{
			"client_id":     "portal-client",
			"client_secret": "portal-secret",
		}), flattenClient(data, client))
	})

	t.Run("it keeps the write-only secret version from the prior state", func(t *testing.T) {
		data := stateData(t, map[string]interface{}{
			"client": clientBlock(map[string]interface{}{"client_id": "portal-client", "client_secret_wo_version": 2}),
		})

		assert.Equal(t, clientBlock(map[string]interface{}{
			"client_id":                "portal-client",
			"client_secret_wo_version": 2,
		}), flattenClient(data, client))
	})

	t.Run("it sets only the client ID after an import", func(t *testing.T) {
		assert.Equal(t, clientBlock(map[string]interface{}{
			"client_id": "portal-client",
		}), flattenClient(stateData(t, nil), client))
	})

	t.Run("it returns no block when the API omits the client", func(t *testing.T) {
		assert.Nil(t, flattenClient(stateData(t, nil), nil))
	})
}

func TestFlattenNavigation(t *testing.T) {
	t.Run("it returns no block when the API omits the navigation", func(t *testing.T) {
		navigation, err := flattenNavigation(nil)

		require.NoError(t, err)
		assert.Nil(t, navigation)
	})

	t.Run("it returns an empty block for a navigation without a sidebar", func(t *testing.T) {
		navigation, err := flattenNavigation(unmarshal[management.PortalNavigation](t, `{}`))

		require.NoError(t, err)
		assert.Equal(t, emptyBlock(), navigation)
	})

	t.Run("it flattens the V2 links", func(t *testing.T) {
		navigation, err := flattenNavigation(unmarshal[management.PortalNavigation](t, `{"sidebar": {"components": [
			{"type": "sidebar:component:auth0:internal_link:v2", "config": {"label": "Account", "page_slug": "account", "icon": "user"}},
			{"type": "sidebar:component:auth0:external_link:v2", "config": {"label": "Help", "href": "https://help.example.com"}}
		]}}`))

		require.NoError(t, err)
		assert.Equal(t, block(map[string]interface{}{
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
						"icon":  "",
					})},
				},
			}),
		}), navigation)
	})

	t.Run("it returns an error for a V1 link", func(t *testing.T) {
		_, err := flattenNavigation(unmarshal[management.PortalNavigation](t, `{"sidebar": {"components": [
			{"type": "sidebar:component:auth0:external_link", "config": {"label": "Help", "url": "https://help.example.com"}}
		]}}`))

		assert.EqualError(t, err, "a component at `navigation.sidebar.components[0]` of type "+
			"`sidebar:component:auth0:external_link` that this provider version does not support")
	})
}

func TestFlattenPages(t *testing.T) {
	t.Run("it returns no block when the API omits the pages", func(t *testing.T) {
		pages, err := flattenPages(nil)

		require.NoError(t, err)
		assert.Nil(t, pages)
	})

	t.Run("it flattens every component type", func(t *testing.T) {
		portal := unmarshal[management.Portal](t, fullConfigJSON)

		pages, err := flattenPages(portal.Pages)

		require.NoError(t, err)
		assert.Equal(t, fullConfig()["pages"], pages)
	})

	t.Run("it returns an error with the path of an unsupported section child", func(t *testing.T) {
		_, err := flattenPages(unmarshal[management.PortalPagesByType](t, `{"standard": [
			{"title": "Account", "slug": "account"},
			{"title": "Settings", "slug": "settings", "components": [
				{"type": "page:component:auth0:typography:rich_text", "config": {"content": "<p>Hello</p>"}},
				{"type": "page:component:auth0:structure:section", "config": {"variant": "none", "children": [
					{"type": "page:component:auth0:unknown"}
				]}}
			]}
		]}`))

		assert.EqualError(t, err, "a component at `pages.standard[1].components[1].config.children[0]` of type "+
			"`page:component:auth0:unknown` that this provider version does not support")
	})
}
