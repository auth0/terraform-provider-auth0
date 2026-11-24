package customdomain

import (
	"testing"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestFlattenCustomDomainList(t *testing.T) {
	data := schema.TestResourceDataRaw(t, NewCustomDomainsDataSource().Schema, nil)

	customDomains := []*management.CustomDomain{
		{
			Domain:    auth0.String("default.auth.example.com"),
			IsDefault: auth0.Bool(true),
		},
		{
			Domain:    auth0.String("other.auth.example.com"),
			IsDefault: auth0.Bool(false),
		},
		{
			Domain: auth0.String("unset.auth.example.com"),
		},
	}

	assert.NoError(t, flattenCustomDomainList(data, customDomains))

	assert.Equal(t, 3, data.Get("custom_domains.#"))
	assert.Equal(t, "default.auth.example.com", data.Get("custom_domains.0.domain"))
	assert.Equal(t, true, data.Get("custom_domains.0.is_default"))
	assert.Equal(t, false, data.Get("custom_domains.1.is_default"))
	assert.Equal(t, false, data.Get("custom_domains.2.is_default"))
}
