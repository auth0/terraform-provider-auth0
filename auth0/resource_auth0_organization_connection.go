package auth0

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func newOrganizationConnection() *schema.Resource {
	return &schema.Resource{
		CreateContext: createOrganizationConnection,
		ReadContext:   readOrganizationConnection,
		UpdateContext: updateOrganizationConnection,
		DeleteContext: deleteOrganizationConnection,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"organization_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"connection_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"assign_membership_on_login": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"strategy": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func createOrganizationConnection(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*management.Management)

	organizationID := data.Get("organization_id").(string)
	organizationConnection := &management.OrganizationConnection{
		ConnectionID:            String(data, "connection_id"),
		AssignMembershipOnLogin: Bool(data, "assign_membership_on_login"),
	}

	if err := api.Organization.AddConnection(organizationID, organizationConnection); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(resource.UniqueId())

	return readOrganizationConnection(ctx, data, meta)
}

func readOrganizationConnection(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*management.Management)

	organizationID := data.Get("organization_id").(string)
	connectionID := data.Get("connection_id").(string)
	organizationConnection, err := api.Organization.Connection(organizationID, connectionID)
	if err != nil {
		return diag.FromErr(err)
	}

	result := multierror.Append(
		data.Set("assign_membership_on_login", organizationConnection.GetAssignMembershipOnLogin()),
		data.Set("name", organizationConnection.GetConnection().GetName()),
		data.Set("strategy", organizationConnection.GetConnection().GetStrategy()),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updateOrganizationConnection(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*management.Management)

	organizationID := data.Get("organization_id").(string)
	connectionID := data.Get("connection_id").(string)
	organizationConnection := &management.OrganizationConnection{
		AssignMembershipOnLogin: Bool(data, "assign_membership_on_login"),
	}
	if err := api.Organization.UpdateConnection(organizationID, connectionID, organizationConnection); err != nil {
		return diag.FromErr(err)
	}

	return readOrganizationConnection(ctx, data, meta)
}

func deleteOrganizationConnection(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*management.Management)

	organizationID := data.Get("organization_id").(string)
	connectionID := data.Get("connection_id").(string)

	if err := api.Organization.DeleteConnection(organizationID, connectionID); err != nil {
		if err, ok := err.(management.Error); ok && err.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	data.SetId("")

	return nil
}
