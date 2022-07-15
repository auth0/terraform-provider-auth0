package auth0

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/auth0/internal/hash"
)

func newOrganization() *schema.Resource {
	return &schema.Resource{
		CreateContext: createOrganization,
		ReadContext:   readOrganization,
		UpdateContext: updateOrganization,
		DeleteContext: deleteOrganization,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of this organization",
			},
			"display_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Friendly name of this organization",
			},
			"branding": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				MinItems:    1,
				Description: "Defines how to style the login pages",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"logo_url": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"colors": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"metadata": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Metadata associated with the organization, Maximum of 10 metadata properties allowed",
			},
			"connections": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connection_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"assign_membership_on_login": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
				Set: hash.StringKey("connection_id"),
			},
			"users": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"role": {
							Type: schema.TypeSet,
							Elem: schema.TypeString,
						},
					},
				},
			},
		},
	}
}

func createOrganization(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	organization := buildOrganization(d)
	api := m.(*management.Management)
	if err := api.Organization.Create(organization); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(organization.GetID())

	d.Partial(true)
	if err := assignOrganizationConnections(d, m); err != nil {
		return diag.FromErr(fmt.Errorf("failed assigning organization connections. %w", err))
	}
	d.Partial(false)

	return readOrganization(ctx, d, m)
}

func assignOrganizationConnections(d *schema.ResourceData, m interface{}) (err error) {
	api := m.(*management.Management)

	add, rm := Diff(d, "connections")

	add.Elem(func(data ResourceData) {
		organizationConnection := &management.OrganizationConnection{
			ConnectionID:            String(data, "connection_id"),
			AssignMembershipOnLogin: Bool(data, "assign_membership_on_login"),
		}

		log.Printf("[DEBUG] (+) auth0_organization.%s.connections.%s", d.Id(), organizationConnection.GetConnectionID())

		err = api.Organization.AddConnection(d.Id(), organizationConnection)
		if err != nil {
			return
		}
	})

	rm.Elem(func(data ResourceData) {
		// Take connectionID before it changed (i.e. removed).
		// Therefore we use GetChange() instead of the typical Get().
		connectionID, _ := data.GetChange("connection_id")

		log.Printf("[DEBUG] (-) auth0_organization.%s.connections.%s", d.Id(), connectionID.(string))

		err = api.Organization.DeleteConnection(d.Id(), connectionID.(string))
		if err != nil {
			return
		}
	})

	// Update existing connections if any mutable properties have changed.
	Set(d, "connections", HasChange()).Elem(func(data ResourceData) {
		connectionID := data.Get("connection_id").(string)
		organizationConnection := &management.OrganizationConnection{
			AssignMembershipOnLogin: Bool(data, "assign_membership_on_login"),
		}

		log.Printf("[DEBUG] (~) auth0_organization.%s.connections.%s", d.Id(), connectionID)

		err = api.Organization.UpdateConnection(d.Id(), connectionID, organizationConnection)
		if err != nil {
			return
		}
	})

	return nil
}

func readOrganization(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)
	organization, err := api.Organization.Read(d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return diag.FromErr(err)
	}

	organizationConnectionList, err := api.Organization.Connections(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	result := multierror.Append(
		d.Set("name", organization.Name),
		d.Set("display_name", organization.DisplayName),
		d.Set("branding", flattenOrganizationBranding(organization.Branding)),
		d.Set("metadata", organization.Metadata),
		d.Set("connections", flattenOrganizationConnections(organizationConnectionList)),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updateOrganization(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	organization := buildOrganization(d)
	api := m.(*management.Management)
	if err := api.Organization.Update(d.Id(), organization); err != nil {
		return diag.FromErr(err)
	}

	d.Partial(true)
	if err := assignOrganizationConnections(d, m); err != nil {
		return diag.FromErr(fmt.Errorf("failed updating organization connections. %w", err))
	}
	d.Partial(false)

	return readOrganization(ctx, d, m)
}

func deleteOrganization(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)
	if err := api.Organization.Delete(d.Id()); err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
	}

	return nil
}

func buildOrganization(d *schema.ResourceData) *management.Organization {
	organization := &management.Organization{
		Name:        String(d, "name"),
		DisplayName: String(d, "display_name"),
		Metadata:    Map(d, "metadata"),
	}

	List(d, "branding").Elem(func(d ResourceData) {
		organization.Branding = &management.OrganizationBranding{
			LogoURL: String(d, "logo_url"),
			Colors:  Map(d, "colors"),
		}
	})

	return organization
}

func flattenOrganizationBranding(organizationBranding *management.OrganizationBranding) []interface{} {
	if organizationBranding == nil {
		return nil
	}
	return []interface{}{
		map[string]interface{}{
			"logo_url": organizationBranding.LogoURL,
			"colors":   organizationBranding.Colors,
		},
	}
}

func flattenOrganizationConnections(organizationConnectionList *management.OrganizationConnectionList) []interface{} {
	if organizationConnectionList == nil {
		return nil
	}

	connections := make([]interface{}, len(organizationConnectionList.OrganizationConnections))
	for index, connection := range organizationConnectionList.OrganizationConnections {
		connections[index] = map[string]interface{}{
			"connection_id":              connection.ConnectionID,
			"assign_membership_on_login": connection.AssignMembershipOnLogin,
		}
	}

	return connections
}
