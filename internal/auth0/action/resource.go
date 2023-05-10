package action

import (
	"context"
	"fmt"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
)

// NewResource will return a new auth0_action resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createAction,
		ReadContext:   readAction,
		UpdateContext: updateAction,
		DeleteContext: deleteAction,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Actions are secure, tenant-specific, versioned functions written in Node.js " +
			"that execute at certain points during the Auth0 runtime. Actions are used to customize " +
			"and extend Auth0's capabilities with custom logic.",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the action.",
			},
			"supported_triggers": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The trigger ID.",
						},
						"version": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The trigger version. This regulates which `runtime` versions are supported.",
						},
					},
				},
				Description: "List of triggers that this action supports. " +
					"At this time, an action can only target a single trigger at a time. " +
					"Read [Retrieving the set of triggers available within actions](https://registry.terraform.io/providers/auth0/auth0/latest/docs/guides/action_triggers) " +
					"to retrieve the latest trigger versions supported.",
			},
			"code": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The source code of the action.",
			},
			"dependencies": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "List of third party npm modules, and their versions, that this action depends on.",
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
			"runtime": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"node12",
					"node16",
					"node18",
				}, false),
				Description: "The Node runtime. Defaults to `node12`. Possible values are: " +
					"`node12`, `node16` or `node18`.",
			},
			"secrets": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of secrets that are included in an action or a version of an action.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Secret name.",
						},
						"value": {
							Type:        schema.TypeString,
							Required:    true,
							Sensitive:   true,
							Description: "Secret value.",
						},
					},
				},
			},
			"deploy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				Description: "Deploying an action will create a new immutable" +
					" version of the action. If the action is currently bound" +
					" to a trigger, then the system will begin executing the " +
					"newly deployed version of the action immediately.",
			},
			"version_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Version ID of the action. This value is available if `deploy` is set to true.",
			},
		},
	}
}

func createAction(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	action := expandAction(d.GetRawConfig())
	if err := api.Action.Create(action); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(action.GetID())

	if result := deployAction(ctx, d, m); result.HasError() {
		return result
	}

	return readAction(ctx, d, m)
}

func readAction(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	action, err := api.Action.Read(d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return diag.FromErr(err)
	}

	result := multierror.Append(
		d.Set("name", action.Name),
		d.Set("supported_triggers", flattenActionTriggers(action.SupportedTriggers)),
		d.Set("code", action.Code),
		d.Set("dependencies", flattenActionDependencies(action.GetDependencies())),
		d.Set("runtime", action.Runtime),
	)

	if action.DeployedVersion != nil {
		result = multierror.Append(result, d.Set("version_id", action.DeployedVersion.GetID()))
	}

	return diag.FromErr(result.ErrorOrNil())
}

func updateAction(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	diagnostics := preventErasingUnmanagedSecrets(d, api)
	if diagnostics.HasError() {
		return diagnostics
	}

	action := expandAction(d.GetRawConfig())
	if err := api.Action.Update(d.Id(), action); err != nil {
		return diag.FromErr(err)
	}

	if result := deployAction(ctx, d, m); result.HasError() {
		return result
	}

	return readAction(ctx, d, m)
}

func deleteAction(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	if err := api.Action.Delete(d.Id()); err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func deployAction(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	deployExists := d.Get("deploy").(bool)
	if !deployExists {
		return nil
	}

	api := m.(*config.Config).GetAPI()

	err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		action, err := api.Action.Read(d.Id())
		if err != nil {
			return retry.NonRetryableError(err)
		}

		if action.GetStatus() == management.ActionStatusFailed {
			return retry.NonRetryableError(
				fmt.Errorf(
					"action %q failed to build, check the Auth0 UI for errors",
					action.GetName(),
				),
			)
		}

		if action.GetStatus() != management.ActionStatusBuilt {
			return retry.RetryableError(
				fmt.Errorf(
					"expected action %q status %q to equal %q",
					action.GetName(),
					action.GetStatus(),
					"built",
				),
			)
		}

		return nil
	})
	if err != nil {
		return diag.FromErr(
			fmt.Errorf(
				"action %q never reached built state: %w",
				d.Get("name").(string),
				err,
			),
		)
	}

	actionVersion, err := api.Action.Deploy(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(d.Set("version_id", actionVersion.GetID()))
}
