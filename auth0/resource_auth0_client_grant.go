package auth0

import (
	"net/http"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func newClientGrant() *schema.Resource {
	return &schema.Resource{
		Create: createClientGrant,
		Read:   readClientGrant,
		Update: updateClientGrant,
		Delete: deleteClientGrant,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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

func createClientGrant(d *schema.ResourceData, m interface{}) error {
	clientGrant := buildClientGrant(d)
	api := m.(*management.Management)
	if err := api.ClientGrant.Create(clientGrant); err != nil {
		return err
	}

	d.SetId(auth0.StringValue(clientGrant.ID))

	return readClientGrant(d, m)
}

func readClientGrant(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)
	clientGrant, err := api.ClientGrant.Read(d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.Set("client_id", clientGrant.ClientID)
	d.Set("audience", clientGrant.Audience)
	d.Set("scope", clientGrant.Scope)

	return nil
}

func updateClientGrant(d *schema.ResourceData, m interface{}) error {
	clientGrant := buildClientGrant(d)
	clientGrant.Audience = nil
	clientGrant.ClientID = nil
	api := m.(*management.Management)
	if err := api.ClientGrant.Update(d.Id(), clientGrant); err != nil {
		return err
	}

	return readClientGrant(d, m)
}

func deleteClientGrant(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)
	if err := api.ClientGrant.Delete(d.Id()); err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return err
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
