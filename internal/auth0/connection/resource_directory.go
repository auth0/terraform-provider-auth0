package connection

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// NewDirectoryResource will return a new auth0_connection_directory (1:1) resource.
func NewDirectoryResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createDirectory,
		UpdateContext: updateDirectory,
		ReadContext:   readDirectory,
		DeleteContext: deleteDirectory,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can configure directory provisioning (directory sync) for " +
			"`Google Workspace` Enterprise connections. This enables automatic user provisioning from the identity provider to Auth0.",
		Schema: map[string]*schema.Schema{
			"connection_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the connection for this directory provisioning configuration.",
			},
			"connection_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the connection for this directory provisioning configuration.",
			},
			"strategy": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Strategy of the connection for this directory provisioning configuration.",
			},
			"mapping": {
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Description: "Mapping between Auth0 attributes and IDP user attributes. Defaults to default mapping for the connection type if not specified.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auth0": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
							Description:  "The field location in the Auth0 schema.",
						},
						"idp": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
							Description:  "The field location in the IDP schema.",
						},
					},
				},
			},
			"synchronize_automatically": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether periodic automatic synchronization is enabled. Defaults to false.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The timestamp at which the directory provisioning configuration was created.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The timestamp at which the directory provisioning configuration was last updated.",
			},
			"last_synchronization_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The timestamp at which the connection was last synchronized.",
			},
			"last_synchronization_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the last synchronization.",
			},
			"last_synchronization_error": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The error message of the last synchronization, if any.",
			},
		},
	}
}

func createDirectory(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()
	connectionID := data.Get("connection_id").(string)

	directoryConfig := expandDirectory(data)

	result, err := apiv2.Connections.DirectoryProvisioning.Create(ctx, connectionID, directoryConfig)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(result.ConnectionID)

	return readDirectory(ctx, data, meta)
}

func updateDirectory(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()
	directoryConfig := expandDirectoryUpdate(data)

	_, err := apiv2.Connections.DirectoryProvisioning.Update(ctx, data.Id(), directoryConfig)
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readDirectory(ctx, data, meta)
}

func readDirectory(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()

	directoryConfig, err := apiv2.Connections.DirectoryProvisioning.Get(ctx, data.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return flattenDirectory(data, directoryConfig)
}

func deleteDirectory(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()

	if err := apiv2.Connections.DirectoryProvisioning.Delete(ctx, data.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
