package organization

import (
	"context"
	"errors"
	"fmt"

	"github.com/auth0/go-auth0"
	managementv2 "github.com/auth0/go-auth0/v2/management"
	"github.com/auth0/go-auth0/v2/management/core"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

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
							Description: "When `true`, all users that log in with this connection will be " +
								"automatically granted membership in the organization. When `false`, users must be " +
								"granted membership in the organization before logging in with this connection.",
						},
						"is_signup_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
							Description: "Determines whether organization sign-up should be enabled for this " +
								"organization connection. Only applicable for database connections. " +
								"Note: `is_signup_enabled` can only be `true` if `assign_membership_on_login` is `true`.",
						},
						"show_as_button": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
							Description: "Determines whether a connection should be displayed on this organization's " +
								"login prompt. Only applicable for enterprise connections.",
						},
						"is_enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Whether the connection is enabled for the organization.",
						},
						"organization_connection_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Name of the connection in the scope of this organization.",
						},
						"organization_access_level": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ValidateFunc: validation.StringInSlice([]string{
								"member", "non_member", "everyone",
							}, false),
							Description: "The access level for this organization connection. Can be `member`, `non_member`, or `everyone`.",
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
	apiv2 := meta.(*config.Config).GetAPIV2()

	organizationID := data.Get("organization_id").(string)

	var alreadyEnabledConnections []*managementv2.OrganizationAllConnectionPost
	page, err := apiv2.Organizations.Connections.List(ctx,
		organizationID,
		&managementv2.ListOrganizationAllConnectionsRequestParameters{IsEnabled: auth0.Bool(true)},
	)
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}
	alreadyEnabledConnections = append(alreadyEnabledConnections, page.Results...)
	for {
		page, err = page.GetNextPage(ctx)
		if err != nil {
			if errors.Is(err, core.ErrNoPages) {
				break
			}
			return diag.FromErr(internalError.HandleAPIError(data, err))
		}
		alreadyEnabledConnections = append(alreadyEnabledConnections, page.Results...)
	}

	data.SetId(organizationID)

	connectionsToAdd := expandOrganizationConnectionsCreate(data.GetRawConfig().GetAttr("enabled_connections"))

	if diagnostics := guardAgainstErasingUnwantedConnectionsV2(
		organizationID,
		alreadyEnabledConnections,
		connectionsToAdd,
	); diagnostics.HasError() {
		data.SetId("")
		return diagnostics
	}

	if len(connectionsToAdd) > len(alreadyEnabledConnections) {
		var result *multierror.Error

		for _, connection := range connectionsToAdd {
			_, err := apiv2.Organizations.Connections.Create(ctx, organizationID, connection)
			result = multierror.Append(result, err)
		}

		if result.ErrorOrNil() != nil {
			return diag.FromErr(result.ErrorOrNil())
		}
	}

	return readOrganizationConnections(ctx, data, meta)
}

func readOrganizationConnections(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()

	var connections []*managementv2.OrganizationAllConnectionPost
	page, err := apiv2.Organizations.Connections.List(ctx,
		data.Id(),
		&managementv2.ListOrganizationAllConnectionsRequestParameters{IsEnabled: auth0.Bool(true)})
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}
	connections = append(connections, page.Results...)
	for {
		page, err = page.GetNextPage(ctx)
		if err != nil {
			if errors.Is(err, core.ErrNoPages) {
				break
			}
			return diag.FromErr(internalError.HandleAPIError(data, err))
		}
		connections = append(connections, page.Results...)
	}

	return diag.FromErr(flattenOrganizationConnectionsV2(data, connections))
}

func updateOrganizationConnections(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()

	organizationID := data.Id()

	connections := expandOrganizationConnectionsCreate(data.GetRawConfig().GetAttr("enabled_connections"))
	connectionMap := make(map[string]*managementv2.CreateOrganizationAllConnectionRequestParameters)
	for _, connection := range connections {
		connectionMap[connection.ConnectionID] = connection
	}

	toAdd, toRemove := value.Difference(data, "enabled_connections")
	var result *multierror.Error

	for _, rmConnection := range toRemove {
		connection := rmConnection.(map[string]interface{})
		connectionID := connection["connection_id"].(string)

		err := apiv2.Organizations.Connections.Delete(ctx, organizationID, connectionID)
		if internalError.IsStatusNotFound(err) {
			err = nil
		}

		result = multierror.Append(result, err)
	}

	for _, addConnection := range toAdd {
		connection := addConnection.(map[string]interface{})
		connectionID := connection["connection_id"].(string)

		_, err := apiv2.Organizations.Connections.Create(ctx, organizationID, connectionMap[connectionID])
		result = multierror.Append(result, err)
	}

	if result.ErrorOrNil() != nil {
		return diag.FromErr(result.ErrorOrNil())
	}

	return readOrganizationConnections(ctx, data, meta)
}

func deleteOrganizationConnections(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()

	connections := expandOrganizationConnectionsCreate(data.GetRawState().GetAttr("enabled_connections"))
	var result *multierror.Error

	for _, conn := range connections {
		err := apiv2.Organizations.Connections.Delete(ctx, data.Id(), conn.ConnectionID)
		if internalError.IsStatusNotFound(err) {
			err = nil
		}

		result = multierror.Append(result, err)
	}

	return diag.FromErr(result.ErrorOrNil())
}

func guardAgainstErasingUnwantedConnectionsV2(
	organizationID string,
	alreadyEnabledConnections []*managementv2.OrganizationAllConnectionPost,
	connectionsToAdd []*managementv2.CreateOrganizationAllConnectionRequestParameters,
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
		connectionIDsToAdd = append(connectionIDsToAdd, conn.ConnectionID)
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
