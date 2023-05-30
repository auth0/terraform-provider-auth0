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
