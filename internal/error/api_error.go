package error

import (
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// HandleAPIError handles the API error.
// If the error code is a 404 it triggers a resource deletion.
func HandleAPIError(data *schema.ResourceData, err error) error {
	if IsStatusNotFound(err) {
		data.SetId("")
		return nil
	}

	return err
}

// IsStatusNotFound checks to see if the error from the Auth0 Management API is a 404.
func IsStatusNotFound(err error) bool {
	if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
		return true
	}

	return false
}
