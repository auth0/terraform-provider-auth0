package organization

import (
	"context"
	"fmt"
	"net/http"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	"github.com/auth0/terraform-provider-auth0/internal/value"
)

// NewConnectionsResource will return a new auth0_organization_connections (1:many) resource.
func NewConnectionsResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"organization_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the organization on which to enable the connections.",
			},
			"enabled_connections": {
				Type: schema.TypeSet,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connection_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The ID of the connection to enable for the organization.",
						},
						"assign_membership_on_login": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
							Description: "When true, all users that log in with this connection will be " +
								"automatically granted membership in the organization. When false, users must be " +
								"granted membership in the organization before logging in with this connection.",
						},
					},
				},
				Required:    true,
				Description: "Connections that are enabled for the organization.",
			},
		},
		CreateContext: createOrganizationConnections,
		ReadContext:   readOrganizationConnections,
		UpdateContext: updateOrganizationConnections,
		DeleteContext: deleteOrganizationConnections,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can manage enabled connections on an organization.",
	}
}

func createOrganizationConnections(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	organizationID := data.Get("organization_id").(string)

	alreadyEnabledConnections, err := api.Organization.Connections(organizationID)
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	data.SetId(organizationID)

	connectionsToAdd := expandOrganizationConnections(data.GetRawConfig().GetAttr("enabled_connections"))

	if diagnostics := guardAgainstErasingUnwantedConnections(
		organizationID,
		alreadyEnabledConnections.OrganizationConnections,
		connectionsToAdd,
	); diagnostics.HasError() {
		data.SetId("")
		return diagnostics
	}

	var result *multierror.Error
	for _, connection := range connectionsToAdd {
		if err := api.Organization.AddConnection(organizationID, connection); err != nil {
			result = multierror.Append(result, err)
		}
	}

	if result.ErrorOrNil() != nil {
		return diag.FromErr(result.ErrorOrNil())
	}

	return readOrganizationConnections(ctx, data, meta)
}

func readOrganizationConnections(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	connections, err := api.Organization.Connections(data.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	result := multierror.Append(
		data.Set("organization_id", data.Id()),
		data.Set("enabled_connections", flattenOrganizationConnections(connections.OrganizationConnections)),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updateOrganizationConnections(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	organizationID := data.Id()

	var result *multierror.Error
	toAdd, toRemove := value.Difference(data, "enabled_connections")

	for _, rmConnection := range toRemove {
		connection := rmConnection.(map[string]interface{})

		if err := api.Organization.DeleteConnection(organizationID, connection["connection_id"].(string)); err != nil {
			if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
				data.SetId("")
				return nil
			}

			result = multierror.Append(result, err)
		}
	}

	for _, addConnection := range toAdd {
		connection := addConnection.(map[string]interface{})

		if err := api.Organization.AddConnection(organizationID, &management.OrganizationConnection{
			ConnectionID:            auth0.String(connection["connection_id"].(string)),
			AssignMembershipOnLogin: auth0.Bool(connection["assign_membership_on_login"].(bool)),
		}); err != nil {
			if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
				data.SetId("")
				return nil
			}

			result = multierror.Append(result, err)
		}
	}

	if result.ErrorOrNil() != nil {
		return diag.FromErr(result.ErrorOrNil())
	}

	return readOrganizationConnections(ctx, data, meta)
}

func deleteOrganizationConnections(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	organizationID := data.Id()

	var result *multierror.Error

	connections := expandOrganizationConnections(data.GetRawState().GetAttr("enabled_connections"))
	for _, conn := range connections {
		err := api.Organization.DeleteConnection(organizationID, conn.GetConnectionID())
		result = multierror.Append(result, err)
	}

	if result.ErrorOrNil() != nil {
		return diag.FromErr(result.ErrorOrNil())
	}

	data.SetId("")
	return nil
}

func guardAgainstErasingUnwantedConnections(
	organizationID string,
	alreadyEnabledConnections []*management.OrganizationConnection,
	connectionsToAdd []*management.OrganizationConnection,
) diag.Diagnostics {
	if len(alreadyEnabledConnections) == 0 {
		return nil
	}

	return diag.Diagnostics{
		diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Organization with non empty enabled connections",
			Detail: cmp.Diff(connectionsToAdd, alreadyEnabledConnections) +
				fmt.Sprintf("\nThe organization already has enabled connections attached to it. "+
					"Import the resource instead in order to proceed with the changes. "+
					"Run: 'terraform import auth0_organization_connections.<given-name> %s'.", organizationID),
		},
	}
}

func expandOrganizationConnections(cfg cty.Value) []*management.OrganizationConnection {
	connections := make([]*management.OrganizationConnection, 0)

	cfg.ForEachElement(func(_ cty.Value, connectionCfg cty.Value) (stop bool) {
		connections = append(connections, &management.OrganizationConnection{
			ConnectionID:            value.String(connectionCfg.GetAttr("connection_id")),
			AssignMembershipOnLogin: value.Bool(connectionCfg.GetAttr("assign_membership_on_login")),
		})

		return stop
	})

	return connections
}
