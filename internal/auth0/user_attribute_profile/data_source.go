package userattributeprofile

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewDataSource will return a new auth0_user_attribute_profile data source.
func NewDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readUserAttributeProfileForDataSource,
		Description: "Data source to retrieve a specific Auth0 User Attribute Profile by `user_attribute_profile_id` or `name`.",
		Schema:      dataSourceSchema(),
	}
}

func dataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(NewResource().Schema)

	dataSourceSchema["user_attribute_profile_id"] = &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Description:  "The ID of the User Attribute Profile.",
		AtLeastOneOf: []string{"user_attribute_profile_id", "name"},
	}

	internalSchema.SetExistingAttributesAsOptional(dataSourceSchema, "name")

	dataSourceSchema["name"].AtLeastOneOf = []string{"user_attribute_profile_id", "name"}

	return dataSourceSchema
}

func readUserAttributeProfileForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	userAttributeProfileID := data.Get("user_attribute_profile_id").(string)
	if userAttributeProfileID != "" {
		data.SetId(userAttributeProfileID)
		return readUserAttributeProfile(ctx, data, meta)
	}

	name := data.Get("name").(string)

	userAttributeProfiles, err := api.UserAttributeProfile.List(ctx)
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	for _, userAttributeProfile := range userAttributeProfiles.UserAttributeProfiles {
		if userAttributeProfile.GetName() == name {
			data.SetId(userAttributeProfile.GetID())
			return diag.FromErr(flattenUserAttributeProfile(data, userAttributeProfile))
		}
	}

	return diag.Errorf("No User Attribute Profile found with name %q", name)
}
