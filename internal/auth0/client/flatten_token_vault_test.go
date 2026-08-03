package client

import (
	"testing"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlattenTokenVaultPrivilegedAccess(t *testing.T) {
	t.Run("returns nil when the object is absent (EA flag off, or never configured)", func(t *testing.T) {
		resourceData := schema.TestResourceDataRaw(t, NewResource().Schema, map[string]interface{}{})

		result := flattenTokenVaultPrivilegedAccess(resourceData, nil)

		assert.Nil(t, result)
	})

	t.Run("flattens ip_allowlist and grants directly from the API response", func(t *testing.T) {
		resourceData := schema.TestResourceDataRaw(t, NewResource().Schema, map[string]interface{}{})

		access := &management.ClientTokenVaultPrivilegedAccess{
			IPAllowlist: &[]string{"10.0.0.1", "192.168.1.0/24"},
			Grants: &[]management.ClientTokenVaultPrivilegedGrant{
				{
					Connection: auth0.String("Username-Password-Authentication"),
					Scopes:     &[]string{"openid", "profile"},
				},
			},
		}

		result := flattenTokenVaultPrivilegedAccess(resourceData, access)

		require.Len(t, result, 1)
		flat, ok := result[0].(map[string]interface{})
		require.True(t, ok)

		assert.ElementsMatch(t, []string{"10.0.0.1", "192.168.1.0/24"}, flat["ip_allowlist"])

		grants, ok := flat["grants"].([]interface{})
		require.True(t, ok)
		require.Len(t, grants, 1)
		grant, ok := grants[0].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Username-Password-Authentication", grant["connection"])
		assert.ElementsMatch(t, []string{"openid", "profile"}, grant["scopes"])
	})

	t.Run("preserves pem from state by matching on credential id, since GET reduces credentials to bare id references", func(t *testing.T) {
		resourceData := schema.TestResourceDataRaw(t, NewResource().Schema, map[string]interface{}{
			"token_vault_privileged_access": []interface{}{
				map[string]interface{}{
					"credentials": []interface{}{
						map[string]interface{}{
							"id":              "cred_123",
							"credential_type": "public_key",
							"pem":             "-----BEGIN PUBLIC KEY-----",
						},
					},
				},
			},
		})

		// The API echoes back only the id, exactly as GET /clients/{id} does.
		access := &management.ClientTokenVaultPrivilegedAccess{
			Credentials: &[]management.ClientCredentialID{{ID: auth0.String("cred_123")}},
			IPAllowlist: &[]string{},
			Grants:      &[]management.ClientTokenVaultPrivilegedGrant{},
		}

		result := flattenTokenVaultPrivilegedAccess(resourceData, access)

		require.Len(t, result, 1)
		flat := result[0].(map[string]interface{})
		credentials, ok := flat["credentials"].([]interface{})
		require.True(t, ok)
		require.Len(t, credentials, 1)

		cred, ok := credentials[0].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "cred_123", cred["id"])
		assert.Equal(t, "public_key", cred["credential_type"])
		assert.Equal(t, "-----BEGIN PUBLIC KEY-----", cred["pem"])
	})

	t.Run("falls back to positional matching for a credential with no id yet in state (just created)", func(t *testing.T) {
		resourceData := schema.TestResourceDataRaw(t, NewResource().Schema, map[string]interface{}{
			"token_vault_privileged_access": []interface{}{
				map[string]interface{}{
					"credentials": []interface{}{
						map[string]interface{}{
							"credential_type": "public_key",
							"pem":             "-----BEGIN PUBLIC KEY-----",
						},
					},
				},
			},
		})

		// The freshly created credential now has a server-assigned id that
		// wasn't in state before this apply.
		access := &management.ClientTokenVaultPrivilegedAccess{
			Credentials: &[]management.ClientCredentialID{{ID: auth0.String("cred_new")}},
			IPAllowlist: &[]string{},
			Grants:      &[]management.ClientTokenVaultPrivilegedGrant{},
		}

		result := flattenTokenVaultPrivilegedAccess(resourceData, access)

		require.Len(t, result, 1)
		flat := result[0].(map[string]interface{})
		credentials := flat["credentials"].([]interface{})
		require.Len(t, credentials, 1)

		cred := credentials[0].(map[string]interface{})
		assert.Equal(t, "cred_new", cred["id"])
		assert.Equal(t, "-----BEGIN PUBLIC KEY-----", cred["pem"])
	})

	t.Run("expand-then-flatten round-trips a populated config", func(t *testing.T) {
		expandData := newTokenVaultResourceData(t, cty.ListVal([]cty.Value{
			cty.ObjectVal(map[string]cty.Value{
				"credentials": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"credential_type": cty.StringVal("public_key"),
						"pem":             cty.StringVal("-----BEGIN PUBLIC KEY-----"),
					}),
				}),
				"ip_allowlist": cty.SetVal([]cty.Value{cty.StringVal("10.0.0.1"), cty.StringVal("192.168.1.0/24")}),
				"grants": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"connection": cty.StringVal("google-oauth2"),
						"scopes":     cty.SetVal([]cty.Value{cty.StringVal("openid")}),
					}),
				}),
			}),
		}))

		expanded := expandTokenVaultPrivilegedAccess(expandData)
		require.NotNil(t, expanded)

		// Simulate the server assigning a credential id on create, since PATCH
		// only ever accepts/returns id references.
		access := &management.ClientTokenVaultPrivilegedAccess{
			Credentials: &[]management.ClientCredentialID{{ID: auth0.String("cred_new")}},
			IPAllowlist: expanded.IPAllowlist,
			Grants:      expanded.Grants,
		}

		// flatten reads prior state via data.Get, which the RawConfig-only
		// ResourceData used for expand above does not populate; TestResourceDataRaw
		// builds a ResourceData with the same config that also supports Get.
		flattenData := schema.TestResourceDataRaw(t, NewResource().Schema, map[string]interface{}{
			"token_vault_privileged_access": []interface{}{
				map[string]interface{}{
					"credentials": []interface{}{
						map[string]interface{}{
							"credential_type": "public_key",
							"pem":             "-----BEGIN PUBLIC KEY-----",
						},
					},
					"ip_allowlist": []interface{}{"10.0.0.1", "192.168.1.0/24"},
					"grants": []interface{}{
						map[string]interface{}{
							"connection": "google-oauth2",
							"scopes":     []interface{}{"openid"},
						},
					},
				},
			},
		})

		flattened := flattenTokenVaultPrivilegedAccess(flattenData, access)

		require.Len(t, flattened, 1)
		flat := flattened[0].(map[string]interface{})
		assert.ElementsMatch(t, []string{"10.0.0.1", "192.168.1.0/24"}, flat["ip_allowlist"])

		credentials := flat["credentials"].([]interface{})
		require.Len(t, credentials, 1)
		cred := credentials[0].(map[string]interface{})
		assert.Equal(t, "cred_new", cred["id"])
		assert.Equal(t, "public_key", cred["credential_type"])
		assert.Equal(t, "-----BEGIN PUBLIC KEY-----", cred["pem"])

		grants := flat["grants"].([]interface{})
		require.Len(t, grants, 1)
		grant := grants[0].(map[string]interface{})
		assert.Equal(t, "google-oauth2", grant["connection"])
		assert.ElementsMatch(t, []string{"openid"}, grant["scopes"])
	})

	t.Run("flattens an empty block distinctly from an absent one", func(t *testing.T) {
		resourceData := schema.TestResourceDataRaw(t, NewResource().Schema, map[string]interface{}{})

		access := &management.ClientTokenVaultPrivilegedAccess{
			Credentials: &[]management.ClientCredentialID{},
			IPAllowlist: &[]string{},
			Grants:      &[]management.ClientTokenVaultPrivilegedGrant{},
		}

		result := flattenTokenVaultPrivilegedAccess(resourceData, access)

		require.Len(t, result, 1)
		flat := result[0].(map[string]interface{})
		assert.Empty(t, flat["credentials"])
		assert.Empty(t, flat["ip_allowlist"])
		assert.Empty(t, flat["grants"])
	})
}
