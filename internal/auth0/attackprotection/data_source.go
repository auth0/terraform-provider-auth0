package attackprotection

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewDataSource will return a new auth0_attack_protection data source.
func NewDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readAttackProtectionForDataSource,
		Description: "Use this data source to access information about the tenant's attack protection settings.",
		Schema:      dataSourceSchema(),
	}
}

func dataSourceSchema() map[string]*schema.Schema {
	return internalSchema.TransformResourceToDataSource(NewResource().Schema)
}

func readAttackProtectionForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	data.SetId(resource.UniqueId())
	return readAttackProtection(ctx, data, meta)
}
