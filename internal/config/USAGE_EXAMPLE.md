# Using Both v1 and v2 Management Clients

## Overview

The `Config` struct now supports both the v1 and v2 Auth0 Go SDK management clients simultaneously. This allows you to gradually migrate from v1 to v2.

## Accessing the Clients

### In Resource/Data Source Code

```go
func resourceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    config := m.(*config.Config)
    
    // Use v1 API client (existing)
    apiV1 := config.GetAPI()
    user, err := apiV1.User.Read(id)
    
    // Use v2 API client (new)
    apiV2 := config.GetAPIv2()
    // Example: apiV2.Users.Get(ctx, id)
    
    return nil
}
```

## Key Differences Between v1 and v2 SDK

### v1 SDK (current)
```go
// Import
import "github.com/auth0/go-auth0/management"

// Usage
api := config.GetAPI()
user, err := api.User.Read(id)
```

### v2 SDK (new)
```go
// Import
import managementv2 "github.com/auth0/go-auth0/v2/management/client"

// Usage
api := config.GetAPIv2()
user, err := api.Users.Get(ctx, id)
```

## Authentication

Both clients are automatically configured with the same authentication credentials:
- API Token (if provided)
- Client Credentials (Client ID + Secret)
- Client Credentials with Audience
- Client Assertion (Private Key JWT)

The `authenticationOptionV2` function in `config.go` handles the v2 authentication setup.

## Migration Strategy

1. **Gradual Migration**: Keep both clients available during migration
2. **Test Thoroughly**: Test v2 implementations alongside v1
3. **Resource by Resource**: Migrate one resource at a time
4. **Fallback Available**: Can always fall back to v1 if issues arise

## Example: Migrating a Resource

```go
// Before (v1)
func resourceClientRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    api := m.(*config.Config).GetAPI()
    client, err := api.Client.Read(d.Id())
    if err != nil {
        return diag.FromErr(err)
    }
    // ... process client
}

// After (v2)
func resourceClientRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    api := m.(*config.Config).GetAPIv2()
    client, err := api.Clients.Get(ctx, d.Id())
    if err != nil {
        return diag.FromErr(err)
    }
    // ... process client
}
```

## Notes

- The v2 SDK uses context in all API calls (following Go best practices)
- The v2 SDK has improved type safety and better error handling
- Both clients share the same HTTP client with retry logic
- Both clients use the same authentication configuration

