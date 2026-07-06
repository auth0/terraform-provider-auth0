package ratelimitpolicy

import (
	"testing"

	"github.com/auth0/go-auth0/v2/management"
	"github.com/stretchr/testify/assert"
)

func TestFlattenConfiguration(t *testing.T) {
	t.Run("nil returns nil", func(t *testing.T) {
		assert.Nil(t, flattenRateLimitPolicyConfiguration(nil))
	})

	t.Run("allow -> only action", func(t *testing.T) {
		out := flattenRateLimitPolicyConfiguration(&management.RateLimitPolicyConfiguration{
			RateLimitPolicyConfigurationZero: &management.RateLimitPolicyConfigurationZero{
				Action: management.RateLimitPolicyConfigurationZeroActionAllow,
			},
		})
		m := out[0].(map[string]interface{})
		assert.Equal(t, "allow", m["action"])
		assert.Nil(t, m["limit"])
		assert.Nil(t, m["redirect_uri"])
	})

	t.Run("block -> action+limit", func(t *testing.T) {
		out := flattenRateLimitPolicyConfiguration(&management.RateLimitPolicyConfiguration{
			RateLimitPolicyConfigurationOne: &management.RateLimitPolicyConfigurationOne{
				Action: management.RateLimitPolicyConfigurationOneActionBlock,
				Limit:  100,
			},
		})
		m := out[0].(map[string]interface{})
		assert.Equal(t, "block", m["action"])
		assert.Equal(t, 100, m["limit"])
		assert.Nil(t, m["redirect_uri"])
	})

	t.Run("redirect -> action+limit+uri", func(t *testing.T) {
		out := flattenRateLimitPolicyConfiguration(&management.RateLimitPolicyConfiguration{
			RateLimitPolicyConfigurationAction: &management.RateLimitPolicyConfigurationAction{
				Action:      management.RateLimitPolicyConfigurationActionActionRedirect,
				Limit:       50,
				RedirectURI: "https://example.com/blocked",
			},
		})
		m := out[0].(map[string]interface{})
		assert.Equal(t, "redirect", m["action"])
		assert.Equal(t, 50, m["limit"])
		assert.Equal(t, "https://example.com/blocked", m["redirect_uri"])
	})
}
