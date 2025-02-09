package validation

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestFrameworkIsURLWithHTTPSorEmptyString(t *testing.T) {
	var testCases = []struct {
		inputURL       interface{}
		expectedErrors []string
	}{
		{
			inputURL: "http://example.com",
			expectedErrors: []string{
				"Attribute theTestURLPath invalid URL, expected protocol scheme \"https\", got: http://example.com",
			},
		},
		{
			inputURL: "http://example.com/foo",
			expectedErrors: []string{
				"Attribute theTestURLPath invalid URL, expected protocol scheme \"https\", got: http://example.com/foo",
			},
		},
		{
			inputURL: "http://example.com#foo",
			expectedErrors: []string{
				"Attribute theTestURLPath invalid URL, expected protocol scheme \"https\", got: http://example.com#foo",
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
				"Attribute theTestURLPath invalid URL, missing host, got: broken/url",
			},
		},
		{
			inputURL: nil,
			expectedErrors: nil,
		},
		{
			inputURL: "://example.com",
			expectedErrors: []string{
				"Attribute theTestURLPath invalid URL: parse \"://example.com\": missing protocol scheme, got: ://example.com",
			},
		},
	}
	urlValidator := IsURLWithHTTPSorEmptyString()

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("test case #%d", i), func(t *testing.T) {
			response := validator.StringResponse{}
			request := validator.StringRequest{
				Path: path.Root("theTestURLPath"),
			}
			if testCase.inputURL == nil {
				request.ConfigValue = types.StringNull()
			} else {
				request.ConfigValue = types.StringValue(testCase.inputURL.(string))
			}
			urlValidator.ValidateString(context.Background(), request, &response)
			assert.Equal(t, len(testCase.expectedErrors), len(response.Diagnostics))
			for i, diagnostic := range response.Diagnostics {
				assert.Regexp(t, regexp.MustCompile(testCase.expectedErrors[i]), diagnostic.Detail())
			}
		})
	}
}
