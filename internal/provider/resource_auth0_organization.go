package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
				Deprecated: "Management of organizations through this property has been deprecated in favor of the " +
					"`auth0_organization_connection` resource and will be deleted in future versions. It is " +
					"advised to migrate all managed organization connections to the new resource type.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connection_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"assign_membership_on_login": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
		},
	}
}

func createOrganization(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	organization := expandOrganization(d)

	api := m.(*management.Management)
	if err := api.Organization.Create(organization); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(organization.GetID())

	if err := updateOrganizationConnections(d, api); err != nil {
		return diag.FromErr(fmt.Errorf("failed to update organization connections: %w", err))
	}

	return readOrganization(ctx, d, m)
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
	organization := expandOrganization(d)

	api := m.(*management.Management)
	if err := api.Organization.Update(d.Id(), organization); err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("connections") {
		if err := updateOrganizationConnections(d, api); err != nil {
			return diag.FromErr(fmt.Errorf("failed to update organization connections: %w", err))
		}
	}

	return readOrganization(ctx, d, m)
}

func deleteOrganization(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)
	if err := api.Organization.Delete(d.Id()); err != nil {
		if err, ok := err.(management.Error); ok && err.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}

func updateOrganizationConnections(d *schema.ResourceData, api *management.Management) error {
	toAdd, toRemove := Diff(d, "connections")

	connectionOperations := make(map[string]string)
	toRemove.Elem(func(data ResourceData) {
		oldConnectionID, _ := data.GetChange("connection_id")
		connectionOperations[oldConnectionID.(string)] = "deleteConnection"
	})

	toAdd.Elem(func(data ResourceData) {
		newConnectionID := data.Get("connection_id").(string)

		if _, ok := connectionOperations[newConnectionID]; ok {
			delete(connectionOperations, newConnectionID)
		} else {
			connectionOperations[newConnectionID] = "addConnection"
		}
	})

	for key, value := range connectionOperations {
		if value == "deleteConnection" {
			if err := api.Organization.DeleteConnection(d.Id(), key); err != nil {
				return err
			}
		}
		if value == "addConnection" {
			organizationConnection := &management.OrganizationConnection{
				ConnectionID: auth0.String(key),
			}
			if err := api.Organization.AddConnection(d.Id(), organizationConnection); err != nil {
				return err
			}
		}
	}

	var err error
	Set(d, "connections").Elem(func(data ResourceData) {
		connectionID := data.Get("connection_id").(string)
		organizationConnection := &management.OrganizationConnection{
			AssignMembershipOnLogin: Bool(data, "assign_membership_on_login"),
		}

		err = api.Organization.UpdateConnection(d.Id(), connectionID, organizationConnection)
		if err != nil {
			return
		}
	})

	return err
}

func expandOrganization(d *schema.ResourceData) *management.Organization {
	organization := &management.Organization{
		Name:        String(d, "name"),
		DisplayName: String(d, "display_name"),
	}

	if d.HasChange("metadata") {
		metadata := Map(d, "metadata")
		organization.Metadata = &metadata
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
			"logo_url": organizationBranding.GetLogoURL(),
			"colors":   organizationBranding.Colors,
		},
	}
}

func flattenOrganizationConnections(connectionList *management.OrganizationConnectionList) []interface{} {
	if connectionList == nil {
		return nil
	}

	connections := make([]interface{}, len(connectionList.OrganizationConnections))
	for index, connection := range connectionList.OrganizationConnections {
		connections[index] = map[string]interface{}{
			"connection_id":              connection.GetConnectionID(),
			"assign_membership_on_login": connection.GetAssignMembershipOnLogin(),
		}
	}

	return connections
}
