package error

import (
	"errors"
	"fmt"
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

// TestV2ErrorCodeMalformedBodies covers scope item E1: every malformed Body
// permutation must fail closed to "" rather than panicking or false-positiving.
func TestV2ErrorCodeMalformedBodies(t *testing.T) {
	var testCases = []struct {
		name string
		body interface{}
	}{
		{"nil body", nil},
		{"string body", "not a map"},
		{"slice body", []interface{}{"a", "b"}},
		{"int body", 403},
		{"bool body", true},
		{"typed struct body", struct{ ErrorCode string }{ErrorCode: "insufficient_entitlement"}},
		{"empty map", map[string]interface{}{}},
		{"missing errorCode key", map[string]interface{}{"statusCode": 403, "error": "Forbidden"}},
		{"errorCode is int", map[string]interface{}{"errorCode": 123}},
		{"errorCode is nil", map[string]interface{}{"errorCode": nil}},
		{"errorCode is bool", map[string]interface{}{"errorCode": true}},
		{"errorCode is nested map", map[string]interface{}{"errorCode": map[string]interface{}{"x": "y"}}},
		{"errorCode is slice", map[string]interface{}{"errorCode": []string{"insufficient_entitlement"}}},
		{"wrong-cased key ErrorCode", map[string]interface{}{"ErrorCode": "insufficient_entitlement"}},
		{"wrong-cased key errorcode", map[string]interface{}{"errorcode": "insufficient_entitlement"}},
		{"map[string]string body", map[string]string{"errorCode": "insufficient_entitlement"}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := &managementv2.ForbiddenError{Body: testCase.body}

			assert.Equal(t, "", v2ErrorCode(err))
			assert.False(t, IsInsufficientEntitlement(err))
			assert.False(t, IsInsufficientScope(err))
		})
	}
}

// TestV2ErrorCodeExactMatchBoundary covers scope item B1: the matcher is an
// exact string comparison, so near-miss variants must not match.
func TestV2ErrorCodeExactMatchBoundary(t *testing.T) {
	var testCases = []struct {
		name      string
		errorCode string
	}{
		{"uppercase", "INSUFFICIENT_ENTITLEMENT"},
		{"hyphenated", "insufficient-entitlement"},
		{"title case", "Insufficient_Entitlement"},
		{"leading whitespace", " insufficient_entitlement"},
		{"trailing whitespace", "insufficient_entitlement "},
		{"plural", "insufficient_entitlements"},
		{"substring only", "entitlement"},
		{"empty string", ""},
		{"unrecognised 403 code", "some_other_403"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := &managementv2.ForbiddenError{
				Body: map[string]interface{}{"errorCode": testCase.errorCode},
			}

			assert.False(t, IsInsufficientEntitlement(err))
			assert.False(t, IsInsufficientScope(err))
		})
	}
}

// TestV2ErrorCodeWrappedErrors covers scope item E2: errors.As must unwrap
// through fmt.Errorf %w chains so SDK/middleware wrapping does not defeat
// detection.
func TestV2ErrorCodeWrappedErrors(t *testing.T) {
	t.Run("single-level wrap preserves entitlement detection", func(t *testing.T) {
		forbidden := &managementv2.ForbiddenError{
			Body: map[string]interface{}{"errorCode": "insufficient_entitlement"},
		}
		wrapped := fmt.Errorf("calling bot detection: %w", forbidden)

		assert.Equal(t, "insufficient_entitlement", v2ErrorCode(wrapped))
		assert.True(t, IsInsufficientEntitlement(wrapped))
		assert.False(t, IsInsufficientScope(wrapped))
	})

	t.Run("multi-level wrap preserves scope detection", func(t *testing.T) {
		forbidden := &managementv2.ForbiddenError{
			Body: map[string]interface{}{"errorCode": "insufficient_scope"},
		}
		wrapped := fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", forbidden))

		assert.True(t, IsInsufficientScope(wrapped))
		assert.False(t, IsInsufficientEntitlement(wrapped))
	})

	t.Run("wrapped non-Forbidden error does not match", func(t *testing.T) {
		wrapped := fmt.Errorf("outer: %w", errors.New("boom"))

		assert.Equal(t, "", v2ErrorCode(wrapped))
		assert.False(t, IsInsufficientEntitlement(wrapped))
		assert.False(t, IsInsufficientScope(wrapped))
	})
}

// TestV2ErrorCodeNilError covers scope item E3: nil input must not panic.
func TestV2ErrorCodeNilError(t *testing.T) {
	assert.Equal(t, "", v2ErrorCode(nil))
	assert.False(t, IsInsufficientEntitlement(nil))
	assert.False(t, IsInsufficientScope(nil))
}

// TestV2ErrorCodeLiveWireFormat covers scope items B1 and N4 by asserting
// against the exact 403 bodies returned by the Management API for a
// non-entitled tenant, including the backend's own copy-paste bug where the
// captcha endpoint's message mentions "bot detection".
func TestV2ErrorCodeLiveWireFormat(t *testing.T) {
	t.Run("bot detection 403 body", func(t *testing.T) {
		err := &managementv2.ForbiddenError{
			Body: map[string]interface{}{
				"statusCode": float64(403),
				"error":      "Forbidden",
				"message":    "Please upgrade your subscription to use bot detection",
				"errorCode":  "insufficient_entitlement",
			},
		}

		assert.True(t, IsInsufficientEntitlement(err))
		assert.False(t, IsInsufficientScope(err))
	})

	t.Run("captcha 403 body (backend message says bot detection)", func(t *testing.T) {
		err := &managementv2.ForbiddenError{
			Body: map[string]interface{}{
				"statusCode": float64(403),
				"error":      "Forbidden",
				"message":    "Please upgrade your subscription to use bot detection",
				"errorCode":  "insufficient_entitlement",
			},
		}

		assert.True(t, IsInsufficientEntitlement(err))
	})
}
