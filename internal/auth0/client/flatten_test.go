package client

import (
	"testing"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestFlattenClientAddonSAML2_ForDefaultValues(t *testing.T) {
	addon := &management.SAML2ClientAddon{
		Recipient:   auth0.String("https://example.com/saml"),
		Destination: auth0.String("https://example.com/saml"),
	}

	result := flattenClientAddonSAML2(addon)

	assert.Len(t, result, 1)
	flat, ok := result[0].(map[string]interface{})
	assert.True(t, ok, "expected result[0] to be a map[string]interface{}")

	assert.Equal(t, "https://example.com/saml", flat["recipient"])
	assert.Equal(t, "https://example.com/saml", flat["destination"])

	// Bool/int fields with non-zero Auth0 defaults must be populated when nil.
	assert.Equal(t, samlDefault.createUPNClaim, flat["create_upn_claim"], "create_upn_claim")
	assert.Equal(t, samlDefault.passthroughClaimsWithNoMapping, flat["passthrough_claims_with_no_mapping"], "passthrough_claims_with_no_mapping")
	assert.Equal(t, samlDefault.mapIdentities, flat["map_identities"], "map_identities")
	assert.Equal(t, samlDefault.typedAttributes, flat["typed_attributes"], "typed_attributes")
	assert.Equal(t, samlDefault.includeAttributeNameFormat, flat["include_attribute_name_format"], "include_attribute_name_format")
	assert.Equal(t, samlDefault.lifetimeInSeconds, flat["lifetime_in_seconds"], "lifetime_in_seconds")
}

// TestFlattenClientAddonSAML2_ForExplicitValues verifies that explicitly set fields
// are used as-is and are present in the flattened map.
func TestFlattenClientAddonSAML2_ForExplicitValues(t *testing.T) {
	addon := &management.SAML2ClientAddon{
		Recipient:                      auth0.String("https://example.com/saml"),
		Destination:                    auth0.String("https://example.com/saml"),
		CreateUPNClaim:                 auth0.Bool(false),
		PassthroughClaimsWithNoMapping: auth0.Bool(false),
		MapUnknownClaimsAsIs:           auth0.Bool(true),
		MapIdentities:                  auth0.Bool(false),
		TypedAttributes:                auth0.Bool(false),
		IncludeAttributeNameFormat:     auth0.Bool(false),
		LifetimeInSeconds:              auth0.Int(7200),
		SignatureAlgorithm:             auth0.String("rsa-sha256"),
		DigestAlgorithm:                auth0.String("sha256"),
		NameIdentifierFormat:           auth0.String("urn:oasis:names:tc:SAML:2.0:nameid-format:persistent"),
	}

	result := flattenClientAddonSAML2(addon)

	assert.Len(t, result, 1)
	flat, ok := result[0].(map[string]interface{})
	assert.True(t, ok, "expected result[0] to be a map[string]interface{}")

	assert.Equal(t, "https://example.com/saml", flat["recipient"])
	assert.Equal(t, "https://example.com/saml", flat["destination"])
	assert.Equal(t, false, flat["create_upn_claim"], "create_upn_claim")
	assert.Equal(t, false, flat["passthrough_claims_with_no_mapping"], "passthrough_claims_with_no_mapping")
	assert.Equal(t, true, flat["map_unknown_claims_as_is"], "map_unknown_claims_as_is")
	assert.Equal(t, false, flat["map_identities"], "map_identities")
	assert.Equal(t, false, flat["typed_attributes"], "typed_attributes")
	assert.Equal(t, false, flat["include_attribute_name_format"], "include_attribute_name_format")
	assert.Equal(t, 7200, flat["lifetime_in_seconds"], "lifetime_in_seconds")
	assert.Equal(t, "rsa-sha256", flat["signature_algorithm"], "signature_algorithm")
	assert.Equal(t, "sha256", flat["digest_algorithm"], "digest_algorithm")
	assert.Equal(t, "urn:oasis:names:tc:SAML:2.0:nameid-format:persistent", flat["name_identifier_format"], "name_identifier_format")
}

// TestFlattenClientGrant_ScopesOmittedWhenAllowAllScopes verifies that
// flattenClientGrant does not write "scopes" into state when allow_all_scopes
// is true. The Auth0 API returns scope:[] in that case; if we wrote that into
// state, terraform plan -generate-config-out would emit scopes = [] alongside
// allow_all_scopes = true — a combination the validator correctly rejects.
func TestFlattenClientGrant_ScopesOmittedWhenAllowAllScopes(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, NewGrantResource().Schema, map[string]interface{}{})

	grant := &management.ClientGrant{
		ClientID:       auth0.String("test-client-id"),
		Audience:       auth0.String("https://api.example.com"),
		Scope:          &[]string{},
		AllowAllScopes: auth0.Bool(true),
	}

	err := flattenClientGrant(resourceData, grant)
	assert.NoError(t, err)

	assert.Equal(t, true, resourceData.Get("allow_all_scopes"), "allow_all_scopes should be set")

	// Scopes must not be written when allow_all_scopes is true — it should
	// remain at its zero value (length 0) so it does not appear as a
	// non-empty value in generated configs.
	assert.Equal(t, 0, resourceData.Get("scopes.#"), "scopes should not be set when allow_all_scopes is true")
}

