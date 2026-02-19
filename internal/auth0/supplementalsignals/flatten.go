package supplementalsignals

import (
	"github.com/auth0/go-auth0/v2/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// flattenSupplementalSignals maps SDK response to Terraform state.
func flattenSupplementalSignals(data *schema.ResourceData, config *management.GetSupplementalSignalsResponseContent) diag.Diagnostics {
	result := multierror.Append(
		data.Set("akamai_enabled", config.GetAkamaiEnabled()),
	)

	return diag.FromErr(result.ErrorOrNil())
}
