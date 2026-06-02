package ratelimitpolicy

import (
	"testing"

	"github.com/auth0/go-auth0"
	"github.com/stretchr/testify/assert"
)

func TestExpandConfigurationUnion_Create(t *testing.T) {
	t.Run("action only -> allow variant", func(t *testing.T) {
		cfg := expandConfigurationUnion("allow", nil, nil)
		assert.NotNil(t, cfg.RateLimitPolicyConfigurationZero)
		assert.Nil(t, cfg.RateLimitPolicyConfigurationOne)
		assert.Nil(t, cfg.RateLimitPolicyConfigurationAction)
		assert.Equal(t, "allow", string(cfg.RateLimitPolicyConfigurationZero.GetAction()))
	})

	t.Run("action+limit -> limited variant", func(t *testing.T) {
		cfg := expandConfigurationUnion("block", auth0.Int(100), nil)
		assert.NotNil(t, cfg.RateLimitPolicyConfigurationOne)
		assert.Equal(t, "block", string(cfg.RateLimitPolicyConfigurationOne.GetAction()))
		assert.Equal(t, 100, cfg.RateLimitPolicyConfigurationOne.GetLimit())
	})

	t.Run("limit=0 is forwarded, not treated as absent", func(t *testing.T) {
		cfg := expandConfigurationUnion("log", auth0.Int(0), nil)
		assert.NotNil(t, cfg.RateLimitPolicyConfigurationOne)
		assert.Equal(t, 0, cfg.RateLimitPolicyConfigurationOne.GetLimit())
	})

	t.Run("action+limit+uri -> redirect variant", func(t *testing.T) {
		cfg := expandConfigurationUnion("redirect", auth0.Int(50), auth0.String("https://example.com/blocked"))
		assert.NotNil(t, cfg.RateLimitPolicyConfigurationAction)
		assert.Equal(t, 50, cfg.RateLimitPolicyConfigurationAction.GetLimit())
		assert.Equal(t, "https://example.com/blocked", cfg.RateLimitPolicyConfigurationAction.GetRedirectURI())
	})

	t.Run("invalid combo is forwarded as-is, not silently dropped", func(t *testing.T) {
		// action "allow" with a limit is invalid per the API, but the provider must forward it
		// so the API returns the error rather than the provider quietly discarding the limit.
		cfg := expandConfigurationUnion("allow", auth0.Int(100), nil)
		assert.NotNil(t, cfg.RateLimitPolicyConfigurationOne)
		assert.Equal(t, "allow", string(cfg.RateLimitPolicyConfigurationOne.GetAction()))
		assert.Equal(t, 100, cfg.RateLimitPolicyConfigurationOne.GetLimit())
	})

	t.Run("empty configuration -> nil", func(t *testing.T) {
		assert.Nil(t, expandConfigurationUnion("", nil, nil))
	})
}

func TestExpandPatchConfigurationUnion(t *testing.T) {
	cfg := expandPatchConfigurationUnion("redirect", auth0.Int(10), auth0.String("https://x.example.com"))
	assert.NotNil(t, cfg.PatchRateLimitPolicyConfigurationRequestContentAction)
	assert.Equal(t, "redirect", string(cfg.PatchRateLimitPolicyConfigurationRequestContentAction.GetAction()))
	assert.Equal(t, 10, cfg.PatchRateLimitPolicyConfigurationRequestContentAction.GetLimit())

	assert.Nil(t, expandPatchConfigurationUnion("", nil, nil))
}
