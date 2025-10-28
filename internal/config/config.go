package config

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/rehttp"
	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/meta"
	"github.com/zalando/go-keyring"

	"github.com/auth0/terraform-provider-auth0/internal/mutex"
)

const providerName = "Terraform-Provider-Auth0"    // #nosec G101
const secretAccessToken = "Auth0 CLI Access Token" // #nosec G101

// Access tokens have no size limit, but should be smaller than (50*2048) bytes.
// The max number of loops safeguards against infinite loops, however unlikely.
const secretAccessTokenMaxChunks = 50

var version = "dev"

// cliConfigPath is the default path to the auth0-cli config file.
var cliConfigPath = path.Join(os.Getenv("HOME"), ".config", "auth0", "config.json")

// SetCliConfigPath sets the path to the auth0-cli config file.
func SetCliConfigPath(p string) {
	cliConfigPath = p
}

// CliConfig holds CLI configuration settings.
type CliConfig struct {
	Tenants map[string]struct {
		AccessToken string    `json:"access_token,omitempty"`
		ExpiresAt   time.Time `json:"expires_at"`
	} `json:"tenants"`
}

// Config is the type used for the
// *schema.Provider meta parameter.
type Config struct {
	api   *management.Management
	mutex *mutex.KeyValue
}

// New instantiates a new Config.
func New(apiClient *management.Management) *Config {
	return &Config{
		api:   apiClient,
		mutex: mutex.New(),
	}
}

// GetAPI fetches an instance of the *management.Management client.
func (c *Config) GetAPI() *management.Management {
	return c.api
}

// GetMutex fetches an instance of the *mutex.KeyValue.
func (c *Config) GetMutex() *mutex.KeyValue {
	return c.mutex
}

type ProviderConfig struct {
	Debug                     bool
	Domain                    string
	ClientID                  string
	ClientSecret              string
	ApiToken                  string
	Audience                  string
	ClientAssertionPrivateKey string
	ClientAssertionSigningAlg string
	CustomDomainHeader        string
}

func ParseResourceConfigData(data *schema.ResourceData) (ProviderConfig, diag.Diagnostics) {
	cfg := ProviderConfig{
		Domain:                    data.Get("domain").(string),
		Debug:                     data.Get("debug").(bool),
		ClientID:                  data.Get("client_id").(string),
		ClientSecret:              data.Get("client_secret").(string),
		ApiToken:                  data.Get("api_token").(string),
		Audience:                  data.Get("audience").(string),
		ClientAssertionPrivateKey: data.Get("client_assertion_private_key").(string),
		ClientAssertionSigningAlg: data.Get("client_assertion_signing_alg").(string),
		CustomDomainHeader:        data.Get("custom_domain_header").(string),
	}

	dynamicCredentials := data.Get("dynamic_credentials").(bool)
	cliLogin := data.Get("cli_login").(bool)

	// Helper to return missing domain diagnostic.
	missingDomain := func(vars string) diag.Diagnostics {
		return diag.Diagnostics{{
			Severity: diag.Error,
			Summary:  "Missing required configuration",
			Detail:   fmt.Sprintf("The 'AUTH0_DOMAIN' must be specified along with %s.", vars),
		}}
	}

	switch {
	case dynamicCredentials:
		if cfg.Domain == "" {
			return ProviderConfig{}, missingDomain("'AUTH0_DYNAMIC_CREDENTIALS'")
		}

	case cliLogin:
		if cfg.Domain == "" {
			return ProviderConfig{}, missingDomain("'AUTH0_CLI_LOGIN'")
		}

		// Fetch and validate CLI token.
		tempToken, diags := fetchAndValidateCLIToken(cfg.Domain)
		if diags != nil {
			return ProviderConfig{}, diags
		}

		// Set the apiToken to the valid tempToken.
		cfg.ApiToken = tempToken

	case cfg.ApiToken != "":
		if cfg.Domain == "" {
			return ProviderConfig{}, missingDomain("'AUTH0_API_TOKEN'")
		}

	case cfg.ClientID != "":
		var (
			hasSecret    = cfg.ClientSecret != ""
			hasAssertion = cfg.ClientAssertionPrivateKey != "" && cfg.ClientAssertionSigningAlg != ""
		)

		if !hasSecret && !hasAssertion {
			return ProviderConfig{}, diag.Diagnostics{{
				Severity: diag.Error,
				Summary:  "Missing required configuration",
				Detail: "When 'AUTH0_CLIENT_ID' is provided, either 'AUTH0_CLIENT_SECRET' or both " +
					"'AUTH0_CLIENT_ASSERTION_PRIVATE_KEY' and 'AUTH0_CLIENT_ASSERTION_SIGNING_ALG' must also be specified.",
			}}
		}

		if cfg.Domain == "" {
			switch {
			case hasSecret:
				return ProviderConfig{}, missingDomain("'AUTH0_CLIENT_ID' and 'AUTH0_CLIENT_SECRET'")
			case hasAssertion:
				return ProviderConfig{}, missingDomain("'AUTH0_CLIENT_ID', 'AUTH0_CLIENT_ASSERTION_PRIVATE_KEY', and 'AUTH0_CLIENT_ASSERTION_SIGNING_ALG'")
			}
		}

	default:
		return ProviderConfig{}, diag.Diagnostics{{
			Severity: diag.Error,
			Summary:  "Missing environment variables",
			Detail: "AUTH0_DOMAIN is required. Then, configure either AUTH0_API_TOKEN, " +
				"or AUTH0_CLIENT_ID and AUTH0_CLIENT_SECRET, " +
				"or AUTH0_CLIENT_ID, AUTH0_CLIENT_ASSERTION_PRIVATE_KEY, and AUTH0_CLIENT_ASSERTION_SIGNING_ALG, " +
				"or enable CLI login with AUTH0_CLI_LOGIN=true.",
		}}
	}

	return cfg, nil
}

