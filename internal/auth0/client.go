// Package auth0 wraps the go-auth0 v2 SDK and provides a single helper to
// build an authenticated Management client from the provider configuration.
package auth0

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"time"

	mgmtclient "github.com/auth0/go-auth0/v2/management/client"
	"github.com/auth0/go-auth0/v2/management/option"
	"golang.org/x/oauth2/clientcredentials"
)

// providerName is reported via the Auth0-Client header.
const providerName = "terraform-provider-auth0"

// AuthMode determines how the provider authenticates to the Auth0 Management API.
type AuthMode int

const (
	// AuthModeUnknown means no credentials were supplied.
	AuthModeUnknown AuthMode = iota
	// AuthModeStaticToken uses a pre-issued Management API access token.
	AuthModeStaticToken
	// AuthModeClientCredentials uses an M2M client_id + client_secret pair.
	AuthModeClientCredentials
	// AuthModePrivateKeyJWT uses an M2M client_id + private key (JWT assertion).
	AuthModePrivateKeyJWT
)

// String returns a human-readable label for the auth mode (used in logs).
func (m AuthMode) String() string {
	switch m {
	case AuthModeStaticToken:
		return "api_token"
	case AuthModeClientCredentials:
		return "client_credentials"
	case AuthModePrivateKeyJWT:
		return "private_key_jwt"
	default:
		return "unknown"
	}
}

// Config carries the resolved provider configuration after env-var fallback.
type Config struct {
	Domain   string
	Audience string

	// Static token auth.
	APIToken string

	// Client credentials auth (one of: secret, or private key JWT).
	ClientID     string
	ClientSecret string

	// Private Key JWT auth.
	ClientAssertionPrivateKey string // PEM-encoded private key.
	ClientAssertionAlgorithm  string // e.g. RS256.

	// Misc.
	Debug bool

	// ProviderVersion / TerraformVersion are reported via the Auth0-Client
	// telemetry header (set by the provider during Configure).
	ProviderVersion  string
	TerraformVersion string
}

// Mode returns the authentication mode implied by the supplied credentials.
func (c Config) Mode() AuthMode {
	switch {
	case c.APIToken != "":
		return AuthModeStaticToken
	case c.ClientID != "" && c.ClientAssertionPrivateKey != "":
		return AuthModePrivateKeyJWT
	case c.ClientID != "" && c.ClientSecret != "":
		return AuthModeClientCredentials
	default:
		return AuthModeUnknown
	}
}

// authenticationOption returns the single SDK option that wires the chosen
// auth mode into the Management client. Mirrors `authenticationOptionV2` in
// the legacy SDKv2 provider.
//
// IMPORTANT: the SDK's client-credentials options capture the supplied context
// by *pointer* and use it later (lazily) to perform the OAuth token exchange.
// If we passed the Configure-scoped ctx, it would be cancelled by the time the
// first CRUD call runs, the OAuth fetch would fail with "context canceled",
// the SDK would silently drop the Authorization header, and the API would
// reply 401 "Missing authentication". So we deliberately use a long-lived
// background context here.
func authenticationOption(_ context.Context, c Config) (option.RequestOption, error) {
	tokenCtx := context.Background()

	switch c.Mode() {
	case AuthModeStaticToken:
		return option.WithToken(c.APIToken), nil

	case AuthModeClientCredentials:
		if c.Audience != "" {
			return option.WithClientCredentialsAndAudience(tokenCtx, c.ClientID, c.ClientSecret, c.Audience), nil
		}
		return option.WithClientCredentials(tokenCtx, c.ClientID, c.ClientSecret), nil

	case AuthModePrivateKeyJWT:
		alg := c.ClientAssertionAlgorithm
		if alg == "" {
			alg = "RS256"
		}
		if c.Audience != "" {
			return option.WithClientCredentialsPrivateKeyJwtAndAudience(
				tokenCtx, c.ClientID, c.ClientAssertionPrivateKey, alg, c.Audience,
			), nil
		}
		return option.WithClientCredentialsPrivateKeyJwt(
			tokenCtx, c.ClientID, c.ClientAssertionPrivateKey, alg,
		), nil

	default:
		return nil, fmt.Errorf("auth0: no credentials supplied – set api_token, client_id+client_secret, or client_id+client_assertion_private_key")
	}
}

