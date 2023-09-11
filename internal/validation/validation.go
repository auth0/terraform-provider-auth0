package validation

import (
	"fmt"
	"net/url"
	"strings"
)

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
