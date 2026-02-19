package action

import (
	"context"
	"fmt"

	"github.com/auth0/go-auth0/v2/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
)

// NewModuleVersionDataSource will return a new auth0_action_module_version data source.
func NewModuleVersionDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readActionModuleVersionForDataSource,
		Description: "Data source to retrieve a specific version of an Auth0 action module.",
		Schema: map[string]*schema.Schema{
			"module_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the action module.",
			},
			"version_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the version to retrieve.",
			},
			"version_number": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The sequential version number.",
			},
			"code": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The exact source code that was published with this version.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The timestamp when this version was created.",
			},
			"dependencies": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Dependencies locked to this version.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Dependency name.",
						},
						"version": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Dependency version.",
						},
					},
				},
			},
			"secrets": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Secrets available to this version (name and updated_at only, values never returned).",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Secret name.",
						},
						"updated_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The time when this secret was last updated.",
						},
					},
				},
			},
		},
	}
}

func readActionModuleVersionForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()
	moduleID := data.Get("module_id").(string)
	versionID := data.Get("version_id").(string)

	version, err := apiv2.Actions.Modules.Versions.Get(ctx, moduleID, versionID)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(fmt.Sprintf("%s:%s", moduleID, versionID))

	return diag.FromErr(flattenActionModuleVersion(data, version))
}

func flattenActionModuleVersion(data *schema.ResourceData, version *management.GetActionModuleVersionResponseContent) error {
	result := multierror.Append(
		data.Set("version_number", version.GetVersionNumber()),
		data.Set("code", version.GetCode()),
		data.Set("dependencies", flattenActionModuleDependencies(version.GetDependencies())),
		data.Set("secrets", flattenActionModuleSecretsReadOnly(version.GetSecrets())),
	)

	if version.CreatedAt != nil {
		result = multierror.Append(result, data.Set("created_at", version.CreatedAt.String()))
	}

	return result.ErrorOrNil()
}
