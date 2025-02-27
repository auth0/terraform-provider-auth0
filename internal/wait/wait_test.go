package wait

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUntil(t *testing.T) {
	t.Run("it returns an error when given a negative pollingInterval", func(t *testing.T) {
		err := Until(-1, 1, func() (bool, error) {
			return true, nil
		})
		assert.Error(t, err)
	})

	t.Run("it returns an error when given a negative pollingCount", func(t *testing.T) {
		err := Until(1, -1, func() (bool, error) {
			return true, nil
		})
		assert.Error(t, err)
	})

	t.Run("it returns an error when pollingFunc returns an error", func(t *testing.T) {
		err := Until(1, 1, func() (bool, error) {
			return false, fmt.Errorf("error foo")
		})
		assert.ErrorContains(t, err, "error foo")
	})

	t.Run("it returns an error when pollingFunc never returns true", func(t *testing.T) {
		err := Until(1, 1, func() (bool, error) {
			return false, nil
		})
		assert.ErrorContains(t, err, "timeout")
	})

	t.Run("it returns nil when pollingFunc returns true", func(t *testing.T) {
		err := Until(1, 1, func() (bool, error) {
			return true, nil
		})
		assert.NoError(t, err)
	})
}
