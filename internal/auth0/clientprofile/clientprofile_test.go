package clientprofile

import (
	"encoding/json"
	"errors"
	"slices"
	"testing"

	"github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var nullString = cty.NullVal(cty.String)

var clientConfigurationKeys = []string{
	"refresh_token", "organization_usage", "organization_require_behavior",
	"organization_discovery_methods", "require_proof_of_possession",
}

// rawConfig builds a raw configuration object, defaulting every unspecified attribute to null. The
// client configuration attributes are given flat and are nested into the client_configuration block.
func rawConfig(overrides map[string]cty.Value) cty.Value {
	attrs := map[string]cty.Value{
		"name":                 cty.StringVal("profile"),
		"description":          nullString,
		"apply_policy":         cty.NullVal(cty.List(cty.EmptyObject)),
		"client_configuration": cty.NullVal(cty.List(cty.EmptyObject)),
	}
	configuration := map[string]cty.Value{}
	for _, key := range clientConfigurationKeys {
		configuration[key] = cty.NullVal(cty.List(cty.EmptyObject))
	}

	for k, v := range overrides {
		if slices.Contains(clientConfigurationKeys, k) {
			configuration[k] = v
		} else {
			attrs[k] = v
		}
	}
	attrs["client_configuration"] = cty.ListVal([]cty.Value{cty.ObjectVal(configuration)})

	return cty.ObjectVal(attrs)
}

// wrap wraps a value in the single-item block that mirrors the API's default_value wrapper.
func wrap(v cty.Value) cty.Value {
	return cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{"default_value": v})})
}

// optionalEnvelope is like envelope but yields a null block for a null value.
func optionalEnvelope(v cty.Value) cty.Value {
	if v.IsNull() {
		return cty.NullVal(cty.List(cty.EmptyObject))
	}

	return wrap(v)
}

func refreshTokenBlock(token, idle, leeway cty.Value) cty.Value {
	return cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
		"token_lifetime":      optionalEnvelope(token),
		"idle_token_lifetime": optionalEnvelope(idle),
		"leeway":              optionalEnvelope(leeway),
	})})
}

func marshal(t *testing.T, v interface{}) string {
	t.Helper()
	out, err := json.Marshal(v)
	require.NoError(t, err)

	return string(out)
}

func TestExpandCreate(t *testing.T) {
	t.Run("minimal sends empty configuration and default apply policy", func(t *testing.T) {
		got := marshal(t, expandClientProfileCreate(rawConfig(nil)))
		assert.JSONEq(t, `{
			"name":"profile",
			"client_configuration":{},
			"apply_policy":{"external_metadata_type":"cimd","external_metadata_created_by":"client"}
		}`, got)
	})

	t.Run("full configuration with zero and false values", func(t *testing.T) {
		raw := rawConfig(map[string]cty.Value{
			"description":                    cty.StringVal("desc"),
			"refresh_token":                  refreshTokenBlock(cty.NumberIntVal(1000), cty.NumberIntVal(500), cty.NumberIntVal(0)),
			"organization_usage":             wrap(cty.StringVal("require")),
			"organization_require_behavior":  wrap(cty.StringVal("pre_login_prompt")),
			"organization_discovery_methods": wrap(cty.SetVal([]cty.Value{cty.StringVal("organization_name"), cty.StringVal("email")})),
			"require_proof_of_possession":    wrap(cty.False),
		})
		got := marshal(t, expandClientProfileCreate(raw))
		assert.JSONEq(t, `{
			"name":"profile",
			"description":"desc",
			"client_configuration":{
				"refresh_token":{
					"token_lifetime":{"default_value":1000},
					"idle_token_lifetime":{"default_value":500},
					"leeway":{"default_value":0}
				},
				"organization_usage":{"default_value":"require"},
				"organization_require_behavior":{"default_value":"pre_login_prompt"},
				"organization_discovery_methods":{"default_value":["email","organization_name"]},
				"require_proof_of_possession":{"default_value":false}
			},
			"apply_policy":{"external_metadata_type":"cimd","external_metadata_created_by":"client"}
		}`, got)
	})

	t.Run("unset refresh token sub-fields are omitted", func(t *testing.T) {
		raw := rawConfig(map[string]cty.Value{
			"refresh_token": refreshTokenBlock(cty.NumberIntVal(1000), cty.NullVal(cty.Number), cty.NullVal(cty.Number)),
		})
		got := marshal(t, expandClientProfileCreate(raw).ClientConfiguration)
		assert.JSONEq(t, `{"refresh_token":{"token_lifetime":{"default_value":1000}}}`, got)
	})
}

