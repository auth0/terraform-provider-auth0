package validation

import (
	"fmt"
	"net/url"
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
