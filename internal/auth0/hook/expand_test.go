package hook

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/stretchr/testify/assert"
)

func TestHookNameRegexp(t *testing.T) {
	for givenHookName, expectedError := range map[string]bool{
		"my-hook-1":                 false,
		"hook 2 name with spaces":   false,
		" hook with a space prefix": true,
		"hook with a space suffix ": true,
		" ":                         true,
		"   ":                       true,
	} {
		validationResult := validateHookName()(givenHookName, cty.Path{cty.GetAttrStep{Name: "name"}})
		assert.Equal(t, expectedError, validationResult.HasError())
	}
}
