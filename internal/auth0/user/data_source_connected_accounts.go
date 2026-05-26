package user

import (
	"context"
	"errors"
	"time"

	managementv2 "github.com/auth0/go-auth0/v2/management"
	managementv2client "github.com/auth0/go-auth0/v2/management/client"
	"github.com/auth0/go-auth0/v2/management/core"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
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

// Uses GetAPIV2 — v2 SDK is the permanent choice for connected accounts.
// The ConnectedAccount struct is Fern-generated from the v2 spec; no migration to v1 is planned.
func readUserConnectedAccountsForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Uses GetAPIV2 — v2 SDK is the permanent choice for connected accounts.
	// The ConnectedAccount struct is Fern-generated from the v2 spec; no migration to v1 is planned.
	apiv2 := meta.(*config.Config).GetAPIV2()

	userID := data.Get("user_id").(string)

	accounts, err := fetchAllConnectedAccounts(ctx, apiv2, userID)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(userID)

	return diag.FromErr(data.Set("connected_accounts", flattenConnectedAccounts(accounts)))
}

func fetchAllConnectedAccounts(ctx context.Context, apiv2 *managementv2client.Management, userID string) ([]*managementv2.ConnectedAccount, error) {
	var accounts []*managementv2.ConnectedAccount

	page, err := apiv2.Users.ConnectedAccounts.List(ctx, userID, &managementv2.GetUserConnectedAccountsRequestParameters{})
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

func flattenConnectedAccounts(accounts []*managementv2.ConnectedAccount) []interface{} {
	result := make([]interface{}, 0, len(accounts))

	for _, a := range accounts {
		expiresAt := ""
		if a.ExpiresAt != nil {
			expiresAt = a.ExpiresAt.UTC().Format(time.RFC3339)
		}

		orgID := ""
		if a.OrganizationID != nil {
			orgID = *a.OrganizationID
		}

		scopes := a.GetScopes()

		createdAt := ""
		if t := a.GetCreatedAt(); !t.IsZero() {
			createdAt = t.UTC().Format(time.RFC3339)
		}

		result = append(result, map[string]interface{}{
			"id":              a.GetID(),
			"connection":      a.GetConnection(),
			"connection_id":   a.GetConnectionID(),
			"strategy":        a.GetStrategy(),
			"access_type":     string(a.GetAccessType()),
			"scopes":          scopes,
			"created_at":      createdAt,
			"expires_at":      expiresAt,
			"organization_id": orgID,
		})
	}

	return result
}
