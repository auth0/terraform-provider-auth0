package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsURLWithHTTPSorEmptyString(t *testing.T) {
	var testCases = []struct {
		inputURL       interface{}
		expectedErrors []string
	}{
		{
			inputURL: "http://example.com",
			expectedErrors: []string{
				"expected \"theTestURL\" to have a url with schema of: \"https\", got http://example.com",
			},
		},
		{
			inputURL: "http://example.com/foo",
			expectedErrors: []string{
				"expected \"theTestURL\" to have a url with schema of: \"https\", got http://example.com/foo",
			},
		},
		{
			inputURL: "http://example.com#foo",
			expectedErrors: []string{
				"expected \"theTestURL\" to have a url with schema of: \"https\", got http://example.com#foo",
			},
		},
		{
			inputURL:       "https://example.com/foo",
			expectedErrors: nil,
		},
		{
			inputURL:       "https://example.com#foo",
			expectedErrors: nil,
		},
		{
			inputURL:       "",
			expectedErrors: nil,
		},
		{
			inputURL: "broken/url",
			expectedErrors: []string{
				"expected \"theTestURL\" to have a host, got broken/url",
			},
		},
		{
			inputURL: nil,
			expectedErrors: []string{
				"expected type of \"theTestURL\" to be string",
			},
		},
		{
			inputURL: "://example.com",
			expectedErrors: []string{
				"expected \"theTestURL\" to be a valid url, got ://example.com: parse \"://example.com\": missing protocol scheme",
			},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("test case #%d", i), func(t *testing.T) {
			var errorsAsString []string
			_, actualErrors := IsURLWithHTTPSorEmptyString(testCase.inputURL, "theTestURL")
			for _, actualError := range actualErrors {
				errorsAsString = append(errorsAsString, actualError.Error())
			}

			assert.Equal(t, testCase.expectedErrors, errorsAsString)
		})
	}
}

func TestIsHTTPSURLOrEmptyStringWithDynamicLoginURIPlaceholders(t *testing.T) {
	var testCases = []struct {
		name           string
		inputURL       interface{}
		expectedErrors []string
	}{
		{
			name:           "plain https url",
			inputURL:       "https://example.com/foo",
			expectedErrors: nil,
		},
		{
			name:           "empty string",
			inputURL:       "",
			expectedErrors: nil,
		},
		{
			name:           "organization metadata placeholder as host",
			inputURL:       "https://{organization.metadata.public_login_host}",
			expectedErrors: nil,
		},
		{
			name:           "organization metadata placeholder with path",
			inputURL:       "https://{organization.metadata.public_login_host}/login",
			expectedErrors: nil,
		},
		{
			name:           "custom_domain metadata placeholder",
			inputURL:       "https://{custom_domain.metadata.public_app_host}/login",
			expectedErrors: nil,
		},
		{
			name:           "both placeholder types combined",
			inputURL:       "https://{custom_domain.metadata.public_app_host}/o/{organization.metadata.public_login_host}",
			expectedErrors: nil,
		},
		{
			inputURL: "http://{organization.metadata.public_login_host}",
			name:     "http scheme is rejected even with placeholder",
			expectedErrors: []string{
				"expected \"theTestURL\" to have a url with schema of: \"https\", got http://{organization.metadata.public_login_host}",
			},
		},
		{
			inputURL: "https://{organization.metadata.no_public_prefix}",
			name:     "metadata key without public_ prefix is rejected",
			expectedErrors: []string{
				"expected \"theTestURL\" to be a valid url, got https://{organization.metadata.no_public_prefix}: parse \"https://{organization.metadata.no_public_prefix}\": invalid character \"{\" in host name",
			},
		},
		{
			inputURL: "https://{unsupported.metadata.public_x}",
			name:     "unsupported placeholder source is rejected",
			expectedErrors: []string{
				"expected \"theTestURL\" to be a valid url, got https://{unsupported.metadata.public_x}: parse \"https://{unsupported.metadata.public_x}\": invalid character \"{\" in host name",
			},
		},
		{
			inputURL: nil,
			name:     "non-string input is rejected",
			expectedErrors: []string{
				"expected type of \"theTestURL\" to be string",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var errorsAsString []string
			_, actualErrors := IsHTTPSURLOrEmptyStringWithDynamicLoginURIPlaceholders(testCase.inputURL, "theTestURL")
			for _, actualError := range actualErrors {
				errorsAsString = append(errorsAsString, actualError.Error())
			}

			assert.Equal(t, testCase.expectedErrors, errorsAsString)
		})
	}
}

func TestUniversalLoginTemplateContainsCorrectTags(t *testing.T) {
	tests := []struct {
		name          string
		input         interface{}
		key           string
		expectedError string
	}{
		{
			name:          "valid input",
			input:         `Some content {%- auth0:head -%} More content {%- auth0:widget -%}`,
			key:           "testKey",
			expectedError: "",
		},
		{
			name:          "missing auth0:head tag",
			input:         `Some content More content {%- auth0:widget -%}`,
			key:           "testKey",
			expectedError: "expected \"testKey\" to contain a single auth0:head tag and at least one auth0:widget tag",
		},
		{
			name:          "missing auth0:widget tag",
			input:         `Some content {%- auth0:head -%} More content`,
			key:           "testKey",
			expectedError: "expected \"testKey\" to contain a single auth0:head tag and at least one auth0:widget tag",
		},
		{
			name:          "incorrect input type",
			input:         42,
			key:           "testKey",
			expectedError: "expected type of \"testKey\" to be string",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, errors := UniversalLoginTemplateContainsCorrectTags(test.input, test.key)

			if test.expectedError != "" {
				assert.EqualError(t, errors[0], test.expectedError)
				return
			}

			assert.Len(t, errors, 0)
		})
	}
}
