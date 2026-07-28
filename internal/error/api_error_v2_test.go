package error

import (
	"errors"
	"testing"

	managementv2 "github.com/auth0/go-auth0/v2/management"
	"github.com/stretchr/testify/assert"
)

func TestV2ErrorCode(t *testing.T) {
	t.Run("returns errorCode from valid ForbiddenError", func(t *testing.T) {
		err := &managementv2.ForbiddenError{
			Body: map[string]interface{}{
				"statusCode": 403,
				"error":      "Forbidden",
				"message":    "Please upgrade your subscription",
				"errorCode":  "insufficient_entitlement",
			},
		}

		code := v2ErrorCode(err)
		assert.Equal(t, "insufficient_entitlement", code)
	})

	t.Run("returns empty string for non-ForbiddenError", func(t *testing.T) {
		err := errors.New("some other error")

		code := v2ErrorCode(err)
		assert.Equal(t, "", code)
	})

	t.Run("returns empty string when Body is not a map", func(t *testing.T) {
		err := &managementv2.ForbiddenError{
			Body: "not a map",
		}

		code := v2ErrorCode(err)
		assert.Equal(t, "", code)
	})

	t.Run("returns empty string when errorCode key is missing", func(t *testing.T) {
		err := &managementv2.ForbiddenError{
			Body: map[string]interface{}{
				"statusCode": 403,
				"error":      "Forbidden",
			},
		}

		code := v2ErrorCode(err)
		assert.Equal(t, "", code)
	})

	t.Run("returns empty string when errorCode is not a string", func(t *testing.T) {
		err := &managementv2.ForbiddenError{
			Body: map[string]interface{}{
				"statusCode": 403,
				"error":      "Forbidden",
				"errorCode":  123,
			},
		}

		code := v2ErrorCode(err)
		assert.Equal(t, "", code)
	})
}

func TestIsInsufficientScope(t *testing.T) {
	t.Run("returns true for insufficient_scope error", func(t *testing.T) {
		err := &managementv2.ForbiddenError{
			Body: map[string]interface{}{
				"errorCode": "insufficient_scope",
			},
		}

		assert.True(t, IsInsufficientScope(err))
	})

	t.Run("returns false for insufficient_entitlement error", func(t *testing.T) {
		err := &managementv2.ForbiddenError{
			Body: map[string]interface{}{
				"errorCode": "insufficient_entitlement",
			},
		}

		assert.False(t, IsInsufficientScope(err))
	})

	t.Run("returns false for non-ForbiddenError", func(t *testing.T) {
		err := errors.New("some other error")

		assert.False(t, IsInsufficientScope(err))
	})

	t.Run("returns false for malformed ForbiddenError", func(t *testing.T) {
		err := &managementv2.ForbiddenError{
			Body: "not a map",
		}

		assert.False(t, IsInsufficientScope(err))
	})
}

func TestIsInsufficientEntitlement(t *testing.T) {
	t.Run("returns true for insufficient_entitlement error", func(t *testing.T) {
		err := &managementv2.ForbiddenError{
			Body: map[string]interface{}{
				"errorCode": "insufficient_entitlement",
			},
		}

		assert.True(t, IsInsufficientEntitlement(err))
	})

	t.Run("returns false for insufficient_scope error", func(t *testing.T) {
		err := &managementv2.ForbiddenError{
			Body: map[string]interface{}{
				"errorCode": "insufficient_scope",
			},
		}

		assert.False(t, IsInsufficientEntitlement(err))
	})

	t.Run("returns false for non-ForbiddenError", func(t *testing.T) {
		err := errors.New("some other error")

		assert.False(t, IsInsufficientEntitlement(err))
	})

	t.Run("returns false for malformed ForbiddenError", func(t *testing.T) {
		err := &managementv2.ForbiddenError{
			Body: "not a map",
		}

		assert.False(t, IsInsufficientEntitlement(err))
	})
}
