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
				Description: "Name of the scope (permission).",
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
	mutex := meta.(*config.Config).GetMutex()

	resourceServerIdentifier := data.Get("resource_server_identifier").(string)
	scope := data.Get("scope").(string)
	description := data.Get("description").(string)

	mutex.Lock(resourceServerIdentifier)
	defer mutex.Unlock(resourceServerIdentifier)

	existingAPI, err := api.ResourceServer.Read(resourceServerIdentifier)
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	data.SetId(fmt.Sprintf(`%s::%s`, resourceServerIdentifier, scope))

	for _, apiScope := range existingAPI.GetScopes() {
		if apiScope.GetValue() == scope {
			return readResourceServerScope(ctx, data, meta)
		}
	}

	scopes := append(existingAPI.GetScopes(), management.ResourceServerScope{
		Value:       &scope,
		Description: &description,
	})
	resourceServer := management.ResourceServer{
		Scopes: &scopes,
	}

	if err := api.ResourceServer.Update(resourceServerIdentifier, &resourceServer); err != nil {
		return diag.FromErr(err)
	}

	return readResourceServerScope(ctx, data, meta)
}

func updateResourceServerScope(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	mutex := meta.(*config.Config).GetMutex()

	resourceServerIdentifier := data.Get("resource_server_identifier").(string)
	scope := data.Get("scope").(string)
	newDescription := data.Get("description").(string)

	mutex.Lock(resourceServerIdentifier)
	defer mutex.Unlock(resourceServerIdentifier)

	existingAPI, err := api.ResourceServer.Read(resourceServerIdentifier)
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	updatedScopes := make([]management.ResourceServerScope, 0)

	found := false
	for _, existingScope := range existingAPI.GetScopes() {
		updated := existingScope
		if existingScope.GetValue() == scope {
			found = true
			updated.Description = &newDescription
		}
		updatedScopes = append(updatedScopes, updated)
	}

	if !found {
		data.SetId("")
		return nil
	}

	if err := api.ResourceServer.Update(resourceServerIdentifier, &management.ResourceServer{
		Scopes: &updatedScopes,
	}); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(fmt.Sprintf(`%s::%s`, resourceServerIdentifier, scope))

	return readResourceServerScope(ctx, data, meta)
}

func readResourceServerScope(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	resourceServerID := data.Get("resource_server_identifier").(string)
	scope := data.Get("scope").(string)

	existingAPI, err := api.ResourceServer.Read(resourceServerID)
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	for _, existingScope := range existingAPI.GetScopes() {
		if existingScope.GetValue() == scope {
			err := data.Set("description", existingScope.GetDescription())
			return diag.FromErr(err)
		}
	}

	data.SetId("")
	return nil
}

func deleteResourceServerScope(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	mutex := meta.(*config.Config).GetMutex()

	resourceServerIdentifier := data.Get("resource_server_identifier").(string)
	scope := data.Get("scope").(string)

	mutex.Lock(resourceServerIdentifier)
	defer mutex.Unlock(resourceServerIdentifier)

	existingAPI, err := api.ResourceServer.Read(resourceServerIdentifier)
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	updateScopes := make([]management.ResourceServerScope, 0)
	for _, existingScope := range existingAPI.GetScopes() {
		if existingScope.GetValue() != scope {
			updateScopes = append(updateScopes, existingScope)
		}
	}

	if err := api.ResourceServer.Update(
		resourceServerIdentifier,
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
