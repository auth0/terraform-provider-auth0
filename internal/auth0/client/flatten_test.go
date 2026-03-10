package client

import (
	"testing"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
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

	// All omitted fields must reflect Auth0 defaults, not Go zero values.
	assert.Equal(t, samlDefault.createUPNClaim, flat["create_upn_claim"], "create_upn_claim")
	assert.Equal(t, samlDefault.passthroughClaimsWithNoMapping, flat["passthrough_claims_with_no_mapping"], "passthrough_claims_with_no_mapping")
	assert.Equal(t, samlDefault.mapUnknownClaimsAsIs, flat["map_unknown_claims_as_is"], "map_unknown_claims_as_is")
	assert.Equal(t, samlDefault.mapIdentities, flat["map_identities"], "map_identities")
	assert.Equal(t, samlDefault.typedAttributes, flat["typed_attributes"], "typed_attributes")
	assert.Equal(t, samlDefault.includeAttributeNameFormat, flat["include_attribute_name_format"], "include_attribute_name_format")
	assert.Equal(t, samlDefault.lifetimeInSeconds, flat["lifetime_in_seconds"], "lifetime_in_seconds")
	assert.Equal(t, samlDefault.signatureAlgorithm, flat["signature_algorithm"], "signature_algorithm")
	assert.Equal(t, samlDefault.digestAlgorithm, flat["digest_algorithm"], "digest_algorithm")
	assert.Equal(t, samlDefault.nameIdentifierFormat, flat["name_identifier_format"], "name_identifier_format")
}

// TestFlattenClientAddonSAML2_ExplicitValues verifies that explicitly set fields
// are used as-is and are not overridden by samlDefault.
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
