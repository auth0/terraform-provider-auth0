// Package framework holds helpers shared by every resource / data source
// implementation in this provider.
package framework

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	mgmt "github.com/auth0/go-auth0/v2/management"
	mgmtclient "github.com/auth0/go-auth0/v2/management/client"
)

// ManagementFromResource extracts the *management.Management value injected by
// the provider's Configure step. Returns false (with a typed diagnostic) when
// the data is missing or of the wrong type, in which case the caller should
// abort.
func ManagementFromResource(req resource.ConfigureRequest, resp *resource.ConfigureResponse) (*mgmtclient.Management, bool) {
	if req.ProviderData == nil {
		return nil, false
	}
	m, ok := req.ProviderData.(*mgmtclient.Management)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *management.Management, got %T. This is a bug in the provider.", req.ProviderData),
		)
		return nil, false
	}
	return m, true
}

// ManagementFromDataSource is the data-source counterpart to
// ManagementFromResource.
func ManagementFromDataSource(req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) (*mgmtclient.Management, bool) {
	if req.ProviderData == nil {
		return nil, false
	}
	m, ok := req.ProviderData.(*mgmtclient.Management)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *management.Management, got %T. This is a bug in the provider.", req.ProviderData),
		)
		return nil, false
	}
	return m, true
}

// IsNotFound reports whether err represents an Auth0 "404 Not Found" response.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	var nf *mgmt.NotFoundError
	return errors.As(err, &nf)
}

// AddAPIError attaches an Auth0 API error to the diagnostics in a uniform way.
func AddAPIError(diags *diag.Diagnostics, summary string, err error) {
	diags.AddError(summary, err.Error())
}
