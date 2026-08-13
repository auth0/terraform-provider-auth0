package organization

import (
	"context"

	managementv3 "github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewClientResource will return a new auth0_organization_client resource (EA only).
func NewClientResource() *schema.Resource {
	return &schema.Resource{
		Description: "With this resource, you can manage the association between an organization and an " +
			"application (client), controlling that application's entitlement to the organization (EA only). " +
			"This is distinct from `auth0_organization_client_grant`, which associates an organization with a " +
			"`client_grant` (a client/audience/scopes triple used for client-credentials exchanges).",
		CreateContext: createOrganizationClient,
		ReadContext:   readOrganizationClient,
		UpdateContext: updateOrganizationClient,
		DeleteContext: deleteOrganizationClient,
		Importer: &schema.ResourceImporter{
			StateContext: internalSchema.ImportResourceGroupID("organization_id", "client_id"),
		},
		Schema: clientResourceSchema(),
	}
}

func clientResourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"organization_id": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The ID of the organization to associate the client (application) with.",
		},
		"client_id": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The ID of the client (application) to associate with the organization.",
		},
		"use_for_member_access": {
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
			Description: "Whether this client is used for member access to the organization.",
		},
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The name of the associated client.",
		},
		"app_type": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The type of the associated client application.",
		},
		"logo_uri": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The URI of the associated client's logo.",
		},
		"is_first_party": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Whether the associated client is a first-party client (`true`) or not (`false`).",
		},
		"grant_types": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "The grant types enabled for the associated client.",
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
	}
}

func createOrganizationClient(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	organizationID := data.Get("organization_id").(string)
	clientID := data.Get("client_id").(string)

	createReq := expandOrganizationClientCreate(data)

	if _, err := apiv3.Organizations.Clients.Create(ctx, organizationID, createReq); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	internalSchema.SetResourceGroupID(data, organizationID, clientID)

	return readOrganizationClient(ctx, data, meta)
}

func readOrganizationClient(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	organizationID := data.Get("organization_id").(string)
	clientID := data.Get("client_id").(string)

	organizationClient, err := apiv3.Organizations.Clients.Get(ctx, organizationID, clientID)
	if err != nil {
		return internalError.HandleReadAPIError("auth0_organization_client", data, err)
	}

	return diag.FromErr(flattenOrganizationClient(
		data,
		organizationClient.GetUseForMemberAccess(),
		organizationClient.GetClient(),
	))
}

func updateOrganizationClient(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	organizationID := data.Get("organization_id").(string)
	clientID := data.Get("client_id").(string)

	updateReq := expandOrganizationClientUpdate(data)

	if _, err := apiv3.Organizations.Clients.Update(ctx, organizationID, clientID, updateReq); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readOrganizationClient(ctx, data, meta)
}

func deleteOrganizationClient(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	organizationID := data.Get("organization_id").(string)
	clientID := data.Get("client_id").(string)

	deleteReq := &managementv3.DeleteOrganizationClientsRequestContent{
		Clients: []string{clientID},
	}

	if err := apiv3.Organizations.Clients.Delete(ctx, organizationID, deleteReq); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
