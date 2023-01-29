package validation

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func IsURLWithHTTPSorEmptyString(i interface{}, s string) ([]string, []error) {
	_, errors := validation.IsURLWithHTTPS(i, s)
	for _, err := range errors {
		if !strings.Contains(err.Error(), "url to not be empty") {
			return nil, errors
		}
	}
	return nil, nil
}