func TestExpandUpdate(t *testing.T) {
	t.Run("removed description is sent as null", func(t *testing.T) {
		got := marshal(t, expandClientProfileUpdate(rawConfig(nil)))
		assert.JSONEq(t, `{
			"name":"profile",
			"description":null,
			"client_configuration":{"refresh_token":{}},
			"apply_policy":{"external_metadata_type":"cimd","external_metadata_created_by":"client"}
		}`, got)
	})

	t.Run("empty description is sent as an empty string", func(t *testing.T) {
		got := marshal(t, expandClientProfileUpdate(rawConfig(map[string]cty.Value{"description": cty.StringVal("")})))
		assert.Contains(t, got, `"description":""`)
	})

	t.Run("removed refresh token block is sent as an empty object", func(t *testing.T) {
		got := marshal(t, expandClientProfileUpdate(rawConfig(nil)).ClientConfiguration)
		assert.JSONEq(t, `{"refresh_token":{}}`, got)
	})

	t.Run("false and zero are sent", func(t *testing.T) {
		raw := rawConfig(map[string]cty.Value{
			"description":                 cty.StringVal("d"),
			"refresh_token":               refreshTokenBlock(cty.NullVal(cty.Number), cty.NullVal(cty.Number), cty.NumberIntVal(0)),
			"require_proof_of_possession": wrap(cty.False),
		})
		got := marshal(t, expandClientProfileUpdate(raw).ClientConfiguration)
		assert.JSONEq(t, `{
			"refresh_token":{"leeway":{"default_value":0}},
			"require_proof_of_possession":{"default_value":false}
		}`, got)
	})
}

func TestFlatten(t *testing.T) {
	resource := NewResource()

	flatten := func(t *testing.T, body string) *schema.ResourceData {
		t.Helper()
		var profile management.ClientProfile
		require.NoError(t, json.Unmarshal([]byte(body), &profile))

		data := resource.TestResourceData()
		data.SetId(profile.GetID())
		require.False(t, flattenClientProfile(data, &profile).HasError())

		return data
	}

	t.Run("full body with server-derived lifetimes", func(t *testing.T) {
		data := flatten(t, `{
			"id":"cpr_123","name":"profile","description":"desc",
			"client_configuration":{
				"refresh_token":{"token_lifetime":{"default_value":3600},"idle_token_lifetime":{"default_value":1800},"leeway":{"default_value":0}},
				"organization_usage":{"default_value":"allow"},
				"organization_require_behavior":{"default_value":"pre_login_prompt"},
				"organization_discovery_methods":{"default_value":["email"]},
				"require_proof_of_possession":{"default_value":true}
			},
			"apply_policy":{"external_metadata_type":"cimd","external_metadata_created_by":"client"},
			"created_at":"2026-10-01T10:00:00.000Z","updated_at":"2026-10-02T10:00:00.000Z"
		}`)

		assert.Equal(t, "desc", data.Get("description"))
		assert.Equal(t, 3600, data.Get("client_configuration.0.refresh_token.0.token_lifetime.0.default_value"))
		assert.Equal(t, 1800, data.Get("client_configuration.0.refresh_token.0.idle_token_lifetime.0.default_value"))
		assert.Equal(t, 0, data.Get("client_configuration.0.refresh_token.0.leeway.0.default_value"))
		assert.Equal(t, "allow", data.Get("client_configuration.0.organization_usage.0.default_value"))
		assert.Equal(t, "pre_login_prompt", data.Get("client_configuration.0.organization_require_behavior.0.default_value"))
		assert.Equal(t, 1, data.Get("client_configuration.0.organization_discovery_methods.0.default_value.#"))
		assert.Equal(t, true, data.Get("client_configuration.0.require_proof_of_possession.0.default_value"))
		assert.Equal(t, "cimd", data.Get("apply_policy.0.external_metadata_type"))
		assert.Equal(t, "client", data.Get("apply_policy.0.external_metadata_created_by"))
		assert.Equal(t, "2026-10-01T10:00:00Z", data.Get("created_at"))
	})

	t.Run("tolerates nil timestamps and envelopes", func(t *testing.T) {
		data := flatten(t, `{"id":"cpr_123","name":"profile","client_configuration":{},"apply_policy":null}`)

		assert.Equal(t, "profile", data.Get("name"))
		assert.Equal(t, "", data.Get("created_at"))
		assert.Equal(t, "", data.Get("updated_at"))
		assert.Equal(t, 0, data.Get("client_configuration.0.refresh_token.#"))
		assert.Equal(t, 0, data.Get("apply_policy.#"))
		assert.Equal(t, 0, data.Get("client_configuration.0.organization_usage.#"))
	})

	t.Run("empty refresh token object is treated as absent", func(t *testing.T) {
		data := flatten(t, `{"id":"cpr_123","name":"p","client_configuration":{"refresh_token":{}},"apply_policy":null}`)
		assert.Equal(t, 0, data.Get("client_configuration.0.refresh_token.#"))
	})
}

