package clientprofile

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewDataSource will return a new auth0_client_profile data source.
func NewDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readClientProfileForDataSource,
		Description: "Data source to retrieve a specific Auth0 client profile (application profile) by `id`. " +
			"This feature is available in Early Access only and requires the `read:client_profiles` " +
			"scope on the Management API client. (EA only)",
		Schema: dataSourceSchema(),
	}
}

func dataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(NewResource().Schema)
	dataSourceSchema["id"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The ID of the client profile.",
	}

	return dataSourceSchema
}

func readClientProfileForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	id := data.Get("id").(string)

	profile, err := apiv3.ClientProfiles.Get(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(profile.GetID())

	return flattenClientProfile(data, profile)
}
