package connection

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"

	"github.com/auth0/go-auth0/management"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// NewSCIMTokenResource will return a new auth0_connection_scim_token resource.
func NewSCIMTokenResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createSCIMToken,
		ReadContext:   readSCIMToken,
		DeleteContext: deleteSCIMToken,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can create and manage SCIM bearer tokens for a connection. " +
			"This resource only works with enterprise connections",
		Schema: map[string]*schema.Schema{
			"connection_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the connection for this SCIM token.",
			},
			"scopes": {
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The scopes associated with the SCIM token.",
			},
			"token": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The SCIM bearer token value.",
			},
			"token_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the SCIM token.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time when the token was created (ISO8601 format).",
			},
		},
	}
}

func createSCIMToken(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	connectionID := data.Get("connection_id").(string)

	scimToken := &management.SCIMToken{
		Scopes: value.Strings(data.GetRawConfig().GetAttr("scopes")),
	}

	if err := api.Connection.CreateSCIMToken(ctx, connectionID, scimToken); err != nil {
		return diag.FromErr(err)
	}

	if scimToken.TokenID != nil {
		data.SetId(scimToken.GetTokenID())
	} else {
		return diag.Diagnostics{{
			Severity: diag.Error,
			Summary:  "Failed to create SCIM token",
			Detail:   "Token ID was not returned from the API",
		}}
	}

	_ = data.Set("connection_id", connectionID)

	return readSCIMToken(ctx, data, meta)
}

func readSCIMToken(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	connectionID := data.Get("connection_id").(string)
	tokenID := data.Id()

	// List all tokens for the connection and find the one we're looking for.
	scimTokens, err := api.Connection.ListSCIMToken(ctx, connectionID)
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	var foundToken *management.SCIMToken
	for _, token := range scimTokens {
		if token.TokenID != nil && *token.TokenID == tokenID {
			foundToken = token
			break
		}
	}

	if foundToken == nil {
		data.SetId("")
		return diag.Diagnostics{{
			Severity: diag.Warning,
			Summary:  "SCIM token not found",
			Detail:   "The SCIM token was not found for the connection. It may have been deleted.",
		}}
	}

	return flattenSCIMToken(data, foundToken)
}

func deleteSCIMToken(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	connectionID := data.Get("connection_id").(string)
	tokenID := data.Id()

	if err := api.Connection.DeleteSCIMToken(ctx, connectionID, tokenID); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}

func flattenSCIMToken(data *schema.ResourceData, scimToken *management.SCIMToken) diag.Diagnostics {
	result := multierror.Append(
		data.Set("token", scimToken.GetToken()),
		data.Set("token_id", scimToken.GetTokenID()),
		data.Set("scopes", scimToken.GetScopes()),
		data.Set("created_at", scimToken.GetCreatedAt()),
	)

	return diag.FromErr(result.ErrorOrNil())
}
