package validation

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// dynamicLoginURIPlaceholderRegex matches the Auth0-documented placeholder
// syntax allowed in initiate_login_uri for tenants using Organizations or
// Multiple Custom Domains. Metadata keys must start with `public_` or `PUBLIC_`
// See https://auth0.com/docs/get-started/applications/wildcards-for-subdomains#validation-rules.
var dynamicLoginURIPlaceholderRegex = regexp.MustCompile(`\{(organization|custom_domain)\.metadata\.(public_|PUBLIC_)[A-Za-z0-9_]+\}`)

// IsURLWithHTTPSorEmptyString is a validation func that checks
// that the given rawURL is a https url or is an empty string.
func IsURLWithHTTPSorEmptyString(rawURL interface{}, key string) ([]string, []error) {
	urlString, ok := rawURL.(string)
	if !ok {
		return nil, []error{
			fmt.Errorf("expected type of %q to be string", key),
		}
	}

	if urlString == "" {
		return nil, nil
	}

	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return nil, []error{
			fmt.Errorf("expected %q to be a valid url, got %v: %+v", key, urlString, err),
		}
	}

	if parsedURL.Host == "" {
		return nil, []error{
			fmt.Errorf("expected %q to have a host, got %v", key, urlString),
		}
	}

	if parsedURL.Scheme != "https" {
		return nil, []error{
			fmt.Errorf("expected %q to have a url with schema of: %q, got %v", key, "https", urlString),
		}
	}

	return nil, nil
}

// IsHTTPSURLOrEmptyStringWithDynamicLoginURIPlaceholders is a validation func
// that accepts an HTTPS URL, an empty string, or an HTTPS URL containing one
// or more Auth0 dynamic login URI metadata placeholders such as
// `{organization.metadata.public_login_host}` or
// `{custom_domain.metadata.public_app_host}`. The Auth0 Management API resolves
// these placeholders at request time for the `initiate_login_uri` field; this
// validator substitutes them with a stable host before invoking net/url so the
// surrounding URL grammar is still enforced.
func IsHTTPSURLOrEmptyStringWithDynamicLoginURIPlaceholders(rawURL interface{}, key string) ([]string, []error) {
	urlString, ok := rawURL.(string)
	if !ok {
		return nil, []error{
			fmt.Errorf("expected type of %q to be string", key),
		}
	}

	if urlString == "" {
		return nil, nil
	}

	substituted := dynamicLoginURIPlaceholderRegex.ReplaceAllString(urlString, "placeholder")

	parsedURL, err := url.Parse(substituted)
	if err != nil {
		return nil, []error{
			fmt.Errorf("expected %q to be a valid url, got %v: %+v", key, urlString, err),
		}
	}

	if parsedURL.Host == "" {
		return nil, []error{
			fmt.Errorf("expected %q to have a host, got %v", key, urlString),
		}
	}

	if parsedURL.Scheme != "https" {
		return nil, []error{
			fmt.Errorf("expected %q to have a url with schema of: %q, got %v", key, "https", urlString),
		}
	}

	return nil, nil
}

// UniversalLoginTemplateContainsCorrectTags is a validation func that checks
// that the given universal login template body contains the correct tags.
func UniversalLoginTemplateContainsCorrectTags(rawBody interface{}, key string) ([]string, []error) {
	v, ok := rawBody.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be string", key)}
	}

	if strings.Contains(v, "{%- auth0:head -%}") && strings.Contains(v, "{%- auth0:widget -%}") {
		return nil, nil
	}

	return nil, []error{
		fmt.Errorf("expected %q to contain a single auth0:head tag and at least one auth0:widget tag", key),
	}
}
