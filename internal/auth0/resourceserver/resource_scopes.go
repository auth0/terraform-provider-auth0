package resourceserver

import (
	"context"
	"fmt"

	"github.com/auth0/go-auth0/management"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// NewScopesResource will return a new auth0_resource_server_scopes (1:many) resource.
func NewScopesResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"resource_server_identifier": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Identifier of the resource server that the scopes (permission) are associated with.",
			},
			"scopes": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							Description: "Name of the scope (permission). Examples include " +
								"`read:appointments` or `delete:appointments`.",
						},
						"description": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     nil,
							Description: "User-friendly description of the scope (permission).",
						},
					},
				},
			},
		},
		CreateContext: createResourceServerScopes,
		ReadContext:   readResourceServerScopes,
		UpdateContext: updateResourceServerScopes,
		DeleteContext: deleteResourceServerScopes,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can manage scopes (permissions) associated with a resource server (API).",
	}
}

func createResourceServerScopes(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	resourceServerIdentifier := data.Get("resource_server_identifier").(string)

	existingResourceServer, err := api.ResourceServer.Read(ctx, resourceServerIdentifier)
	if err != nil {
		return diag.FromErr(err)
	}

	updatedResourceServer := &management.ResourceServer{
		Scopes: expandResourceServerScopes(data.GetRawConfig().GetAttr("scopes")),
	}

	if diagnostics := guardAgainstErasingUnwantedScopes(
		existingResourceServer.GetIdentifier(),
		existingResourceServer.GetScopes(),
		updatedResourceServer.GetScopes(),
	); diagnostics.HasError() {
		data.SetId("")
		return diagnostics
	}

	if err := api.ResourceServer.Update(ctx, resourceServerIdentifier, updatedResourceServer); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(resourceServerIdentifier)

	return readResourceServerScopes(ctx, data, meta)
}

func readResourceServerScopes(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	resourceServer, err := api.ResourceServer.Read(ctx, data.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return diag.FromErr(flattenResourceServerScopes(data, resourceServer))
}

func updateResourceServerScopes(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	resourceServerIdentifier := data.Get("resource_server_identifier").(string)

	updatedResourceServer := &management.ResourceServer{
		Scopes: expandResourceServerScopes(data.GetRawConfig().GetAttr("scopes")),
	}

	if err := api.ResourceServer.Update(ctx, resourceServerIdentifier, updatedResourceServer); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readResourceServerScopes(ctx, data, meta)
}

func deleteResourceServerScopes(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	resourceServer := &management.ResourceServer{
		Scopes: &[]management.ResourceServerScope{},
	}

	if err := api.ResourceServer.Update(ctx, data.Id(), resourceServer); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}

func guardAgainstErasingUnwantedScopes(
	apiIdentifier string,
	apiScopes []management.ResourceServerScope,
	configScopes []management.ResourceServerScope,
) diag.Diagnostics {
	if len(apiScopes) == 0 {
		return nil
	}

	if cmp.Equal(configScopes, apiScopes) {
		return nil
	}

	return diag.Diagnostics{
		diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Resource Server with non empty scopes",
			Detail: cmp.Diff(configScopes, apiScopes) +
				fmt.Sprintf("\nThe resource server already has scopes attached to it. "+
					"Import the resource instead in order to proceed with the changes. "+
					"Run: 'terraform import auth0_resource_server_scopes.<given-name> %s'.", apiIdentifier),
		},
	}
}
