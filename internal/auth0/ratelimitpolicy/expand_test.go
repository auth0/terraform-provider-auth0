package ratelimitpolicy

import (
	"testing"

	"github.com/auth0/go-auth0"
	"github.com/stretchr/testify/assert"
)

func TestExpandConfigurationUnion_Create(t *testing.T) {
	t.Run("allow -> allow variant, no limit/uri", func(t *testing.T) {
		cfg := expandConfigurationUnion("allow", nil, nil)
		assert.NotNil(t, cfg.RateLimitPolicyConfigurationZero)
		assert.Nil(t, cfg.RateLimitPolicyConfigurationOne)
		assert.Nil(t, cfg.RateLimitPolicyConfigurationAction)
		assert.Equal(t, "allow", string(cfg.RateLimitPolicyConfigurationZero.GetAction()))
	})

	t.Run("allow ignores a generated limit=0", func(t *testing.T) {
		// Generated config emits limit = 0 for allow; the action drives the variant, so it's dropped.
		cfg := expandConfigurationUnion("allow", auth0.Int(0), nil)
		assert.NotNil(t, cfg.RateLimitPolicyConfigurationZero)
		assert.Nil(t, cfg.RateLimitPolicyConfigurationOne)
		assert.Nil(t, cfg.RateLimitPolicyConfigurationAction)
	})

	t.Run("block -> limited variant", func(t *testing.T) {
		cfg := expandConfigurationUnion("block", auth0.Int(100), nil)
		assert.NotNil(t, cfg.RateLimitPolicyConfigurationOne)
		assert.Nil(t, cfg.RateLimitPolicyConfigurationZero)
		assert.Nil(t, cfg.RateLimitPolicyConfigurationAction)
		assert.Equal(t, "block", string(cfg.RateLimitPolicyConfigurationOne.GetAction()))
		assert.Equal(t, 100, cfg.RateLimitPolicyConfigurationOne.GetLimit())
	})

	t.Run("log with limit=0 is preserved", func(t *testing.T) {
		cfg := expandConfigurationUnion("log", auth0.Int(0), nil)
		assert.NotNil(t, cfg.RateLimitPolicyConfigurationOne)
		assert.Equal(t, "log", string(cfg.RateLimitPolicyConfigurationOne.GetAction()))
		assert.Equal(t, 0, cfg.RateLimitPolicyConfigurationOne.GetLimit())
	})

	t.Run("redirect -> redirect variant", func(t *testing.T) {
		cfg := expandConfigurationUnion("redirect", auth0.Int(50), auth0.String("https://example.com/blocked"))
		assert.NotNil(t, cfg.RateLimitPolicyConfigurationAction)
		assert.Nil(t, cfg.RateLimitPolicyConfigurationZero)
		assert.Nil(t, cfg.RateLimitPolicyConfigurationOne)
		assert.Equal(t, 50, cfg.RateLimitPolicyConfigurationAction.GetLimit())
		assert.Equal(t, "https://example.com/blocked", cfg.RateLimitPolicyConfigurationAction.GetRedirectURI())
	})

	t.Run("empty action -> nil", func(t *testing.T) {
		assert.Nil(t, expandConfigurationUnion("", nil, nil))
	})
}

func TestExpandPatchConfigurationUnion(t *testing.T) {
	t.Run("redirect -> redirect variant", func(t *testing.T) {
		cfg := expandPatchConfigurationUnion("redirect", auth0.Int(10), auth0.String("https://x.example.com"))
		assert.NotNil(t, cfg.PatchRateLimitPolicyConfigurationRequestContentAction)
		assert.Equal(t, "redirect", string(cfg.PatchRateLimitPolicyConfigurationRequestContentAction.GetAction()))
		assert.Equal(t, 10, cfg.PatchRateLimitPolicyConfigurationRequestContentAction.GetLimit())
	})

	t.Run("allow ignores a generated limit=0", func(t *testing.T) {
		cfg := expandPatchConfigurationUnion("allow", auth0.Int(0), nil)
		assert.NotNil(t, cfg.PatchRateLimitPolicyConfigurationRequestContentZero)
		assert.Nil(t, cfg.PatchRateLimitPolicyConfigurationRequestContentOne)
		assert.Nil(t, cfg.PatchRateLimitPolicyConfigurationRequestContentAction)
	})

	t.Run("empty action -> nil", func(t *testing.T) {
		assert.Nil(t, expandPatchConfigurationUnion("", nil, nil))
	})
}

func TestCheckRateLimitPolicyConfiguration(t *testing.T) {
	t.Run("allow is valid regardless of limit or uri", func(t *testing.T) {
		// Inapplicable fields are ignored for allow, not rejected, so generated config still plans.
		assert.NoError(t, checkRateLimitPolicyConfiguration("allow", nil, nil))
		assert.NoError(t, checkRateLimitPolicyConfiguration("allow", auth0.Int(0), nil))
		assert.NoError(t, checkRateLimitPolicyConfiguration("allow", auth0.Int(100), auth0.String("https://example.com")))
	})

	t.Run("block with limit is valid", func(t *testing.T) {
		assert.NoError(t, checkRateLimitPolicyConfiguration("block", auth0.Int(100), nil))
	})

	t.Run("block with limit=0 is valid", func(t *testing.T) {
		assert.NoError(t, checkRateLimitPolicyConfiguration("block", auth0.Int(0), nil))
	})

	t.Run("log with limit is valid", func(t *testing.T) {
		assert.NoError(t, checkRateLimitPolicyConfiguration("log", auth0.Int(250), nil))
	})

	t.Run("block without limit is rejected", func(t *testing.T) {
		err := checkRateLimitPolicyConfiguration("block", nil, nil)
		assert.ErrorContains(t, err, "`limit` is required when `action` is \"block\"")
	})

	t.Run("redirect with limit and uri is valid", func(t *testing.T) {
		assert.NoError(t, checkRateLimitPolicyConfiguration("redirect", auth0.Int(50), auth0.String("https://example.com/blocked")))
	})

	t.Run("redirect without limit is rejected", func(t *testing.T) {
		err := checkRateLimitPolicyConfiguration("redirect", nil, auth0.String("https://example.com/blocked"))
		assert.ErrorContains(t, err, "`limit` is required when `action` is \"redirect\"")
	})

	t.Run("redirect without uri is rejected", func(t *testing.T) {
		err := checkRateLimitPolicyConfiguration("redirect", auth0.Int(50), nil)
		assert.ErrorContains(t, err, "`redirect_uri` is required when `action` is \"redirect\"")
	})

	t.Run("empty action (no configuration) is a no-op", func(t *testing.T) {
		assert.NoError(t, checkRateLimitPolicyConfiguration("", nil, nil))
	})
}
