package user

import (
	"context"
	"errors"

	managementv3 "github.com/auth0/go-auth0/v3/management"
	managementv3client "github.com/auth0/go-auth0/v3/management/client"
	"github.com/auth0/go-auth0/v3/management/core"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	"github.com/auth0/terraform-provider-auth0/internal/value"
)

// NewConnectedAccountsDataSource will return a new auth0_user_connected_accounts data source.
func NewConnectedAccountsDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readUserConnectedAccountsForDataSource,
		Description: "Data source to retrieve all connected accounts for a specific Auth0 user.",
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the user.",
			},
			"connected_accounts": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of connected accounts for the user.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The unique identifier for the connected account.",
						},
						"connection": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the connection associated with the account.",
						},
						"connection_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The unique identifier of the connection associated with the account.",
						},
						"strategy": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The authentication strategy used by the connection.",
						},
						"access_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The access type for the connected account.",
						},
						"scopes": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The scopes granted for this connected account.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"created_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ISO 8601 timestamp when the connected account was created.",
						},
						"expires_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ISO 8601 timestamp when the connected account expires. Empty string if not set.",
						},
						"organization_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The identifier of the organization associated with the connected account. Empty string if not set.",
						},
					},
				},
			},
		},
	}
}

func readUserConnectedAccountsForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	userID := data.Get("user_id").(string)

	accounts, err := fetchAllConnectedAccounts(ctx, apiv3, userID)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(userID)

	return diag.FromErr(data.Set("connected_accounts", flattenConnectedAccounts(accounts)))
}

func fetchAllConnectedAccounts(ctx context.Context, apiv3 *managementv3client.Management, userID string) ([]*managementv3.ConnectedAccount, error) {
	var accounts []*managementv3.ConnectedAccount

	page, err := apiv3.Users.ConnectedAccounts.List(ctx, userID, &managementv3.GetUserConnectedAccountsRequestParameters{})
	if err != nil {
		return nil, err
	}
	accounts = append(accounts, page.Results...)

	for !page.RawResponse.Done {
		page, err = page.GetNextPage(ctx)
		if page == nil && err == nil {
			break
		}
		if err != nil {
			if errors.Is(err, core.ErrNoPages) {
				break
			}
			return nil, err
		}
		accounts = append(accounts, page.Results...)
	}

	return accounts, nil
}

func flattenConnectedAccounts(accounts []*managementv3.ConnectedAccount) []interface{} {
	result := make([]interface{}, 0, len(accounts))

	for _, a := range accounts {
		orgID := ""
		if a.OrganizationID != nil {
			orgID = *a.OrganizationID
		}

		result = append(result, map[string]interface{}{
			"id":              a.GetID(),
			"connection":      a.GetConnection(),
			"connection_id":   a.GetConnectionID(),
			"strategy":        a.GetStrategy(),
			"access_type":     string(a.GetAccessType()),
			"scopes":          a.GetScopes(),
			"created_at":      value.FormatTime(a.GetCreatedAt()),
			"expires_at":      value.FormatTime(a.GetExpiresAt()),
			"organization_id": orgID,
		})
	}

	return result
}