// userAgent returns the User-Agent string sent on every request, e.g.
// "Terraform/1.7.5 (+https://www.terraform.io) terraform-provider-auth0/2.0.0".
func userAgent(providerVersion, tfVersion string) string {
	if providerVersion == "" {
		providerVersion = "dev"
	}
	if tfVersion == "" {
		tfVersion = "0.0.0"
	}
	return fmt.Sprintf(
		"Terraform/%s (+https://www.terraform.io) %s/%s (%s; %s)",
		tfVersion, providerName, providerVersion, runtime.GOOS, runtime.GOARCH,
	)
}

// httpClientWithRetries returns the *http.Client used by the SDK. We give it
// a saner timeout than the Go default (which is none) so a hung Auth0 API
// can't wedge a Terraform run forever.
func httpClientWithRetries() *http.Client {
	return &http.Client{Timeout: 60 * time.Second}
}

// NewManagement builds a *management.Management client from the resolved Config.
//
// We also eagerly perform a token fetch (for non-static-token modes) so that
// invalid credentials surface as a clear provider configuration error rather
// than a misleading 401 "Missing authentication" on the first resource call —
// the SDK silently drops the Authorization header when its TokenSource returns
// an error.
func NewManagement(ctx context.Context, c Config) (*mgmtclient.Management, error) {
	if c.Domain == "" {
		return nil, fmt.Errorf("auth0: domain is required")
	}

	authOpt, err := authenticationOption(ctx, c)
	if err != nil {
		return nil, err
	}

	opts := []option.RequestOption{
		authOpt,
		option.WithUserAgent(userAgent(c.ProviderVersion, c.TerraformVersion)),
		option.WithAuth0ClientEnvEntry(providerName, c.ProviderVersion),
		option.WithHTTPClient(httpClientWithRetries()),
		option.WithDebug(c.Debug),
	}

	mgmt, err := mgmtclient.New(c.Domain, opts...)
	if err != nil {
		return nil, err
	}

	if err := verifyCredentials(ctx, c); err != nil {
		return nil, err
	}

	return mgmt, nil
}

// verifyCredentials forces a token fetch for the chosen auth mode. We rebuild
// the option independently of the SDK because the SDK swallows token errors
// in its `ToHeader()` path. Static-token mode has nothing to verify.
func verifyCredentials(ctx context.Context, c Config) error {
	if c.Mode() == AuthModeStaticToken || c.Mode() == AuthModeUnknown {
		return nil
	}

	tokenURL := "https://" + stripScheme(c.Domain) + "/oauth/token"
	audience := c.Audience
	if audience == "" {
		audience = "https://" + stripScheme(c.Domain) + "/api/v2/"
	}

	switch c.Mode() {
	case AuthModeClientCredentials:
		cfg := &clientcredentials.Config{
			ClientID:       c.ClientID,
			ClientSecret:   c.ClientSecret,
			TokenURL:       tokenURL,
			EndpointParams: url.Values{"audience": []string{audience}},
		}
		if _, err := cfg.TokenSource(ctx).Token(); err != nil {
			return fmt.Errorf("auth0: failed to obtain Management API token via client_credentials at %s (audience=%s): %w", tokenURL, audience, err)
		}
	case AuthModePrivateKeyJWT:
		// PKJWT verification is more involved; defer to first API call. We at
		// least DNS-resolve the token URL to catch obvious typos.
		if _, err := url.Parse(tokenURL); err != nil {
			return fmt.Errorf("auth0: invalid token URL %q: %w", tokenURL, err)
		}
	}
	return nil
}

func stripScheme(domain string) string {
	if i := indexOf(domain, "//"); i != -1 {
		return domain[i+2:]
	}
	return domain
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
