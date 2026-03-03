package action

import (
	"context"

	"github.com/auth0/go-auth0/v2/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
)

// NewModuleVersionsDataSource will return a new auth0_action_module_versions data source.
func NewModuleVersionsDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readActionModuleVersionsForDataSource,
		Description: "Data source to retrieve all published versions of a specific Auth0 action module.",
		Schema: map[string]*schema.Schema{
			"module_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the action module.",
			},
			"versions": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of all published versions of the action module.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The unique identifier of the version.",
						},
						"module_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the parent module.",
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
				},
			},
		},
	}
}

func readActionModuleVersionsForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()
	moduleID := data.Get("module_id").(string)

	versionsPage, err := apiv2.Actions.Modules.Versions.List(ctx, moduleID, &management.GetActionModuleVersionsRequestParameters{})
	if err != nil {
		return diag.FromErr(err)
	}

	// Collect all versions using the iterator.
	var allVersions []*management.ActionModuleVersion
	iterator := versionsPage.Iterator()
	for iterator.Next(ctx) {
		allVersions = append(allVersions, iterator.Current())
	}
	if err := iterator.Err(); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(moduleID)

	return diag.FromErr(flattenActionModuleVersions(data, allVersions))
}

func flattenActionModuleVersions(data *schema.ResourceData, versions []*management.ActionModuleVersion) error {
	result := multierror.Append(
		data.Set("versions", flattenActionModuleVersionsList(versions)),
	)

	return result.ErrorOrNil()
}

func flattenActionModuleVersionsList(versions []*management.ActionModuleVersion) []interface{} {
	var result []interface{}

	for _, version := range versions {
		versionMap := map[string]interface{}{
			"id":             version.GetID(),
			"module_id":      version.GetModuleID(),
			"version_number": version.GetVersionNumber(),
			"code":           version.GetCode(),
			"dependencies":   flattenActionModuleDependencies(version.GetDependencies()),
			"secrets":        flattenActionModuleSecretsReadOnly(version.GetSecrets()),
		}

		if version.CreatedAt != nil {
			versionMap["created_at"] = version.CreatedAt.String()
		}

		result = append(result, versionMap)
	}

	return result
}

func flattenActionModuleSecretsReadOnly(secrets []*management.ActionModuleSecret) []interface{} {
	var result []interface{}

	for _, secret := range secrets {
		secretMap := map[string]interface{}{
			"name": secret.GetName(),
		}

		if secret.UpdatedAt != nil {
			secretMap["updated_at"] = secret.UpdatedAt.String()
		}

		result = append(result, secretMap)
	}

	return result
}