// ConfigureProvider will configure the *schema.Provider so that
// *management.Management client and *mutex.KeyValue is stored
// and passed into the subsequent resources as the meta parameter.
func ConfigureProvider(terraformVersion *string) schema.ConfigureContextFunc {
	return func(_ context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
		config, d := ParseResourceConfigData(data)
		if d != nil {
			return nil, d
		}

		apiClient, err := management.New(config.Domain,
			authenticationOption(config),
			management.WithDebug(config.Debug),
			management.WithUserAgent(userAgent(terraformVersion)),
			management.WithAuth0ClientEnvEntry(providerName, version),
			management.WithNoRetries(),
			management.WithClient(customClientWithRetries()),
			management.WithCustomDomainHeader(config.CustomDomainHeader))

		if err != nil {
			return nil, diag.FromErr(err)
		}

		return New(apiClient), nil
	}
}

func fetchAndValidateCLIToken(domain string) (string, diag.Diagnostics) {
	var tempToken string
	for i := 0; i < secretAccessTokenMaxChunks; i++ {
		chunk, err := keyring.Get(fmt.Sprintf("%s %d", secretAccessToken, i), domain)
		if errors.Is(err, keyring.ErrNotFound) || err != nil {
			break
		}
		tempToken += chunk
	}

	// If no token was found, try to fetch token from CLI config file.
	if tempToken == "" {
		configToken, err := getAccessTokenFromCliConfigFile(domain)
		if err != nil {
			return "", diag.Diagnostics{{
				Severity: diag.Error,
				Summary:  "Failed to read config.json file",
				Detail:   err.Error(),
			}}
		}

		tempToken = configToken
	}

	// If no token was found, error out.
	if tempToken == "" {
		return "", diag.Diagnostics{{
			Severity: diag.Error,
			Summary:  "Authentication required",
			Detail:   "No CLI token found. Please log in using 'auth0 login' via auth0-cli or disable 'AUTH0_CLI_LOGIN'.",
		}}
	}

	// Check if the token is expired.
	if err := validateTokenExpiry(tempToken); err != nil {
		return "", diag.Diagnostics{{
			Severity: diag.Error,
			Summary:  "Token validation failed",
			Detail:   err.Error(),
		}}
	}

	return tempToken, nil
}

func decodeSegment(seg string) ([]byte, error) {
	// Add padding if necessary.
	if l := len(seg) % 4; l > 0 {
		seg += strings.Repeat("=", 4-l)
	}
	return base64.URLEncoding.DecodeString(seg)
}

func decodeJWT(token string) (map[string]interface{}, map[string]interface{}, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, nil, fmt.Errorf("invalid JWT format")
	}

	headerBytes, err := decodeSegment(parts[0])
	if err != nil {
		return nil, nil, fmt.Errorf("error decoding header: %w", err)
	}
	payloadBytes, err := decodeSegment(parts[1])
	if err != nil {
		return nil, nil, fmt.Errorf("error decoding payload: %w", err)
	}

	var header, payload map[string]interface{}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, nil, fmt.Errorf("error unmarshalling header: %w", err)
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, nil, fmt.Errorf("error unmarshalling payload: %w", err)
	}

	return header, payload, nil
}

func validateTokenExpiry(tokenString string) error {
	_, payload, err := decodeJWT(tokenString)
	if err != nil {
		return err
	}

	if exp, ok := payload["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return fmt.Errorf("expired token: the stored auth0-cli token has expired. Please log in again")
		}
	} else {
		return fmt.Errorf("missing expiration: the token does not contain an expiration claim")
	}

	return nil
}

