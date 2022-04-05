package auth0

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func newDataGlobalClient() *schema.Resource {
	return &schema.Resource{
		ReadContext: readDataGlobalClient,
		Schema:      newClientSchema(),
	}
}

func readDataGlobalClient(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if err := readGlobalClientID(ctx, d, m); err != nil {
		return err
	}
	return readClient(ctx, d, m)
}
