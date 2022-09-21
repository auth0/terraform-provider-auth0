package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func newAction() *schema.Resource {
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
					"At this time, an action can only target a single trigger at a time.",
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
				}, false),
				Description: "The Node runtime, e.g. `node16`. Defaults to `node12`.",
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
	api := m.(*management.Management)

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
	api := m.(*management.Management)

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
		d.Set("dependencies", flattenActionDependencies(action.Dependencies)),
		d.Set("runtime", action.Runtime),
	)

	if action.DeployedVersion != nil {
		result = multierror.Append(result, d.Set("version_id", action.DeployedVersion.GetID()))
	}

	return diag.FromErr(result.ErrorOrNil())
}

func updateAction(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

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
	api := m.(*management.Management)

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

	api := m.(*management.Management)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		action, err := api.Action.Read(d.Id())
		if err != nil {
			return resource.NonRetryableError(err)
		}

		if action.GetStatus() == management.ActionStatusFailed {
			return resource.NonRetryableError(
				fmt.Errorf(
					"action %q failed to build, check the Auth0 UI for errors",
					action.GetName(),
				),
			)
		}

		if action.GetStatus() != management.ActionStatusBuilt {
			return resource.RetryableError(
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

func preventErasingUnmanagedSecrets(d *schema.ResourceData, api *management.Management) diag.Diagnostics {
	if !d.HasChange("secrets") {
		return nil
	}

	preUpdateAction, err := api.Action.Read(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// We need to also include the secrets that we're about to remove
	// against the checks, not just the ones with which we are left.
	oldSecrets, newSecrets := d.GetChange("secrets")
	allSecrets := append(oldSecrets.([]interface{}), newSecrets.([]interface{})...)

	return checkForUnmanagedActionSecrets(allSecrets, preUpdateAction.Secrets)
}

func checkForUnmanagedActionSecrets(
	secretsFromConfig []interface{},
	secretsFromAPI []*management.ActionSecret,
) diag.Diagnostics {
	secretKeysInConfigMap := make(map[string]bool, len(secretsFromConfig))
	for _, secret := range secretsFromConfig {
		secretKeyName := secret.(map[string]interface{})["name"].(string)
		secretKeysInConfigMap[secretKeyName] = true
	}

	var diagnostics diag.Diagnostics
	for _, secret := range secretsFromAPI {
		if _, ok := secretKeysInConfigMap[secret.GetName()]; !ok {
			diagnostics = append(diagnostics, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unmanaged Action Secret",
				Detail: fmt.Sprintf("Detected an action secret not managed though Terraform: %s. If you proceed, "+
					"this secret will get deleted. It is required to add this secret to your action configuration "+
					"to prevent unintentionally destructive results.",
					secret.GetName(),
				),
				AttributePath: cty.Path{cty.GetAttrStep{Name: "secrets"}},
			})
		}
	}

	return diagnostics
}

func expandAction(config cty.Value) *management.Action {
	action := &management.Action{
		Name:    value.String(config.GetAttr("name")),
		Code:    value.String(config.GetAttr("code")),
		Runtime: value.String(config.GetAttr("runtime")),
	}

	config.GetAttr("supported_triggers").ForEachElement(func(_ cty.Value, triggers cty.Value) (stop bool) {
		action.SupportedTriggers = []*management.ActionTrigger{
			{
				ID:      value.String(triggers.GetAttr("id")),
				Version: value.String(triggers.GetAttr("version")),
			},
		}

		return stop
	})

	config.GetAttr("dependencies").ForEachElement(func(_ cty.Value, deps cty.Value) (stop bool) {
		action.Dependencies = append(action.Dependencies, &management.ActionDependency{
			Name:    value.String(deps.GetAttr("name")),
			Version: value.String(deps.GetAttr("version")),
		})

		return true
	})

	config.GetAttr("secrets").ForEachElement(func(_ cty.Value, secrets cty.Value) (stop bool) {
		action.Secrets = append(action.Secrets, &management.ActionSecret{
			Name:  value.String(secrets.GetAttr("name")),
			Value: value.String(secrets.GetAttr("value")),
		})

		return true
	})

	return action
}

func flattenActionTriggers(triggers []*management.ActionTrigger) []interface{} {
	var result []interface{}
	for _, trigger := range triggers {
		result = append(result, map[string]interface{}{
			"id":      trigger.GetID(),
			"version": trigger.GetVersion(),
		})
	}
	return result
}

func flattenActionDependencies(dependencies []*management.ActionDependency) []interface{} {
	var result []interface{}
	for _, dependency := range dependencies {
		result = append(result, map[string]interface{}{
			"name":    dependency.GetName(),
			"version": dependency.GetVersion(),
		})
	}
	return result
}
