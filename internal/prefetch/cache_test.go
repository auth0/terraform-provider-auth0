package prefetch

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCache(t *testing.T) {
	t.Run("it returns an initialised cache", func(t *testing.T) {
		c := NewCache()
		assert.NotNil(t, c)
	})
}

func TestCache_GetEntry(t *testing.T) {
	t.Run("it returns false on a cache miss", func(t *testing.T) {
		c := NewCache()
		v, ok := c.getEntry(resourceTypeClient, "does-not-exist")
		assert.False(t, ok)
		assert.Nil(t, v)
	})

	t.Run("it returns the stored value on a cache hit", func(t *testing.T) {
		c := NewCache()
		sentinel := struct{ ID string }{ID: "abc"}
		c.setEntries(resourceTypeClient, map[string]interface{}{"abc": &sentinel}, false)

		v, ok := c.getEntry(resourceTypeClient, "abc")
		assert.True(t, ok)
		assert.Equal(t, &sentinel, v)
	})
}

func TestCache_SetEntries(t *testing.T) {
	t.Run("it advances the page cursor after each call", func(t *testing.T) {
		c := NewCache()
		assert.Equal(t, 0, c.nextPage(resourceTypeClient))

		c.setEntries(resourceTypeClient, map[string]interface{}{}, true)
		assert.Equal(t, 1, c.nextPage(resourceTypeClient))

		c.setEntries(resourceTypeClient, map[string]interface{}{}, true)
		assert.Equal(t, 2, c.nextPage(resourceTypeClient))
	})

	t.Run("it marks the type exhausted when hasMore is false", func(t *testing.T) {
		c := NewCache()
		assert.False(t, c.isExhausted(resourceTypeClient))

		c.setEntries(resourceTypeClient, map[string]interface{}{}, false)
		assert.True(t, c.isExhausted(resourceTypeClient))
	})

	t.Run("it does not mark the type exhausted when hasMore is true", func(t *testing.T) {
		c := NewCache()
		c.setEntries(resourceTypeClient, map[string]interface{}{}, true)
		assert.False(t, c.isExhausted(resourceTypeClient))
	})
}

func TestCache_ResourceTypeIsolation(t *testing.T) {
	t.Run("it tracks client and client_grant pages independently", func(t *testing.T) {
		c := NewCache()
		c.setEntries(resourceTypeClient, map[string]interface{}{}, true)
		c.setEntries(resourceTypeClientGrant, map[string]interface{}{}, false)

		assert.False(t, c.isExhausted(resourceTypeClient))
		assert.True(t, c.isExhausted(resourceTypeClientGrant))
		assert.Equal(t, 1, c.nextPage(resourceTypeClient))
		assert.Equal(t, 1, c.nextPage(resourceTypeClientGrant))
	})
}

func TestCache_ConcurrentAccess(t *testing.T) {
	t.Run("it does not race on concurrent reads and writes", func(t *testing.T) {
		c := NewCache()
		var wg sync.WaitGroup

		for i := 0; i < 50; i++ {
			wg.Add(2)
			go func() {
				defer wg.Done()
				c.setEntries(resourceTypeClient, map[string]interface{}{"x": "v"}, true)
			}()
			go func() {
				defer wg.Done()
				c.getEntry(resourceTypeClient, "x")
			}()
		}

		wg.Wait()
	})
}
