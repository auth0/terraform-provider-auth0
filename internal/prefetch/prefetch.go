package prefetch

import (
	"context"
	"fmt"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const defaultPageSize = 50

// GetClient fetches a client by ID using the opportunistic pre-fetch heuristic.
//
// It checks the cache first. On a miss, if pages remain, it acquires the
// per-type fetch mutex, re-checks the cache (another goroutine may have just
// populated it), then fetches the next page and stores all results. After the
// fetch the mutex is released and the cache is re-checked. If the resource is
// still not found it falls back to a direct API call.
func GetClient(ctx context.Context, cache *Cache, api *management.Management, id string) (*management.Client, error) {
	// Check cache first.
	if v, ok := cache.getEntry(resourceTypeClient, id); ok {
		cache.recordHit(resourceTypeClient)
		return v.(*management.Client), nil
	}

	// Fetch the next page if there are pages remaining.
	if !cache.isExhausted(resourceTypeClient) {
		// Serialise page fetches so parallel goroutines don't all fetch the
		// same page concurrently.
		cache.lockFetch(resourceTypeClient)

		// Re-check inside the lock — another goroutine may have just fetched
		// and cached this resource while we were waiting.
		if v, ok := cache.getEntry(resourceTypeClient, id); ok {
			cache.unlockFetch(resourceTypeClient)
			cache.recordHit(resourceTypeClient)
			return v.(*management.Client), nil
		}

		if !cache.isExhausted(resourceTypeClient) {
			page := cache.nextPage(resourceTypeClient)

			list, err := api.Client.List(ctx,
				management.Page(page),
				management.PerPage(defaultPageSize),
				management.IncludeTotals(true),
			)
			if err != nil {
				cache.unlockFetch(resourceTypeClient)
				cache.recordMiss(resourceTypeClient)
				return api.Client.Read(ctx, id)
			}

			items := make(map[string]interface{}, len(list.Clients))
			for _, c := range list.Clients {
				items[c.GetClientID()] = c
			}
			nowExhausted := cache.setEntries(resourceTypeClient, items, list.HasNext())
			cache.recordPageFetch(resourceTypeClient)

			tflog.Trace(ctx, "prefetch: fetched client page", map[string]interface{}{
				"page":     page,
				"count":    len(list.Clients),
				"has_more": list.HasNext(),
				"total":    list.Total,
			})

			if nowExhausted {
				s := cache.Summary(resourceTypeClient)
				tflog.Debug(ctx, "prefetch: client cache exhausted", map[string]interface{}{
					"pages_fetched": s.PagesFetched,
					"cached":        s.Cached,
					"hits":          s.Hits,
					"misses":        s.Misses,
					"hit_rate_pct":  fmt.Sprintf("%.1f", s.HitRate()),
				})
			}
		}

		cache.unlockFetch(resourceTypeClient)

		// Re-check after the page load.
		if v, ok := cache.getEntry(resourceTypeClient, id); ok {
			cache.recordHit(resourceTypeClient)
			return v.(*management.Client), nil
		}
	}

	// Fall back to a direct single-resource fetch.
	cache.recordMiss(resourceTypeClient)
	return api.Client.Read(ctx, id)
}

// GetClientGrant fetches a client grant by ID using the opportunistic pre-fetch heuristic.
//
// It checks the cache first. On a miss, if pages remain, it acquires the
// per-type fetch mutex, re-checks the cache (another goroutine may have just
// populated it), then fetches the next page and stores all results. After the
// fetch the mutex is released and the cache is re-checked. If the resource is
// still not found it falls back to a direct API call.
func GetClientGrant(ctx context.Context, cache *Cache, api *management.Management, id string) (*management.ClientGrant, error) {
	// Check cache first.
	if v, ok := cache.getEntry(resourceTypeClientGrant, id); ok {
		cache.recordHit(resourceTypeClientGrant)
		return v.(*management.ClientGrant), nil
	}

	// Fetch the next page if there are pages remaining.
	if !cache.isExhausted(resourceTypeClientGrant) {
		// Serialise page fetches so parallel goroutines don't all fetch the
		// same page concurrently.
		cache.lockFetch(resourceTypeClientGrant)

		// Re-check inside the lock — another goroutine may have just fetched
		// and cached this resource while we were waiting.
		if v, ok := cache.getEntry(resourceTypeClientGrant, id); ok {
			cache.unlockFetch(resourceTypeClientGrant)
			cache.recordHit(resourceTypeClientGrant)
			return v.(*management.ClientGrant), nil
		}

		if !cache.isExhausted(resourceTypeClientGrant) {
			page := cache.nextPage(resourceTypeClientGrant)

			list, err := api.ClientGrant.List(ctx,
				management.Page(page),
				management.PerPage(defaultPageSize),
				management.IncludeTotals(true),
			)
			if err != nil {
				cache.unlockFetch(resourceTypeClientGrant)
				cache.recordMiss(resourceTypeClientGrant)
				return api.ClientGrant.Read(ctx, id)
			}

			items := make(map[string]interface{}, len(list.ClientGrants))
			for _, g := range list.ClientGrants {
				items[g.GetID()] = g
			}
			nowExhausted := cache.setEntries(resourceTypeClientGrant, items, list.HasNext())
			cache.recordPageFetch(resourceTypeClientGrant)

			tflog.Trace(ctx, "prefetch: fetched client_grant page", map[string]interface{}{
				"page":     page,
				"count":    len(list.ClientGrants),
				"has_more": list.HasNext(),
				"total":    list.Total,
			})

			if nowExhausted {
				s := cache.Summary(resourceTypeClientGrant)
				tflog.Debug(ctx, "prefetch: client_grant cache exhausted", map[string]interface{}{
					"pages_fetched": s.PagesFetched,
					"cached":        s.Cached,
					"hits":          s.Hits,
					"misses":        s.Misses,
					"hit_rate_pct":  fmt.Sprintf("%.1f", s.HitRate()),
				})
			}
		}

		cache.unlockFetch(resourceTypeClientGrant)

		// Re-check after the page load.
		if v, ok := cache.getEntry(resourceTypeClientGrant, id); ok {
			cache.recordHit(resourceTypeClientGrant)
			return v.(*management.ClientGrant), nil
		}
	}

	// Fall back to a direct single-resource fetch.
	cache.recordMiss(resourceTypeClientGrant)
	return api.ClientGrant.Read(ctx, id)
}
