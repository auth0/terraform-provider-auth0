package provider

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func newClientGrant() *schema.Resource {
	return &schema.Resource{
		CreateContext: createClientGrant,
		ReadContext:   readClientGrant,
		UpdateContext: updateClientGrant,
		DeleteContext: deleteClientGrant,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"audience": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"scope": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
			},
		},
	}
}

func createClientGrant(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clientGrant := buildClientGrant(d)
	api := m.(*management.Management)
	if err := api.ClientGrant.Create(clientGrant); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(auth0.StringValue(clientGrant.ID))

	return readClientGrant(ctx, d, m)
}

func readClientGrant(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)
	clientGrant, err := api.ClientGrant.Read(d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return diag.FromErr(err)
	}

	result := multierror.Append(
		d.Set("client_id", clientGrant.ClientID),
		d.Set("audience", clientGrant.Audience),
		d.Set("scope", clientGrant.Scope),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updateClientGrant(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clientGrant := buildClientGrant(d)
	clientGrant.Audience = nil
	clientGrant.ClientID = nil
	api := m.(*management.Management)
	if err := api.ClientGrant.Update(d.Id(), clientGrant); err != nil {
		return diag.FromErr(err)
	}

	return readClientGrant(ctx, d, m)
}

func deleteClientGrant(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)
	if err := api.ClientGrant.Delete(d.Id()); err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return diag.FromErr(err)
	}

	return nil
}

func buildClientGrant(d *schema.ResourceData) *management.ClientGrant {
	clientGrant := &management.ClientGrant{
		ClientID: String(d, "client_id"),
		Audience: String(d, "audience"),
	}

	clientGrant.Scope = []interface{}{}
	if scope, ok := d.GetOk("scope"); ok {
		clientGrant.Scope = scope.([]interface{})
	}

	return clientGrant
}
