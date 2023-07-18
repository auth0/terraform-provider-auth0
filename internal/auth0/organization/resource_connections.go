package organization

import (
	"context"
	"fmt"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
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

	alreadyEnabledConnections, err := api.Organization.Connections(ctx, organizationID)
	if err != nil {
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

	if len(connectionsToAdd) > len(alreadyEnabledConnections.OrganizationConnections) {
		var result *multierror.Error

		for _, connection := range connectionsToAdd {
			err := api.Organization.AddConnection(ctx, organizationID, connection)
			result = multierror.Append(result, err)
		}

		if result.ErrorOrNil() != nil {
			return diag.FromErr(result.ErrorOrNil())
		}
	}

	return readOrganizationConnections(ctx, data, meta)
}

func readOrganizationConnections(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	connections, err := api.Organization.Connections(ctx, data.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return diag.FromErr(flattenOrganizationConnections(data, connections.OrganizationConnections))
}

func updateOrganizationConnections(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	organizationID := data.Id()

	toAdd, toRemove := value.Difference(data, "enabled_connections")
	var result *multierror.Error

	for _, rmConnection := range toRemove {
		connection := rmConnection.(map[string]interface{})
		connectionID := connection["connection_id"].(string)

		err := api.Organization.DeleteConnection(ctx, organizationID, connectionID)
		if internalError.IsStatusNotFound(err) {
			err = nil
		}

		result = multierror.Append(result, err)
	}

	for _, addConnection := range toAdd {
		connection := addConnection.(map[string]interface{})
		connectionToAdd := &management.OrganizationConnection{
			ConnectionID:            auth0.String(connection["connection_id"].(string)),
			AssignMembershipOnLogin: auth0.Bool(connection["assign_membership_on_login"].(bool)),
		}

		err := api.Organization.AddConnection(ctx, organizationID, connectionToAdd)
		result = multierror.Append(result, err)
	}

	if result.ErrorOrNil() != nil {
		return diag.FromErr(result.ErrorOrNil())
	}

	return readOrganizationConnections(ctx, data, meta)
}

func deleteOrganizationConnections(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	connections := expandOrganizationConnections(data.GetRawState().GetAttr("enabled_connections"))
	var result *multierror.Error

	for _, conn := range connections {
		err := api.Organization.DeleteConnection(ctx, data.Id(), conn.GetConnectionID())
		if internalError.IsStatusNotFound(err) {
			err = nil
		}

		result = multierror.Append(result, err)
	}

	return diag.FromErr(result.ErrorOrNil())
}

func guardAgainstErasingUnwantedConnections(
	organizationID string,
	alreadyEnabledConnections []*management.OrganizationConnection,
	connectionsToAdd []*management.OrganizationConnection,
) diag.Diagnostics {
	if len(alreadyEnabledConnections) == 0 {
		return nil
	}

	alreadyEnabledConnectionsIDs := make([]string, 0)
	for _, conn := range alreadyEnabledConnections {
		alreadyEnabledConnectionsIDs = append(alreadyEnabledConnectionsIDs, conn.GetConnectionID())
	}

	connectionIDsToAdd := make([]string, 0)
	for _, conn := range connectionsToAdd {
		connectionIDsToAdd = append(connectionIDsToAdd, conn.GetConnectionID())
	}

	if cmp.Equal(connectionIDsToAdd, alreadyEnabledConnectionsIDs) {
		return nil
	}

	return diag.Diagnostics{
		diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Organization with non empty enabled connections",
			Detail: cmp.Diff(connectionIDsToAdd, alreadyEnabledConnectionsIDs) +
				fmt.Sprintf("\nThe organization already has enabled connections attached to it. "+
					"Import the resource instead in order to proceed with the changes. "+
					"Run: 'terraform import auth0_organization_connections.<given-name> %s'.", organizationID),
		},
	}
}
