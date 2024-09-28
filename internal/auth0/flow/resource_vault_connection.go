package flow

import (
	"context"
	"encoding/json"
	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
)

// NewVaultConnectionResource will return a new auth0_flow_vault_connection resource.
func NewVaultConnectionResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createVaultConnection,
		ReadContext:   readVaultConnection,
		UpdateContext: updateVaultConnection,
		DeleteContext: deleteVaultConnection,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can create and manage flow vault connections for a tenant.",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the vault connection.",
			},
			"app_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "App identifier of the vault connection.",
				ValidateFunc: validation.StringInSlice([]string{
					"ACTIVECAMPAIGN",
					"AIRTABLE",
					"AUTH0",
					"BIGQUERY",
					"CLEARBIT",
					"DOCUSIGN",
					"GOOGLE_SHEETS",
					"HTTP",
					"HUBSPOT",
					"JWT",
					"MAILCHIMP",
					"MAILJET",
					"PIPEDRIVE",
					"SALESFORCE",
					"SENDGRID",
					"SLACK",
					"STRIPE",
					"TELEGRAM",
					"TWILIO",
					"WHATSAPP",
					"ZAPIER",
				}, false),
			},
			"environment": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Environment of the vault connection.",
			},
			"setup": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Configuration of the vault connection.",
				Sensitive:   true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"account_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Custom account name of the vault connection.",
			},
			"ready": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates if the vault connection is configured.",
			},
			"fingerprint": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Fingerprint of the vault connection.",
			},
		},
	}
}

func createVaultConnection(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	vaultConnection, err := expandVaultConnection(data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := api.Flow.Vault.CreateConnection(ctx, vaultConnection); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(vaultConnection.GetID())

	return readVaultConnection(ctx, data, meta)
}

func readVaultConnection(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	vaultConnection, err := api.Flow.Vault.GetConnection(ctx, data.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return diag.FromErr(flattenVaultConnection(data, vaultConnection))
}

func updateVaultConnection(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	vaultConnection, err := expandVaultConnection(data)
	if err != nil {
		return diag.FromErr(err)
	}

	d, _ := json.Marshal(vaultConnection)
	log.Println("vaultConnection", string(d))

	if err := api.Flow.Vault.UpdateConnection(ctx, data.Id(), vaultConnection); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readVaultConnection(ctx, data, meta)
}

func deleteVaultConnection(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	if err := api.Flow.Vault.DeleteConnection(ctx, data.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
