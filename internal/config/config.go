package config

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/PuerkitoBio/rehttp"
	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/meta"

	"github.com/auth0/terraform-provider-auth0/internal/mutex"
)

const providerName = "Terraform-Provider-Auth0" // #nosec G101

var version = "dev"

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

// ConfigureProvider will configure the *schema.Provider so that
// *management.Management client and *mutex.KeyValue is stored
// and passed into the subsequent resources as the meta parameter.
func ConfigureProvider(terraformVersion *string) schema.ConfigureContextFunc {
	return func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
		domain := data.Get("domain").(string)
		clientID := data.Get("client_id").(string)
		clientSecret := data.Get("client_secret").(string)
		apiToken := data.Get("api_token").(string)
		audience := data.Get("audience").(string)
		debug := data.Get("debug").(bool)

		apiClient, err := management.New(domain,
			authenticationOption(clientID, clientSecret, apiToken, audience),
			management.WithDebug(debug),
			management.WithUserAgent(userAgent(terraformVersion)),
			management.WithAuth0ClientEnvEntry(providerName, version),
			management.WithNoRetries(),
			management.WithClient(customClientWithRetries()),
		)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		return New(apiClient), nil
	}
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
func authenticationOption(clientID, clientSecret, apiToken, audience string) management.Option {
	ctx := context.Background()

	if apiToken != "" {
		return management.WithStaticToken(apiToken)
	}

	if audience != "" {
		return management.WithClientCredentialsAndAudience(
			ctx,
			clientID,
			clientSecret,
			audience,
		)
	}

	return management.WithClientCredentials(ctx, clientID, clientSecret)
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
