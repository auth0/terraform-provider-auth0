package client

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// NewGrantResource will return a new auth0_client_grant resource.
func NewGrantResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createClientGrant,
		ReadContext:   readClientGrant,
		UpdateContext: updateClientGrant,
		DeleteContext: deleteClientGrant,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Auth0 uses various grant types, or methods by which you grant limited access to your " +
			"resources to another entity without exposing credentials. The OAuth 2.0 protocol supports " +
			"several types of grants, which allow different types of access. This resource allows " +
			"you to create and manage client grants used with configured Auth0 clients.",
		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the client for this grant.",
			},
			"audience": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Audience or API Identifier for this grant.",
			},
			"scopes": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringIsNotEmpty,
				},
				Required:    true,
				Description: "Permissions (scopes) included in this grant.",
			},
			"organization_usage": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"allow", "deny", "require",
				}, true),
				Description: "Defines whether organizations can be used with client credentials exchanges " +
					"for this grant. (defaults to deny when not defined)",
			},
			"allow_any_organization": {
				Type:     schema.TypeBool,
				Optional: true,
				Description: "If enabled, any organization can be used with this grant. If disabled (default), " +
					"the grant must be explicitly assigned to the desired organizations.",
			},
		},
	}
}

func createClientGrant(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	grantList, err := api.ClientGrant.List(
		ctx,
		management.Parameter("audience", data.Get("audience").(string)),
		management.Parameter("client_id", data.Get("client_id").(string)),
	)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(grantList.ClientGrants) != 0 {
		data.SetId(grantList.ClientGrants[0].GetID())
		return readClientGrant(ctx, data, meta)
	}

	clientGrant := expandClientGrant(data)

	if err := api.ClientGrant.Create(ctx, clientGrant); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(clientGrant.GetID())

	return readClientGrant(ctx, data, meta)
}

func readClientGrant(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	clientGrant, err := api.ClientGrant.Read(ctx, data.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return diag.FromErr(flattenClientGrant(data, clientGrant))
}

func updateClientGrant(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	if clientGrant := expandClientGrant(data); clientGrantHasChange(clientGrant) {
		if err := api.ClientGrant.Update(ctx, data.Id(), clientGrant); err != nil {
			return diag.FromErr(internalError.HandleAPIError(data, err))
		}
	}

	return readClientGrant(ctx, data, meta)
}

func deleteClientGrant(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	if err := api.ClientGrant.Delete(ctx, data.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}

func clientGrantHasChange(clientGrant *management.ClientGrant) bool {
	// Hacky but we need to tell if an
	// empty json is sent to the api.
	return clientGrant.String() != "{}"
}
