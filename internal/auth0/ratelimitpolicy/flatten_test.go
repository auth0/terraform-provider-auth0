package ratelimitpolicy

import (
	"encoding/json"
	"testing"

	"github.com/auth0/go-auth0/v2/management"
	"github.com/stretchr/testify/assert"
)

func TestFlattenConfiguration(t *testing.T) {
	t.Run("nil returns nil", func(t *testing.T) {
		assert.Nil(t, flattenRateLimitPolicyConfiguration(nil))
	})

	t.Run("empty union returns nil", func(t *testing.T) {
		assert.Nil(t, flattenRateLimitPolicyConfiguration(&management.RateLimitPolicyConfiguration{}))
	})
}

// TestFlattenConfigurationFromAPIResponse exercises the real deserialization path: the SDK's oneOf
// unmarshalling greedily matches the allow (Zero) variant for every response and drops the
// block/log/redirect fields into extra properties. Flatten must still recover limit and
// redirect_uri from there so they are not lost from state.
func TestFlattenConfigurationFromAPIResponse(t *testing.T) {
	unmarshal := func(t *testing.T, raw string) *management.RateLimitPolicyConfiguration {
		t.Helper()
		var cfg management.RateLimitPolicyConfiguration
		assert.NoError(t, json.Unmarshal([]byte(raw), &cfg))
		return &cfg
	}

	t.Run("block response keeps limit", func(t *testing.T) {
		out := flattenRateLimitPolicyConfiguration(unmarshal(t, `{"action":"block","limit":100}`))
		m := out[0].(map[string]interface{})
		assert.Equal(t, "block", m["action"])
		assert.Equal(t, 100, m["limit"])
		assert.Nil(t, m["redirect_uri"])
	})

	t.Run("log response keeps limit=0", func(t *testing.T) {
		out := flattenRateLimitPolicyConfiguration(unmarshal(t, `{"action":"log","limit":0}`))
		m := out[0].(map[string]interface{})
		assert.Equal(t, "log", m["action"])
		assert.Equal(t, 0, m["limit"])
	})

	t.Run("redirect response keeps limit and uri", func(t *testing.T) {
		out := flattenRateLimitPolicyConfiguration(unmarshal(t, `{"action":"redirect","limit":50,"redirect_uri":"https://example.com/blocked"}`))
		m := out[0].(map[string]interface{})
		assert.Equal(t, "redirect", m["action"])
		assert.Equal(t, 50, m["limit"])
		assert.Equal(t, "https://example.com/blocked", m["redirect_uri"])
	})

	t.Run("allow response has no limit or uri", func(t *testing.T) {
		out := flattenRateLimitPolicyConfiguration(unmarshal(t, `{"action":"allow"}`))
		m := out[0].(map[string]interface{})
		assert.Equal(t, "allow", m["action"])
		assert.Nil(t, m["limit"])
		assert.Nil(t, m["redirect_uri"])
	})
}