// TestFlattenClientGrant_ScopesSetWhenAllowAllScopesFalse verifies that
// flattenClientGrant writes scopes into state normally when allow_all_scopes
// is false.
func TestFlattenClientGrant_ScopesSetWhenAllowAllScopesFalse(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, NewGrantResource().Schema, map[string]interface{}{})

	grant := &management.ClientGrant{
		ClientID:       auth0.String("test-client-id"),
		Audience:       auth0.String("https://api.example.com"),
		Scope:          &[]string{"read:data", "write:data"},
		AllowAllScopes: auth0.Bool(false),
	}

	err := flattenClientGrant(resourceData, grant)
	assert.NoError(t, err)

	assert.Equal(t, false, resourceData.Get("allow_all_scopes"))
	assert.Equal(t, 2, resourceData.Get("scopes.#"))
}

// TestClientGrantScopesConflictWithAllowAll exercises the extracted predicate
// that backs validateClientGrant. The five cases cover the full truth table:
//
//   - allow_all_scopes=true  + scopes=null        → no conflict (valid: omit scopes)
//   - allow_all_scopes=true  + scopes=[]           → no conflict (valid: generated config)
//   - allow_all_scopes=true  + scopes=["read:foo"] → conflict   (invalid)
//   - allow_all_scopes=false + scopes=null         → no conflict (handled by separate guard)
//   - allow_all_scopes=false + scopes=["read:foo"] → no conflict (valid: explicit scopes)
func TestClientGrantScopesConflictWithAllowAll(t *testing.T) {
	t.Parallel()

	nullList := cty.NullVal(cty.List(cty.String))
	emptyList := cty.ListValEmpty(cty.String)
	nonEmptyList := cty.ListVal([]cty.Value{cty.StringVal("read:foo")})

	tests := []struct {
		name           string
		allowAllScopes bool
		scopes         cty.Value
		wantConflict   bool
	}{
		{
			name:           "allow_all_scopes=true, scopes omitted (null) — valid",
			allowAllScopes: true,
			scopes:         nullList,
			wantConflict:   false,
		},
		{
			name:           "allow_all_scopes=true, scopes=[] (empty list) — valid (generated config case)",
			allowAllScopes: true,
			scopes:         emptyList,
			wantConflict:   false,
		},
		{
			name:           "allow_all_scopes=true, scopes=[\"read:foo\"] — conflict",
			allowAllScopes: true,
			scopes:         nonEmptyList,
			wantConflict:   true,
		},
		{
			name:           "allow_all_scopes=false, scopes omitted (null) — no conflict",
			allowAllScopes: false,
			scopes:         nullList,
			wantConflict:   false,
		},
		{
			name:           "allow_all_scopes=false, scopes=[\"read:foo\"] — no conflict",
			allowAllScopes: false,
			scopes:         nonEmptyList,
			wantConflict:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := clientGrantScopesConflictWithAllowAll(tc.allowAllScopes, tc.scopes)
			assert.Equal(t, tc.wantConflict, got)
		})
	}
}

func TestFlattenClientIdentityAssertionAuthorizationGrant(t *testing.T) {
	t.Run("returns nil when grant is nil", func(t *testing.T) {
		assert.Nil(t, flattenClientIdentityAssertionAuthorizationGrant(nil))
	})

	t.Run("flattens active=true", func(t *testing.T) {
		result := flattenClientIdentityAssertionAuthorizationGrant(&management.IdentityAssertionAuthorizationGrant{
			Active: auth0.Bool(true),
		})

		assert.Len(t, result, 1)
		flat, ok := result[0].(map[string]interface{})
		assert.True(t, ok, "expected result[0] to be a map[string]interface{}")
		assert.Equal(t, true, flat["active"])
	})

	t.Run("flattens active=false", func(t *testing.T) {
		result := flattenClientIdentityAssertionAuthorizationGrant(&management.IdentityAssertionAuthorizationGrant{
			Active: auth0.Bool(false),
		})

		assert.Len(t, result, 1)
		flat, ok := result[0].(map[string]interface{})
		assert.True(t, ok, "expected result[0] to be a map[string]interface{}")
		assert.Equal(t, false, flat["active"])
	})
}
