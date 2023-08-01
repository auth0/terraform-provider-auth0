package hook

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
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
			"you can use hooks with Database Connections and/or Passwordless Connections." +
			"\n\n!> This resource is deprecated. Refer to the [guide on how to migrate from hooks to actions](https://auth0.com/docs/customize/actions/migrate/migrate-from-hooks-to-actions) " +
			"and manage your actions using the `auth0_action` resource.",
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

func createHook(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	hook := expandHook(data)

	if err := api.Hook.Create(ctx, hook); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(hook.GetID())

	if err := upsertHookSecrets(ctx, data, meta); err != nil {
		return diag.FromErr(err)
	}

	return readHook(ctx, data, meta)
}

func readHook(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	hook, err := api.Hook.Read(ctx, data.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	hookSecrets, err := api.Hook.Secrets(ctx, data.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	configSecrets := data.Get("secrets").(map[string]interface{})

	diagnostics := checkForUntrackedHookSecrets(hookSecrets, configSecrets)

	if err := flattenHook(data, hook); err != nil {
		diagnostics = append(diagnostics, diag.FromErr(err)...)
	}

	return diagnostics
}

func updateHook(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	hook := expandHook(data)

	if err := api.Hook.Update(ctx, data.Id(), hook); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	if err := upsertHookSecrets(ctx, data, meta); err != nil {
		return diag.FromErr(err)
	}

	return readHook(ctx, data, meta)
}

func deleteHook(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	if err := api.Hook.Delete(ctx, data.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}

func upsertHookSecrets(ctx context.Context, data *schema.ResourceData, meta interface{}) error {
	if data.IsNewResource() || data.HasChange("secrets") {
		api := meta.(*config.Config).GetAPI()

		hookSecrets := value.MapOfStrings(data.GetRawConfig().GetAttr("secrets"))
		if hookSecrets == nil {
			return nil
		}

		return api.Hook.ReplaceSecrets(ctx, data.Id(), *hookSecrets)
	}

	return nil
}
