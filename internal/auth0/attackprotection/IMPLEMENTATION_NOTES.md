# Bot Detection and CAPTCHA Implementation Notes

## Overview

This implementation adds support for Bot Detection and CAPTCHA configuration in Auth0's Attack Protection module, based on PR #634 in the go-auth0 repository.

## Implementation Status

✅ **Completed:**
- Schema definitions for `bot_detection` and `captcha` blocks in the attack protection resource
- Expand functions to convert Terraform configuration to SDK types
- Flatten functions to convert SDK types to Terraform state
- Updated CRUD operations (Create, Read, Update, Delete)
- Support for multiple CAPTCHA providers:
  - Arkose Labs
  - Auth0 (legacy)
  - Auth0 V2
  - Google reCAPTCHA v2
  - Google reCAPTCHA Enterprise
  - hCaptcha
  - Friendly Captcha

⏳ **Pending:**
- go-auth0 SDK update (PR #634 needs to be merged and released)
- Integration tests
- Documentation updates

## Required SDK Types

The implementation expects the following types from `github.com/auth0/go-auth0/management`:

### Bot Detection Types
```go
type BotDetection struct {
    AllowList             []string
    MonitoringModeEnabled *bool
}
```

### CAPTCHA Types
```go
type Captcha struct {
    ActiveProviderID          *string
    ArkoseConfig              *ArkoseCaptchaConfig
    Auth0Config               *Auth0CaptchaConfig
    Auth0V2Config             *Auth0V2CaptchaConfig
    RecaptchaV2Config         *RecaptchaV2CaptchaConfig
    RecaptchaEnterpriseConfig *RecaptchaEnterpriseCaptchaConfig
    HcaptchaConfig            *HcaptchaCaptchaConfig
    FriendlyCaptchaConfig     *FriendlyCaptchaConfig
}

type ArkoseCaptchaConfig struct {
    SiteKey         *string
    Secret          *string
    ClientSubdomain *string
    VerifySubdomain *string
    FailOpen        *bool
}

type Auth0CaptchaConfig struct {
    // Empty struct for legacy Auth0 CAPTCHA
}

type Auth0V2CaptchaConfig struct {
    FailOpen *bool
}

type RecaptchaV2CaptchaConfig struct {
    SiteKey  *string
    Secret   *string
    FailOpen *bool
}

type RecaptchaEnterpriseCaptchaConfig struct {
    SiteKey  *string
    Secret   *string
    FailOpen *bool
}

type HcaptchaCaptchaConfig struct {
    SiteKey  *string
    Secret   *string
    FailOpen *bool
}

type FriendlyCaptchaConfig struct {
    SiteKey  *string
    Secret   *string
    FailOpen *bool
}
```

### API Methods
The implementation expects these methods on `api.AttackProtection`:
- `GetBotDetection(ctx context.Context) (*BotDetection, error)`
- `UpdateBotDetection(ctx context.Context, bd *BotDetection) error`
- `GetCaptcha(ctx context.Context) (*Captcha, error)`
- `UpdateCaptcha(ctx context.Context, c *Captcha) error`

## Usage Example

```hcl
resource "auth0_attack_protection" "my_protection" {
  # Existing configurations...
  breached_password_detection { ... }
  brute_force_protection { ... }
  suspicious_ip_throttling { ... }

  # New Bot Detection configuration
  bot_detection {
    allowlist = [
      "192.168.1.0/24",
      "10.0.0.1"
    ]
    monitoring_mode_enabled = true
  }

  # New CAPTCHA configuration
  captcha {
    active_provider_id = "recaptcha_v2"
    
    recaptcha_v2_config {
      site_key  = "your-recaptcha-site-key"
      secret    = "your-recaptcha-secret"
      fail_open = false
    }
  }
}
```

## Next Steps

1. **Wait for SDK Update**: Monitor PR #634 in the go-auth0 repository
2. **Update go.mod**: Once the PR is merged and released, update the dependency:
   ```bash
   go get github.com/auth0/go-auth0/v2@latest
   ```
3. **Run Tests**: Execute integration tests to verify functionality
4. **Update Documentation**: Generate Terraform documentation with:
   ```bash
   make docs
   ```

## Testing

Once the SDK is available, test with:

```bash
# Run unit tests
go test ./internal/auth0/attackprotection/...

# Run acceptance tests (requires Auth0 tenant)
TF_ACC=1 go test ./internal/auth0/attackprotection/... -v -run TestAccAttackProtection
```

## Error Handling

The implementation includes graceful error handling for backward compatibility:
- If bot detection API is not available, it continues without error
- If CAPTCHA API is not available, it continues without error
- This allows the provider to work with older Auth0 tenants

## Security Considerations

- CAPTCHA secrets are marked as sensitive in the schema
- Secrets are not logged or exposed in Terraform output
- Use environment variables or secure secret management for CAPTCHA credentials

## References

- go-auth0 PR #634: https://github.com/auth0/go-auth0/pull/634
- Auth0 Attack Protection Docs: https://auth0.com/docs/secure/attack-protection
- Auth0 Bot Detection Docs: https://auth0.com/docs/secure/attack-protection/bot-detection

