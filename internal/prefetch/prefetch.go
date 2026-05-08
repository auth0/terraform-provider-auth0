package prefetch

import (
	"context"

	"github.com/auth0/go-auth0/management"
)

const defaultPageSize = 50

// GetClient fetches a client by ID using the opportunistic pre-fetch heuristic.
//
// It checks the cache first. On a miss, if pages remain, it fetches the next
// page and stores all results before checking the cache again. If the resource
// is still not found after the page fetch, it falls back to a direct API call.
func GetClient(ctx context.Context, cache *Cache, api *management.Management, id string) (*management.Client, error) {
	// Check cache first.
	if v, ok := cache.getEntry(resourceTypeClient, id); ok {
		return v.(*management.Client), nil
	}

	// Fetch the next page if there are pages remaining.
	if !cache.isExhausted(resourceTypeClient) {
		page := cache.nextPage(resourceTypeClient)

		list, err := api.Client.List(ctx,
			management.Page(page),
			management.PerPage(defaultPageSize),
		)
		if err != nil {
			return api.Client.Read(ctx, id)
		}

		items := make(map[string]interface{}, len(list.Clients))
		for _, c := range list.Clients {
			items[c.GetClientID()] = c
		}
		cache.setEntries(resourceTypeClient, items, list.HasNext())

		// Re-check after the page load.
		if v, ok := cache.getEntry(resourceTypeClient, id); ok {
			return v.(*management.Client), nil
		}
	}

	// Fall back to a direct single-resource fetch.
	return api.Client.Read(ctx, id)
}

// GetClientGrant fetches a client grant by ID using the opportunistic pre-fetch heuristic.
//
// It checks the cache first. On a miss, if pages remain, it fetches the next
// page and stores all results before checking the cache again. If the resource
// is still not found after the page fetch, it falls back to a direct API call.
func GetClientGrant(ctx context.Context, cache *Cache, api *management.Management, id string) (*management.ClientGrant, error) {
	// Check cache first.
	if v, ok := cache.getEntry(resourceTypeClientGrant, id); ok {
		return v.(*management.ClientGrant), nil
	}

	// Fetch the next page if there are pages remaining.
	if !cache.isExhausted(resourceTypeClientGrant) {
		page := cache.nextPage(resourceTypeClientGrant)

		list, err := api.ClientGrant.List(ctx,
			management.Page(page),
			management.PerPage(defaultPageSize),
		)
		if err != nil {
			return api.ClientGrant.Read(ctx, id)
		}

		items := make(map[string]interface{}, len(list.ClientGrants))
		for _, g := range list.ClientGrants {
			items[g.GetID()] = g
		}
		cache.setEntries(resourceTypeClientGrant, items, list.HasNext())

		// Re-check after the page load.
		if v, ok := cache.getEntry(resourceTypeClientGrant, id); ok {
			return v.(*management.ClientGrant), nil
		}
	}

	// Fall back to a direct single-resource fetch.
	return api.ClientGrant.Read(ctx, id)
}
