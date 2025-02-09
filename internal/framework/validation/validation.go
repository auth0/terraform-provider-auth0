package validation

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// isURLWithHTTPSorEmptyStringValidator validates that a string Attribute is an https URL or is an empty string.
type isURLWithHTTPSorEmptyStringValidator struct {
}

// Description describes the validation in plain text formatting.
func (v isURLWithHTTPSorEmptyStringValidator) Description(_ context.Context) string {
	return fmt.Sprintf("string must be a valid https URL or empty")
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v isURLWithHTTPSorEmptyStringValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v isURLWithHTTPSorEmptyStringValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	urlString := request.ConfigValue.ValueString()

	if urlString == "" {
		return
	}

	parsedURL, err := url.Parse(urlString)
	if err != nil {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			fmt.Sprintf("invalid URL: %+v", err),
			urlString,
		))
	} else if parsedURL.Host == "" {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			fmt.Sprintf("invalid URL, missing host"),
			urlString,
		))
	} else if parsedURL.Scheme != "https" {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			fmt.Sprintf("invalid URL, expected protocol scheme %q", "https"),
			urlString,
		))
	}
}

// IsURLWithHTTPSorEmptyString returns a validator which ensures
// that the given rawURL is a https url or is an empty string.
// Null (unconfigured) and unknown (known after apply)
// values are skipped.
func IsURLWithHTTPSorEmptyString() validator.String {
	return isURLWithHTTPSorEmptyStringValidator{}
}
