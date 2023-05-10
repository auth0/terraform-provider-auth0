package client

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	"github.com/auth0/terraform-provider-auth0/internal/value"
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
			"scope": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    true,
				Description: "Permissions (scopes) included in this grant.",
			},
		},
	}
}

func createClientGrant(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	grantList, err := api.ClientGrant.List(
		management.Parameter("audience", d.Get("audience").(string)),
		management.Parameter("client_id", d.Get("client_id").(string)),
	)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(grantList.ClientGrants) != 0 {
		d.SetId(grantList.ClientGrants[0].GetID())
		return readClientGrant(ctx, d, m)
	}

	clientGrant := expandClientGrant(d)
	if err := api.ClientGrant.Create(clientGrant); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(clientGrant.GetID())

	return readClientGrant(ctx, d, m)
}

func readClientGrant(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	clientGrant, err := api.ClientGrant.Read(d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	result := multierror.Append(
		d.Set("client_id", clientGrant.GetClientID()),
		d.Set("audience", clientGrant.GetAudience()),
		d.Set("scope", clientGrant.Scope),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updateClientGrant(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	clientGrant := expandClientGrant(d)
	if clientGrantHasChange(clientGrant) {
		if err := api.ClientGrant.Update(d.Id(), clientGrant); err != nil {
			return diag.FromErr(err)
		}
	}

	return readClientGrant(ctx, d, m)
}

func deleteClientGrant(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	if err := api.ClientGrant.Delete(d.Id()); err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func expandClientGrant(d *schema.ResourceData) *management.ClientGrant {
	config := d.GetRawConfig()

	clientGrant := &management.ClientGrant{}

	if d.IsNewResource() {
		clientGrant.ClientID = value.String(config.GetAttr("client_id"))
		clientGrant.Audience = value.String(config.GetAttr("audience"))
	}

	if d.IsNewResource() || d.HasChange("scope") {
		scopeListFromConfig := d.Get("scope").([]interface{})
		scopeList := make([]string, 0)
		for _, scope := range scopeListFromConfig {
			scopeList = append(scopeList, scope.(string))
		}
		clientGrant.Scope = scopeList
	}

	return clientGrant
}

func clientGrantHasChange(clientGrant *management.ClientGrant) bool {
	// Hacky but we need to tell if an
	// empty json is sent to the api.
	return clientGrant.String() != "{}"
}
