package resourceserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	"github.com/auth0/terraform-provider-auth0/internal/value"
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

	existingResourceServer, err := api.ResourceServer.Read(resourceServerIdentifier)
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	updatedResourceServer := &management.ResourceServer{
		Scopes: expandAPIScopes(data.GetRawConfig().GetAttr("scopes")),
	}

	if diagnostics := guardAgainstErasingUnwantedScopes(
		existingResourceServer.GetIdentifier(),
		existingResourceServer.GetScopes(),
		updatedResourceServer.GetScopes(),
	); diagnostics.HasError() {
		data.SetId("")
		return diagnostics
	}

	if err := api.ResourceServer.Update(resourceServerIdentifier, updatedResourceServer); err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	data.SetId(resourceServerIdentifier)

	return readResourceServerScopes(ctx, data, meta)
}

func readResourceServerScopes(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	resourceServer, err := api.ResourceServer.Read(data.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	result := multierror.Append(
		data.Set("resource_server_identifier", resourceServer.GetIdentifier()),
		data.Set("scopes", flattenAPIScopes(resourceServer.GetScopes())),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updateResourceServerScopes(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	resourceServerIdentifier := data.Get("resource_server_identifier").(string)

	updatedResourceServer := &management.ResourceServer{
		Scopes: expandAPIScopes(data.GetRawConfig().GetAttr("scopes")),
	}

	if err := api.ResourceServer.Update(resourceServerIdentifier, updatedResourceServer); err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	return readResourceServerScopes(ctx, data, meta)
}

func deleteResourceServerScopes(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	resourceServer := &management.ResourceServer{
		Scopes: &[]management.ResourceServerScope{},
	}

	if err := api.ResourceServer.Update(data.Id(), resourceServer); err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	data.SetId("")

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

func expandAPIScopes(scopes cty.Value) *[]management.ResourceServerScope {
	resourceServerScopes := make([]management.ResourceServerScope, 0)

	scopes.ForEachElement(func(_ cty.Value, scope cty.Value) (stop bool) {
		resourceServerScopes = append(resourceServerScopes, management.ResourceServerScope{
			Value:       value.String(scope.GetAttr("name")),
			Description: value.String(scope.GetAttr("description")),
		})

		return stop
	})

	return &resourceServerScopes
}

func flattenAPIScopes(resourceServerScopes []management.ResourceServerScope) []map[string]interface{} {
	scopes := make([]map[string]interface{}, len(resourceServerScopes))

	for index, scope := range resourceServerScopes {
		scopes[index] = map[string]interface{}{
			"name":        scope.GetValue(),
			"description": scope.GetDescription(),
		}
	}

	return scopes
}
