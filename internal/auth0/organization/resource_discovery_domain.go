package organization

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewDiscoveryDomainResource will return a new auth0_organization_discovery_domain resource.
func NewDiscoveryDomainResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createOrganizationDiscoveryDomain,
		ReadContext:   readOrganizationDiscoveryDomain,
		UpdateContext: updateOrganizationDiscoveryDomain,
		DeleteContext: deleteOrganizationDiscoveryDomain,
		Importer: &schema.ResourceImporter{
			StateContext: internalSchema.ImportResourceGroupID("organization_id", "id"),
		},
		Description: "Manage organization discovery domains for Home Realm Discovery. These domains help automatically route users to the correct organization based on their email domain.",
		Schema: map[string]*schema.Schema{
			"organization_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the organization.",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the discovery domain.",
			},
			"domain": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The domain name for organization discovery.",
			},
			"status": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"pending", "verified"}, false),
				Description:  "Verification status. Must be either 'pending' or 'verified'.",
			},
			"verification_txt": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "TXT record value for domain verification.",
			},
			"verification_host": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The full domain where the TXT record should be added.",
			},
			"use_for_organization_discovery": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates whether this domain should be used for organization discovery during login.",
			},
		},
	}
}

func createOrganizationDiscoveryDomain(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	organizationID := data.Get("organization_id").(string)
	discoveryDomain := expandOrganizationDiscoveryDomain(data)

	if err := api.Organization.CreateDiscoveryDomain(ctx, organizationID, discoveryDomain); err != nil {
		return diag.FromErr(err)
	}

	internalSchema.SetResourceGroupID(data, organizationID, discoveryDomain.GetID())

	return readOrganizationDiscoveryDomain(ctx, data, meta)
}

func readOrganizationDiscoveryDomain(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	organizationID, domainID, err := parseOrganizationDiscoveryDomainID(data.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	discoveryDomain, err := api.Organization.DiscoveryDomain(ctx, organizationID, domainID)
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return diag.FromErr(flattenOrganizationDiscoveryDomain(data, discoveryDomain, organizationID))
}

func updateOrganizationDiscoveryDomain(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	organizationID, domainID, err := parseOrganizationDiscoveryDomainID(data.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	discoveryDomain := expandOrganizationDiscoveryDomain(data)

	if err := api.Organization.UpdateDiscoveryDomain(ctx, organizationID, domainID, discoveryDomain); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readOrganizationDiscoveryDomain(ctx, data, meta)
}

func deleteOrganizationDiscoveryDomain(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	organizationID, domainID, err := parseOrganizationDiscoveryDomainID(data.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := api.Organization.DeleteDiscoveryDomain(ctx, organizationID, domainID); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}

// parseOrganizationDiscoveryDomainID parses the resource ID in format "organization_id::id".
func parseOrganizationDiscoveryDomainID(id string) (organizationID, domainID string, err error) {
	parts := strings.Split(id, "::")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid organization discovery domain ID format: %s, expected format: organization_id::id", id)
	}
	return parts[0], parts[1], nil
}
