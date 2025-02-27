package connection

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// NewSCIMConfigurationResource will return a new auth0_connection_scim_configuration (1:1) resource.
func NewSCIMConfigurationResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createSCIMConfiguration,
		UpdateContext: updateSCIMConfiguration,
		ReadContext:   readSCIMConfiguration,
		DeleteContext: deleteSCIMConfiguration,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can configure [SCIM(System for Cross-domain Identity Management)](https://simplecloud.info/) support " +
			"for `SAML` and `OpenID Connect` Enterprise connections.",
		Schema: map[string]*schema.Schema{
			"connection_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the connection for this SCIM configuration.",
			},
			"connection_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the connection for this SCIM configuration.",
			},
			"strategy": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Schema of the connection for this SCIM configuration.",
			},
			"tenant_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the tenant for this SCIM configuration.",
			},
			"user_id_attribute": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"user_id_attribute", "mapping"},
				Computed:     true,
				Description: "User ID attribute for generation unique of user ids. If `user_id_attribute` is set, `mapping` " +
					"must be set as well. Defaults to `userName` for SAML connections and `externalId` for OIDC connections.",
			},
			"mapping": {
				Type:         schema.TypeSet,
				Optional:     true,
				RequiredWith: []string{"user_id_attribute", "mapping"},
				Computed:     true,
				Description: "Mapping between Auth0 attributes and SCIM attributes. If `user_id_attribute` is set, `mapping` " +
					"must be set as well.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auth0": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
							Description:  "The field location in the Auth0 schema.",
						},
						"scim": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
							Description:  "The field location in the SCIM schema.",
						},
					},
				},
			},
		},
	}
}

func createSCIMConfiguration(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	connectionID := data.Get("connection_id").(string)
	scimConfiguration := expandSCIMConfiguration(data)

	if err := api.Connection.CreateSCIMConfiguration(ctx, connectionID, scimConfiguration); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(connectionID)

	return readSCIMConfiguration(ctx, data, meta)
}

func updateSCIMConfiguration(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	scimConfiguration := expandSCIMConfiguration(data)

	if err := api.Connection.UpdateSCIMConfiguration(ctx, data.Id(), scimConfiguration); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readSCIMConfiguration(ctx, data, meta)
}

func readSCIMConfiguration(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	scimConfiguration, err := api.Connection.ReadSCIMConfiguration(ctx, data.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return flattenSCIMConfiguration(data, scimConfiguration)
}

func deleteSCIMConfiguration(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	if err := api.Connection.DeleteSCIMConfiguration(ctx, data.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
