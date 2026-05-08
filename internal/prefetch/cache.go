// Package prefetch implements an opportunistic pre-fetch cache for Auth0 resources.
// It uses page-based list endpoints to batch API calls and reduce total request count.
package prefetch

import (
	"sync"
)

// resourceType identifies the kind of resource stored in the cache.
type resourceType string

const (
	resourceTypeClient      resourceType = "client"
	resourceTypeClientGrant resourceType = "client_grant"
)

// cacheKey uniquely identifies a cached resource.
type cacheKey struct {
	kind resourceType
	id   string
}

// pageState tracks pagination for a given resource type.
type pageState struct {
	nextPage  int
	exhausted bool
}

// Cache is an in-memory, thread-safe store for pre-fetched Auth0 resources.
// It tracks both the resource values and the page-fetch cursor per resource type.
type Cache struct {
	mu         sync.RWMutex
	entries    map[cacheKey]interface{}
	pageStates map[resourceType]*pageState
}

// NewCache returns an initialised *Cache.
func NewCache() *Cache {
	return &Cache{
		entries:    make(map[cacheKey]interface{}),
		pageStates: make(map[resourceType]*pageState),
	}
}

// getEntry returns the cached value and whether it was found.
func (c *Cache) getEntry(kind resourceType, id string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	v, ok := c.entries[cacheKey{kind: kind, id: id}]
	return v, ok
}

// setEntries stores a batch of values in the cache and advances the page cursor.
func (c *Cache) setEntries(kind resourceType, items map[string]interface{}, hasMore bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for id, v := range items {
		c.entries[cacheKey{kind: kind, id: id}] = v
	}

	ps := c.getOrInitPageState(kind)
	ps.nextPage++
	if !hasMore {
		ps.exhausted = true
	}
}

// nextPage returns the next page number to fetch for the given resource type.
func (c *Cache) nextPage(kind resourceType) int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.getOrInitPageState(kind).nextPage
}

// isExhausted reports whether all pages have already been fetched.
func (c *Cache) isExhausted(kind resourceType) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.getOrInitPageState(kind).exhausted
}



// getOrInitPageState returns the pageState for kind, initialising it if necessary.
// Callers must hold c.mu (read or write).
func (c *Cache) getOrInitPageState(kind resourceType) *pageState {
	ps, ok := c.pageStates[kind]
	if !ok {
		ps = &pageState{}
		c.pageStates[kind] = ps
	}
	return ps
}
