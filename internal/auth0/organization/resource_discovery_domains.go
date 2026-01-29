package organization

import (
	"context"
	"fmt"

	"github.com/auth0/go-auth0/management"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	"github.com/auth0/terraform-provider-auth0/internal/value"
)

// NewDiscoveryDomainsResource will return a new auth0_organization_discovery_domains (1:many) resource.
func NewDiscoveryDomainsResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"organization_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the organization on which to manage the discovery domains.",
			},
			"discovery_domains": {
				Type: schema.TypeSet,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The domain name for organization discovery.",
						},
						"status": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"pending", "verified"}, false),
							Description:  "Verification status. Must be either 'pending' or 'verified'.",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the discovery domain.",
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
							Computed:    true,
							Description: "Indicates whether this domain should be used for organization discovery during login.",
						},
					},
				},
				Required:    true,
				Description: "Discovery domains that are configured for the organization.",
			},
		},
		CreateContext: createOrganizationDiscoveryDomains,
		ReadContext:   readOrganizationDiscoveryDomains,
		UpdateContext: updateOrganizationDiscoveryDomains,
		DeleteContext: deleteOrganizationDiscoveryDomains,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can manage discovery domains on an organization.",
	}
}

func createOrganizationDiscoveryDomains(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	organizationID := data.Get("organization_id").(string)

	var alreadyEnabledDomains []*management.OrganizationDiscoveryDomain
	var checkpoint string
	for {
		var opts []management.RequestOption
		if checkpoint != "" {
			opts = append(opts, management.From(checkpoint))
		}
		opts = append(opts, management.Take(100))

		domainList, err := api.Organization.DiscoveryDomains(
			ctx,
			organizationID,
			opts...,
		)
		if err != nil {
			return diag.FromErr(internalError.HandleAPIError(data, err))
		}

		alreadyEnabledDomains = append(alreadyEnabledDomains, domainList.Domains...)

		if !domainList.HasNext() {
			break
		}

		checkpoint = domainList.Next
	}

	data.SetId(organizationID)

	domainsToAdd := expandOrganizationDiscoveryDomains(data.GetRawConfig().GetAttr("discovery_domains"))

	if diagnostics := guardAgainstErasingUnwantedDiscoveryDomains(
		organizationID,
		alreadyEnabledDomains,
		domainsToAdd,
	); diagnostics.HasError() {
		data.SetId("")
		return diagnostics
	}

	if len(domainsToAdd) > len(alreadyEnabledDomains) {
		var result *multierror.Error

		for _, domain := range domainsToAdd {
			err := api.Organization.CreateDiscoveryDomain(ctx, organizationID, domain)
			result = multierror.Append(result, err)
		}

		if result.ErrorOrNil() != nil {
			return diag.FromErr(result.ErrorOrNil())
		}
	}

	return readOrganizationDiscoveryDomains(ctx, data, meta)
}

func readOrganizationDiscoveryDomains(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	var domains []*management.OrganizationDiscoveryDomain
	var checkpoint string
	for {
		var opts []management.RequestOption
		if checkpoint != "" {
			opts = append(opts, management.From(checkpoint))
		}
		opts = append(opts, management.Take(100))

		domainList, err := api.Organization.DiscoveryDomains(
			ctx,
			data.Id(),
			opts...,
		)
		if err != nil {
			return diag.FromErr(internalError.HandleAPIError(data, err))
		}

		domains = append(domains, domainList.Domains...)

		if !domainList.HasNext() {
			break
		}

		checkpoint = domainList.Next
	}

	result := multierror.Append(
		data.Set("organization_id", data.Id()),
		flattenOrganizationDiscoveryDomains(data, domains),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updateOrganizationDiscoveryDomains(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	organizationID := data.Id()

	// We are using the data from expandOrganizationDiscoveryDomains and not value.Difference to preserve the nilness of the data.
	domains := expandOrganizationDiscoveryDomains(data.GetRawConfig().GetAttr("discovery_domains"))
	domainMap := make(map[string]*management.OrganizationDiscoveryDomain)
	for _, domain := range domains {
		domainMap[domain.GetDomain()] = domain
	}

	toAdd, toRemove := value.Difference(data, "discovery_domains")
	var result *multierror.Error

	for _, rmDomain := range toRemove {
		domain := rmDomain.(map[string]interface{})
		domainID := domain["id"].(string)

		err := api.Organization.DeleteDiscoveryDomain(ctx, organizationID, domainID)
		if internalError.IsStatusNotFound(err) {
			err = nil
		}

		result = multierror.Append(result, err)
	}

	for _, addDomain := range toAdd {
		domain := addDomain.(map[string]interface{})
		domainName := domain["domain"].(string)

		err := api.Organization.CreateDiscoveryDomain(ctx, organizationID, domainMap[domainName])
		result = multierror.Append(result, err)
	}

	if result.ErrorOrNil() != nil {
		return diag.FromErr(result.ErrorOrNil())
	}

	return readOrganizationDiscoveryDomains(ctx, data, meta)
}

func deleteOrganizationDiscoveryDomains(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	// Fetch all current discovery domains to get their IDs.
	var domains []*management.OrganizationDiscoveryDomain
	var checkpoint string
	for {
		var opts []management.RequestOption
		if checkpoint != "" {
			opts = append(opts, management.From(checkpoint))
		}
		opts = append(opts, management.Take(100))

		domainList, err := api.Organization.DiscoveryDomains(
			ctx,
			data.Id(),
			opts...,
		)
		if err != nil {
			return diag.FromErr(internalError.HandleAPIError(data, err))
		}

		domains = append(domains, domainList.Domains...)

		if !domainList.HasNext() {
			break
		}

		checkpoint = domainList.Next
	}

	var result *multierror.Error

	for _, domain := range domains {
		err := api.Organization.DeleteDiscoveryDomain(ctx, data.Id(), domain.GetID())
		if internalError.IsStatusNotFound(err) {
			err = nil
		}

		result = multierror.Append(result, err)
	}

	return diag.FromErr(result.ErrorOrNil())
}

func guardAgainstErasingUnwantedDiscoveryDomains(
	organizationID string,
	alreadyEnabled []*management.OrganizationDiscoveryDomain,
	desired []*management.OrganizationDiscoveryDomain,
) diag.Diagnostics {
	if len(alreadyEnabled) == 0 {
		return nil
	}

	alreadyEnabledDomains := make([]string, 0)
	for _, domain := range alreadyEnabled {
		alreadyEnabledDomains = append(alreadyEnabledDomains, domain.GetDomain())
	}

	desiredDomains := make([]string, 0)
	for _, domain := range desired {
		desiredDomains = append(desiredDomains, domain.GetDomain())
	}

	if !cmp.Equal(alreadyEnabledDomains, desiredDomains) {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Organization with non empty enabled discovery domains",
				Detail: fmt.Sprintf("Detected an organization (%s) with already enabled discovery domains. Import the "+
					"auth0_organization_discovery_domains resource instead of creating it.", organizationID),
			},
		}
	}

	return nil
}
