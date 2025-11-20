package networkacl

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// NewResource will return a new auth0_network_acl resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createNetworkACL,
		ReadContext:   readNetworkACL,
		UpdateContext: updateNetworkACL,
		DeleteContext: deleteNetworkACL,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can create and manage NetworkACLs for a tenant.",
		Schema: map[string]*schema.Schema{
			"description": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The description of the Network ACL",
			},
			"active": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Whether the Network ACL is active",
			},
			"priority": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The priority of the Network ACL. Must be unique between 1 and 10.",
			},
			"rule": networkACLRuleSchema,
		},
	}
}

var networkACLRuleSchema = &schema.Schema{
	Type:        schema.TypeList,
	Required:    true,
	MaxItems:    1,
	Description: "The rule of the Network ACL",
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"action": {
				Type:        networkACLRuleActionSchema.Type,
				Required:    true,
				MaxItems:    networkACLRuleActionSchema.MaxItems,
				Description: networkACLRuleActionSchema.Description,
				Elem:        networkACLRuleActionSchema.Elem,
			},
			"match": {
				Type:         networkACLRuleMatchSchema.Type,
				Optional:     true,
				MaxItems:     networkACLRuleMatchSchema.MaxItems,
				Description:  networkACLRuleMatchSchema.Description,
				Elem:         networkACLRuleMatchSchema.Elem,
				AtLeastOneOf: []string{"rule.0.match", "rule.0.not_match"},
			},
			"not_match": {
				Type:         networkACLRuleMatchSchema.Type,
				Optional:     true,
				MaxItems:     networkACLRuleMatchSchema.MaxItems,
				Description:  networkACLRuleMatchSchema.Description,
				Elem:         networkACLRuleMatchSchema.Elem,
				AtLeastOneOf: []string{"rule.0.match", "rule.0.not_match"},
			},
			"scope": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The scope of the Network ACL Rule",
				ValidateFunc: validation.StringInSlice([]string{
					"management",
					"authentication",
					"tenant",
					"dynamic_client_registration",
				}, false),
			},
		},
	},
}

var networkACLRuleMatchSchema = &schema.Schema{
	Type:        schema.TypeList,
	Optional:    true,
	MaxItems:    1,
	Description: "The configuration for the Network ACL Rule",
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"asns": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
				Description: "ASNs. Must contain between 1 and 10 unique items.",
			},
			"geo_country_codes": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Geo Country Codes. Must contain between 1 and 10 unique items.",
			},
			"geo_subdivision_codes": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Geo Subdivision Codes. Must contain between 1 and 10 unique items.",
			},
			"ipv4_cidrs": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "IPv4 CIDRs. Must contain between 1 and 10 unique items. Can be IPv4 addresses or CIDR blocks.",
			},
			"ipv6_cidrs": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "IPv6 CIDRs. Must contain between 1 and 10 unique items. Can be IPv6 addresses or CIDR blocks.",
			},
			"ja3_fingerprints": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "JA3 Fingerprints. Must contain between 1 and 10 unique items.",
			},
			"ja4_fingerprints": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "JA4 Fingerprints. Must contain between 1 and 10 unique items.",
			},
			"user_agents": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "User Agents. Must contain between 1 and 10 unique items.",
			},
		},
	},
}

var networkACLRuleActionSchema = &schema.Schema{
	Type:        schema.TypeList,
	Required:    true,
	MaxItems:    1,
	Description: "The action configuration for the Network ACL Rule. Only one action type (block, allow, log, or redirect) should be specified.",
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"block": {
				Type:         schema.TypeBool,
				Optional:     true,
				Description:  "If true, blocks the request. When using block action, no other properties should be set.",
				AtLeastOneOf: []string{"rule.0.action.0.block", "rule.0.action.0.allow", "rule.0.action.0.log", "rule.0.action.0.redirect"},
			},
			"allow": {
				Type:         schema.TypeBool,
				Optional:     true,
				Description:  "If true, allows the request. When using allow action, no other properties should be set.",
				AtLeastOneOf: []string{"rule.0.action.0.block", "rule.0.action.0.allow", "rule.0.action.0.log", "rule.0.action.0.redirect"},
			},
			"log": {
				Type:         schema.TypeBool,
				Optional:     true,
				Description:  "If true, logs the request. When using log action, no other properties should be set.",
				AtLeastOneOf: []string{"rule.0.action.0.block", "rule.0.action.0.allow", "rule.0.action.0.log", "rule.0.action.0.redirect"},
			},
			"redirect": {
				Type:         schema.TypeBool,
				Optional:     true,
				Description:  "If true, redirects the request. When using redirect action, redirect_uri must also be specified.",
				AtLeastOneOf: []string{"rule.0.action.0.block", "rule.0.action.0.allow", "rule.0.action.0.log", "rule.0.action.0.redirect"},
			},
			"redirect_uri": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The URI to redirect to when redirect is true. Required when redirect is true. Must be between 1 and 2000 characters.",
			},
		},
	},
}

func createNetworkACL(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	networkACL, err := expandNetworkACL(data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := api.NetworkACL.Create(ctx, networkACL); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(networkACL.GetID())

	return readNetworkACL(ctx, data, meta)
}

func readNetworkACL(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	networkACL, err := api.NetworkACL.Read(ctx, data.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return diag.FromErr(flattenNetworkACL(data, networkACL))
}

func updateNetworkACL(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	networkACL, err := expandNetworkACL(data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := api.NetworkACL.Update(ctx, data.Id(), networkACL); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readNetworkACL(ctx, data, meta)
}

func deleteNetworkACL(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	if err := api.NetworkACL.Delete(ctx, data.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
