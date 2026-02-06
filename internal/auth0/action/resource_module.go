package action

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// NewModuleResource will return a new auth0_action_module resource.
func NewModuleResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createActionModule,
		ReadContext:   readActionModule,
		UpdateContext: updateActionModule,
		DeleteContext: deleteActionModule,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Action Modules are reusable code packages that can be shared across multiple actions. " +
			"They allow you to write common functionality once and use it in any action that needs it.",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the action module.",
			},
			"code": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The source code of the action module.",
			},
			"dependencies": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "List of third party npm modules, and their versions, that this action module depends on.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Dependency name, e.g. `lodash`.",
						},
						"version": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Dependency version, e.g. `latest` or `4.17.21`.",
						},
					},
				},
			},
			"secrets": {
				Type:     schema.TypeSet,
				Optional: true,
				Description: "List of secrets that are included in the action module. " +
					"Partial management of secrets is not supported.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Secret name. Required when configuring secrets",
						},
						"value": {
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
							Description: "Secret value. Required when configuring secrets",
						},
						"updated_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Last update time",
						},
					},
				},
			},
			"actions_using_module_total": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of deployed actions using this module.",
			},
			"all_changes_published": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether all draft changes have been published as a version.",
			},
			"latest_version_number": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The version number of the latest published version.",
			},
			"latest_version": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The latest published version of the action module.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The unique identifier of the version.",
						},
						"version_number": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The version number.",
						},
						"code": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The source code of this version.",
						},
						"created_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The time when this version was created.",
						},
						"dependencies": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of third party npm modules, and their versions, that this version depends on.",
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
							Description: "List of secrets that are included in this version.",
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
			"publish": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				Description: "Publishing a module will create a new immutable " +
					"version of the module from the current draft. Actions using " +
					"this module can then reference the published version.",
			},
			"version_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Version ID of the module. This value is available if `publish` is set to true.",
			},
		},
	}
}

func createActionModule(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()

	moduleRequest := expandActionModule(data)

	result, err := apiv2.Actions.Modules.Create(ctx, moduleRequest)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(result.GetID())

	if err := publishActionModule(ctx, data, meta); err != nil {
		return diag.FromErr(err)
	}

	return readActionModule(ctx, data, meta)
}

func readActionModule(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()

	module, err := apiv2.Actions.Modules.Get(ctx, data.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return diag.FromErr(flattenActionModule(data, module))
}

func updateActionModule(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()

	moduleRequest := expandActionModuleUpdate(data)

	_, err := apiv2.Actions.Modules.Update(ctx, data.Id(), moduleRequest)
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	if err := publishActionModule(ctx, data, meta); err != nil {
		return diag.FromErr(err)
	}

	return readActionModule(ctx, data, meta)
}

func deleteActionModule(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()

	if err := apiv2.Actions.Modules.Delete(ctx, data.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}

func publishActionModule(ctx context.Context, data *schema.ResourceData, meta interface{}) error {
	shouldPublish := data.Get("publish").(bool)
	if !shouldPublish {
		return nil
	}

	apiv2 := meta.(*config.Config).GetAPIV2()

	moduleVersion, err := apiv2.Actions.Modules.Versions.Create(ctx, data.Id())
	if err != nil {
		return err
	}

	return data.Set("version_id", moduleVersion.GetID())
}
