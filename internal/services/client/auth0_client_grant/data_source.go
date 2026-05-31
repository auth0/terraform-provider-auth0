// data_source.go — implements the auth0_client_grant data source.
//
// Lookup options:
//   - by `id` (highest precedence; uses `GET /client-grants/{id}`)
//   - by `client_id` + `audience` (uses `GET /client-grants` with filters; we
//     expect exactly one match, since (client_id, audience) is unique per
//     Auth0's data model)
//
// Either form populates the same set of computed attributes.
package auth0clientgrant

import (
	"context"

	mgmt "github.com/auth0/go-auth0/v2/management"
	mgmtclient "github.com/auth0/go-auth0/v2/management/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/auth0/terraform-provider-auth0/v2/internal/framework"
)

var (
	_ datasource.DataSource              = (*clientGrantDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*clientGrantDataSource)(nil)
)

// NewDataSource returns a fresh auth0_client_grant data source implementation.
func NewDataSource() datasource.DataSource { return &clientGrantDataSource{} }

type clientGrantDataSource struct {
	mgmt *mgmtclient.Management
}

func (d *clientGrantDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_client_grant"
}

func (d *clientGrantDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if m, ok := framework.ManagementFromDataSource(req, resp); ok {
		d.mgmt = m
	}
}

func (d *clientGrantDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Look up an Auth0 client grant — either by its `id`, or by the (`client_id`, `audience`) pair which uniquely identifies a grant.",
		Attributes: map[string]dsschema.Attribute{
			// Lookup keys (all optional, validated in Read).
			"id":        dsschema.StringAttribute{Optional: true, Computed: true, Description: "Auth0 client grant identifier. If set, takes precedence over `client_id` + `audience`."},
			"client_id": dsschema.StringAttribute{Optional: true, Computed: true, Description: "Client ID that owns this grant. Use together with `audience` for lookup."},
			"audience":  dsschema.StringAttribute{Optional: true, Computed: true, Description: "Audience (API identifier). Use together with `client_id` for lookup."},

			// Computed attributes.
			"scopes":                      dsschema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Scopes allowed for this grant."},
			"organization_usage":          dsschema.StringAttribute{Computed: true, Description: "Organization usage policy."},
			"allow_any_organization":      dsschema.BoolAttribute{Computed: true, Description: "Whether any organization may use this grant."},
			"default_for":                 dsschema.StringAttribute{Computed: true, Description: "Default-for category, e.g. `third_party_clients`."},
			"is_system":                   dsschema.BoolAttribute{Computed: true, Description: "Whether this grant is Auth0-managed and read-only."},
			"subject_type":                dsschema.StringAttribute{Computed: true, Description: "Subject type (`client` or `user`)."},
			"authorization_details_types": dsschema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Authorization details types (RFC 9396)."},
			"allow_all_scopes":            dsschema.BoolAttribute{Computed: true, Description: "Whether all resource-server scopes are allowed for this grant."},
		},
	}
}

func (d *clientGrantDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg clientGrantResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !cfg.ID.IsNull() && !cfg.ID.IsUnknown() && cfg.ID.ValueString() != ""
	hasClientID := !cfg.ClientID.IsNull() && !cfg.ClientID.IsUnknown() && cfg.ClientID.ValueString() != ""
	hasAudience := !cfg.Audience.IsNull() && !cfg.Audience.IsUnknown() && cfg.Audience.ValueString() != ""

	if !hasID && (!hasClientID || !hasAudience) {
		resp.Diagnostics.AddError(
			"Missing lookup parameters",
			"Set either `id`, or both `client_id` and `audience`, to look up a client grant.",
		)
		return
	}

	state := clientGrantResourceModel{}

	if hasID {
		got, err := d.mgmt.ClientGrants.Get(ctx, cfg.ID.ValueString())
		if err != nil {
			framework.AddAPIError(&resp.Diagnostics, "Failed to read client grant", err)
			return
		}
		flattenGet(&state, got)
	} else {
		// Look up via List with filters; expect exactly one match.
		params := &mgmt.ListClientGrantsRequestParameters{}
		cid := cfg.ClientID.ValueString()
		aud := cfg.Audience.ValueString()
		params.ClientID = &cid
		params.Audience = &aud

		page, err := d.mgmt.ClientGrants.List(ctx, params)
		if err != nil {
			framework.AddAPIError(&resp.Diagnostics, "Failed to list client grants", err)
			return
		}

		var matches []*mgmt.ClientGrantResponseContent
		iter := page.Iterator()
		for iter.Next(ctx) {
			matches = append(matches, iter.Current())
		}
		if err := iter.Err(); err != nil {
			framework.AddAPIError(&resp.Diagnostics, "Failed to iterate client grants", err)
			return
		}

		switch len(matches) {
		case 0:
			resp.Diagnostics.AddError(
				"Client grant not found",
				"No client grant found for the given `client_id` and `audience`.",
			)
			return
		case 1:
			flattenList(&state, matches[0])
		default:
			resp.Diagnostics.AddError(
				"Multiple client grants found",
				"More than one client grant matched the given `client_id` and `audience`. This is unexpected — try looking up by `id`.",
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// flattenList adapts the (paginated) ClientGrantResponseContent into the
// resource model. It maps to the same fields as flattenGet but the SDK type
// is distinct.
func flattenList(m *clientGrantResourceModel, c *mgmt.ClientGrantResponseContent) {
	flattenInto(m, commonClientGrantFields{
		ID:                        c.ID,
		ClientID:                  c.ClientID,
		Audience:                  c.Audience,
		Scope:                     c.Scope,
		OrganizationUsage:         c.OrganizationUsage,
		AllowAnyOrganization:      c.AllowAnyOrganization,
		DefaultFor:                c.DefaultFor,
		IsSystem:                  c.IsSystem,
		SubjectType:               c.SubjectType,
		AuthorizationDetailsTypes: c.AuthorizationDetailsTypes,
		AllowAllScopes:            c.AllowAllScopes,
	})
}
