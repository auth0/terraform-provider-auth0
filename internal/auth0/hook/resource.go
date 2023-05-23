package hook

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	"github.com/auth0/terraform-provider-auth0/internal/value"
)

// NewResource will return a new auth0_hook resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: "This resource is deprecated. Refer to the [guide on how to migrate from hooks to actions](https://auth0.com/docs/customize/actions/migrate/migrate-from-hooks-to-actions) " +
			"and manage your actions using the `auth0_action` resource.",
		CreateContext: createHook,
		ReadContext:   readHook,
		UpdateContext: updateHook,
		DeleteContext: deleteHook,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Hooks are secure, self-contained functions that allow you to customize the behavior of " +
			"Auth0 when executed for selected extensibility points of the Auth0 platform. Auth0 invokes Hooks " +
			"during runtime to execute your custom Node.js code. Depending on the extensibility point, " +
			"you can use Hooks with Database Connections and/or Passwordless Connections.",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateHookName(),
				Description:      "Name of this hook.",
			},
			"dependencies": {
				Type:        schema.TypeMap,
				Elem:        schema.TypeString,
				Optional:    true,
				Description: "Dependencies of this hook used by the WebTask server.",
			},
			"script": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Code to be executed when this hook runs.",
			},
			"trigger_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"credentials-exchange",
					"pre-user-registration",
					"post-user-registration",
					"post-change-password",
					"send-phone-message",
				}, false),
				Description: "Execution stage of this rule. Can be " +
					"credentials-exchange, pre-user-registration, " +
					"post-user-registration, post-change-password" +
					", or send-phone-message.",
			},
			"secrets": {
				Type:        schema.TypeMap,
				Elem:        schema.TypeString,
				Sensitive:   true,
				Optional:    true,
				Description: "The secrets associated with the hook.",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether the hook is enabled, or disabled.",
			},
		},
	}
}

func createHook(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	hook := expandHook(d)
	if err := api.Hook.Create(hook); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(hook.GetID())

	if err := upsertHookSecrets(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return readHook(ctx, d, m)
}

func readHook(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()
	hook, err := api.Hook.Read(d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return diag.FromErr(err)
	}

	diagnostics := checkForUntrackedHookSecrets(ctx, d, m)

	result := multierror.Append(
		d.Set("name", hook.Name),
		d.Set("dependencies", hook.Dependencies),
		d.Set("script", hook.Script),
		d.Set("trigger_id", hook.TriggerID),
		d.Set("enabled", hook.Enabled),
	)

	if err = result.ErrorOrNil(); err != nil {
		diagnostics = append(diagnostics, diag.FromErr(err)...)
	}

	return diagnostics
}

func updateHook(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	hook := expandHook(d)
	api := m.(*config.Config).GetAPI()
	if err := api.Hook.Update(d.Id(), hook); err != nil {
		return diag.FromErr(err)
	}

	if err := upsertHookSecrets(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return readHook(ctx, d, m)
}

func deleteHook(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()
	if err := api.Hook.Delete(d.Id()); err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return diag.FromErr(err)
	}

	return nil
}

func upsertHookSecrets(_ context.Context, d *schema.ResourceData, m interface{}) error {
	if d.IsNewResource() || d.HasChange("secrets") {
		api := m.(*config.Config).GetAPI()

		hookSecrets := value.MapOfStrings(d.GetRawConfig().GetAttr("secrets"))
		if hookSecrets == nil {
			return nil
		}

		return api.Hook.ReplaceSecrets(d.Id(), *hookSecrets)
	}

	return nil
}
