package auth0

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// CheckFor404Error accepts and executes a resource's read function within the context
// of a data source and will appropriately error when a resource is not found.
func CheckFor404Error(ctx context.Context, readFunc func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	diags := readFunc(ctx, d, m)
	if diags != nil {
		return diags
	}
	if d.Id() == "" {
		return diag.FromErr(errors.New("data source with that identifier not found (404)"))
	}
	return nil
}
