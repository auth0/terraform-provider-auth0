---
page_title: Opportunistic Pre-fetch
description: |-
  Reduce API call count for large Auth0 deployments by enabling opportunistic pre-fetch mode.
---

# Opportunistic Pre-fetch

Large Auth0 deployments — for example, those with 100+ clients or 100+ client grants — can hit
excessive latency and rate-limit exhaustion during `terraform plan` runs. By default, the provider
fetches each `auth0_client` and `auth0_client_grant` resource individually, resulting in one API
call per resource.

**Opportunistic pre-fetch mode** batches those upstream calls using the Auth0 page-based list
endpoints, dramatically reducing total API call count.

## How it works

When pre-fetch is enabled, any individual resource read follows this heuristic:

1. Check an in-memory cache (keyed by resource ID).
2. On a cache miss, if pages remain unfetched, fetch the **next page** (50 resources per page),
   store all results in the cache, and advance the page cursor.
3. Re-check the cache — return the resource if it is now present.
4. Fall back to a direct single-resource API call if the resource was still not found.

This approach balances latency (the provider never waits for all pages before returning) with
opportunistic bulk loading across a `terraform plan` run.

## Enabling pre-fetch

Add `prefetch = true` to your provider block:

```terraform
provider "auth0" {
  domain        = var.auth0_domain
  client_id     = var.auth0_client_id
  client_secret = var.auth0_client_secret
  prefetch      = true
}
```

Alternatively, set the environment variable:

```shell
export AUTH0_PREFETCH=true
```

Pre-fetch is **opt-in** and defaults to `false`. No changes to individual resource configurations
are required.

## Scope

The initial release covers `auth0_client` and `auth0_client_grant`. The `auth0_client_credentials`
resource benefits automatically because it reads client data through the same code path.

Additional resource types (for example, `auth0_resource_server`) may be added in future releases
without a breaking change.
