package client

import (
	"context"

	"github.com/auth0/terraform-provider-auth0/internal/config"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// NewClientGrantsDataSource will return a new auth0_client_grants data source.
func NewClientGrantsDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceClientGrantsRead,
		Description: "Data source to retrieve a client grants based on client_id and/or audience",
		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of the client to filter by.",
			},
			"audience": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The audience to filter by.",
			},
			"client_grants": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of client grants matching the criteria.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the client grant.",
						},
						"client_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The client ID associated with the grant.",
						},
						"audience": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The audience of the client grant.",
						},
						"scope": {
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
							Description: "List of granted scopes.",
						},
						"subject_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The subject type (usually 'client').",
						},
						"allow_all_scopes": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "When enabled, all scopes configured on the resource server are allowed for this client grant. EA Only.",
						},
					},
				},
			},
		},
	}
}

func dataSourceClientGrantsRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	clientID := data.Get("client_id").(string)
	audience := data.Get("audience").(string)

	var allGrants []*management.ClientGrant
	var from string

	options := []management.RequestOption{
		management.Take(100),
	}

	if clientID != "" {
		options = append(options, management.Parameter("client_id", clientID))
	}
	if audience != "" {
		options = append(options, management.Parameter("audience", audience))
	}

	for {
		if from != "" {
			options = append(options, management.From(from))
		}

		grantList, err := api.ClientGrant.List(ctx, options...)
		if err != nil {
			return diag.FromErr(err)
		}
		allGrants = append(allGrants, grantList.ClientGrants...)

		if !grantList.HasNext() {
			break
		}

		from = grantList.Next
	}

	var flattened []map[string]interface{}

	for _, cg := range allGrants {
		item := map[string]interface{}{
			"id":               cg.GetID(),
			"client_id":        cg.GetClientID(),
			"audience":         cg.GetAudience(),
			"scope":            cg.GetScope(),
			"subject_type":     cg.GetSubjectType(),
			"allow_all_scopes": cg.GetAllowAllScopes(),
		}
		flattened = append(flattened, item)
	}

	if err := data.Set("client_grants", flattened); err != nil {
		return diag.FromErr(err)
	}

	// Setting a synthetic ID here - this data source represents a list.
	data.SetId("client_grants_list")

	return nil
}