// userAgent computes the desired User-Agent header for the *management.Management client.
func userAgent(terraformVersion *string) string {
	sdkVersion := auth0.Version
	terraformSDKVersion := meta.SDKVersionString() //nolint:staticcheck

	userAgent := fmt.Sprintf(
		"%s/%s (Go-Auth0-SDK/%s; Terraform-SDK/%s; Terraform/%s)",
		providerName,
		version,
		sdkVersion,
		terraformSDKVersion,
		*terraformVersion,
	)

	return userAgent
}

// authenticationOption computes the desired authentication option for the *management.Management client.
func authenticationOption(cfg ProviderConfig) management.Option {
	ctx := context.Background()

	switch {
	case cfg.ApiToken != "":
		return management.WithStaticToken(cfg.ApiToken)
	case cfg.Audience != "":
		if cfg.ClientAssertionPrivateKey != "" {
			return management.WithClientCredentialsPrivateKeyJwtAndAudience(
				ctx,
				cfg.ClientID,
				cfg.ClientAssertionPrivateKey,
				cfg.ClientAssertionSigningAlg,
				cfg.Audience,
			)
		}
		return management.WithClientCredentialsAndAudience(
			ctx,
			cfg.ClientID,
			cfg.ClientSecret,
			cfg.Audience,
		)
	case cfg.ClientAssertionPrivateKey != "":
		return management.WithClientCredentialsPrivateKeyJwt(
			ctx,
			cfg.ClientID,
			cfg.ClientAssertionPrivateKey,
			cfg.ClientAssertionSigningAlg,
		)

	default:
		return management.WithClientCredentials(ctx, cfg.ClientID, cfg.ClientSecret)
	}
}

func customClientWithRetries() *http.Client {
	client := &http.Client{
		Transport: rateLimitTransport(
			retryableErrorTransport(
				http.DefaultTransport,
			),
		),
	}

	return client
}

func rateLimitTransport(tripper http.RoundTripper) http.RoundTripper {
	return rehttp.NewTransport(tripper, rateLimitRetry, rateLimitDelay)
}

func rateLimitRetry(attempt rehttp.Attempt) bool {
	if attempt.Response == nil {
		return false
	}

	return attempt.Response.StatusCode == http.StatusTooManyRequests
}

func rateLimitDelay(attempt rehttp.Attempt) time.Duration {
	resetAt := attempt.Response.Header.Get("X-RateLimit-Reset")

	resetAtUnix, err := strconv.ParseInt(resetAt, 10, 64)
	if err != nil {
		resetAtUnix = time.Now().Add(5 * time.Second).Unix()
	}

	return time.Duration(resetAtUnix-time.Now().Unix()) * time.Second
}

func retryableErrorTransport(tripper http.RoundTripper) http.RoundTripper {
	return rehttp.NewTransport(
		tripper,
		rehttp.RetryAll(
			rehttp.RetryMaxRetries(3),
			rehttp.RetryAny(
				rehttp.RetryStatuses(
					http.StatusServiceUnavailable,
					http.StatusInternalServerError,
					http.StatusBadGateway,
					http.StatusGatewayTimeout,
					// Cloudflare-specific server error that is generated
					// because Cloudflare did not receive an HTTP response
					// from the origin server after an HTTP Connection was made.
					524,
				),
				rehttp.RetryIsErr(retryableErrorRetryFunc),
			),
		),
		rehttp.ExpJitterDelay(500*time.Millisecond, 10*time.Second),
	)
}

func retryableErrorRetryFunc(err error) bool {
	if err == nil {
		return false
	}

	if v, ok := err.(*url.Error); ok {
		// Don't retry if the error was due to too many redirects.
		if regexp.MustCompile(`stopped after \d+ redirects\z`).MatchString(v.Error()) {
			return false
		}

		// Don't retry if the error was due to an invalid protocol scheme.
		if regexp.MustCompile(`unsupported protocol scheme`).MatchString(v.Error()) {
			return false
		}

		// Don't retry if the certificate issuer is unknown.
		if _, ok := v.Err.(*tls.CertificateVerificationError); ok {
			return false
		}

		// Don't retry if the certificate issuer is unknown.
		if _, ok := v.Err.(x509.UnknownAuthorityError); ok {
			return false
		}
	}

	// The error is likely recoverable so retry.
	return true
}

// getAccessTokenFromCliConfigFile reads the Auth0 CLI config file and returns the access token for the given tenant.
func getAccessTokenFromCliConfigFile(domain string) (string, error) {
	if _, err := os.Stat(cliConfigPath); os.IsNotExist(err) {
		return "", nil
	}

	buffer, err := os.ReadFile(cliConfigPath)
	if err != nil {
		return "", fmt.Errorf("failed to read config.json: %w", err)
	}

	c := &CliConfig{}
	err = json.Unmarshal(buffer, c)
	if err != nil {
		return "", fmt.Errorf("failed to parse config.json: %w", err)
	}

	tenant, ok := c.Tenants[domain]
	if !ok {
		return "", nil
	}

	return tenant.AccessToken, nil
}
