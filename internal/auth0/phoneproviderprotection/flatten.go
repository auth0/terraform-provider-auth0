package phoneproviderprotection

import (
	"github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenPhoneProviderProtection(d *schema.ResourceData, cfg *management.GetPhoneProviderProtectionResponseContent) diag.Diagnostics {
	result := multierror.Append(
		d.Set("type", string(cfg.GetType())),
	)

	return diag.FromErr(result.ErrorOrNil())
}
