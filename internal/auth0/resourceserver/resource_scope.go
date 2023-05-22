package resourceserver

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
)

// NewScopeResource will return a new auth0_connection_client resource.
func NewScopeResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"scope": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the scope.",
			},
			"resource_server_identifier": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Identifier of the resource server that the scope (permission) is associated with.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the scope (permission).",
			},
		},
		CreateContext: createResourceServerScope,
		UpdateContext: updateResourceServerScope,
		ReadContext:   readResourceServerScope,
		DeleteContext: deleteResourceServerScope,
		Importer: &schema.ResourceImporter{
			StateContext: importResourceServerScope,
		},
		Description: "With this resource, you can manage scopes (permissions) associated with a resource server (API).",
	}
}

func createResourceServerScope(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	resourceServerID := data.Get("resource_server_identifier").(string)
	scope := data.Get("scope").(string)
	description := data.Get("description").(string)

	currentScopes, err := api.ResourceServer.Read(resourceServerID)
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	scopes := append(currentScopes.GetScopes(), management.ResourceServerScope{
		Value:       &scope,
		Description: &description,
	})
	resourceServer := management.ResourceServer{
		Scopes: &scopes,
	}

	if err := api.ResourceServer.Update(resourceServerID, &resourceServer); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(fmt.Sprintf(`%s::%s`, resourceServerID, scope))

	return readResourceServerScope(ctx, data, meta)
}

func updateResourceServerScope(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	resourceServerID := data.Get("resource_server_identifier").(string)
	scope := data.Get("scope").(string)
	newDescription := data.Get("description").(string)

	existingScopes, err := api.ResourceServer.Read(resourceServerID)
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	updatedScopes := []management.ResourceServerScope{}

	found := false
	for _, p := range existingScopes.GetScopes() {
		updated := p
		if p.GetValue() == scope {
			found = true
			updated.Description = &newDescription
		}
		updatedScopes = append(updatedScopes, updated)
	}

	if !found {
		data.SetId("")
		return nil
	}

	if err := api.ResourceServer.Update(resourceServerID, &management.ResourceServer{
		Scopes: &updatedScopes,
	}); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(fmt.Sprintf(`%s::%s`, resourceServerID, scope))

	return readResourceServerScope(ctx, data, meta)
}

func readResourceServerScope(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	resourceServerID := data.Get("resource_server_identifier").(string)
	scope := data.Get("scope").(string)

	existingScopes, err := api.ResourceServer.Read(resourceServerID)
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	for _, p := range existingScopes.GetScopes() {
		if p.GetValue() == scope {
			result := multierror.Append(
				data.Set("description", p.GetDescription()),
			)
			return diag.FromErr(result.ErrorOrNil())
		}
	}

	data.SetId("")
	return nil
}

func deleteResourceServerScope(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	resourceServerID := data.Get("resource_server_identifier").(string)
	scope := data.Get("scope").(string)

	existingScopes, err := api.ResourceServer.Read(resourceServerID)
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	updateScopes := []management.ResourceServerScope{}
	for _, p := range existingScopes.GetScopes() {
		if p.GetValue() != scope {
			updateScopes = append(updateScopes, p)
		}
	}

	if err := api.ResourceServer.Update(
		resourceServerID,
		&management.ResourceServer{
			Scopes: &updateScopes,
		},
	); err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	data.SetId("")
	return nil
}

func importResourceServerScope(
	_ context.Context,
	data *schema.ResourceData,
	_ interface{},
) ([]*schema.ResourceData, error) {
	rawID := data.Id()
	if rawID == "" {
		return nil, fmt.Errorf("ID cannot be empty")
	}

	if !strings.Contains(rawID, "::") {
		return nil, fmt.Errorf("ID must be formatted as <resourceServerIdentifier>::<scope>")
	}

	idPair := strings.Split(rawID, "::")
	if len(idPair) != 2 {
		return nil, fmt.Errorf("ID must be formatted as <resourceServerIdentifier>::<scope>")
	}

	result := multierror.Append(
		data.Set("resource_server_identifier", idPair[0]),
		data.Set("scope", idPair[1]),
	)

	return []*schema.ResourceData{data}, result.ErrorOrNil()
}