func TestCheckClientProfileConfig(t *testing.T) {
	t.Run("valid lifetimes", func(t *testing.T) {
		raw := rawConfig(map[string]cty.Value{
			"refresh_token": refreshTokenBlock(cty.NumberIntVal(1000), cty.NumberIntVal(1000), cty.NullVal(cty.Number)),
		})
		assert.NoError(t, checkClientProfileConfig(raw))
	})

	t.Run("idle lifetime above token lifetime", func(t *testing.T) {
		raw := rawConfig(map[string]cty.Value{
			"refresh_token": refreshTokenBlock(cty.NumberIntVal(1000), cty.NumberIntVal(2000), cty.NullVal(cty.Number)),
		})
		assert.ErrorContains(t, checkClientProfileConfig(raw), "must not exceed")
	})

	t.Run("unknown lifetime is skipped", func(t *testing.T) {
		raw := rawConfig(map[string]cty.Value{
			"refresh_token": refreshTokenBlock(cty.UnknownVal(cty.Number), cty.NumberIntVal(2000), cty.NullVal(cty.Number)),
		})
		assert.NoError(t, checkClientProfileConfig(raw))
	})

	t.Run("discovery methods need pre_login_prompt", func(t *testing.T) {
		methods := cty.SetVal([]cty.Value{cty.StringVal("email")})

		raw := rawConfig(map[string]cty.Value{"organization_discovery_methods": wrap(methods)})
		assert.Error(t, checkClientProfileConfig(raw))

		raw = rawConfig(map[string]cty.Value{
			"organization_discovery_methods": wrap(methods),
			"organization_require_behavior":  wrap(cty.StringVal("no_prompt")),
		})
		assert.Error(t, checkClientProfileConfig(raw))

		raw = rawConfig(map[string]cty.Value{
			"organization_discovery_methods": wrap(methods),
			"organization_require_behavior":  wrap(cty.StringVal("pre_login_prompt")),
		})
		assert.NoError(t, checkClientProfileConfig(raw))

		raw = rawConfig(map[string]cty.Value{
			"organization_discovery_methods": wrap(methods),
			"organization_require_behavior":  wrap(cty.UnknownVal(cty.String)),
		})
		assert.NoError(t, checkClientProfileConfig(raw))
	})
}

func TestMapCreateError(t *testing.T) {
	tooMany := &management.ForbiddenError{Body: map[string]interface{}{"errorCode": "too_many_entities"}}
	assert.ErrorContains(t, mapCreateError(tooMany), "already exists")

	other := &management.ForbiddenError{Body: map[string]interface{}{"errorCode": "insufficient_scope"}}
	assert.Equal(t, error(other), mapCreateError(other))

	plain := errors.New("boom")
	assert.Equal(t, plain, mapCreateError(plain))
}

func TestSchemaInternalValidate(t *testing.T) {
	assert.NoError(t, NewResource().InternalValidate(nil, true))
}
