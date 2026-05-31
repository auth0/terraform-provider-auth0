// data_source_plural.go — implements the auth0_clients (plural) data source.
//
// Returns a list of clients matching optional filters. Each entry is a SUBSET
// of the per-client schema (the most useful identifying fields), to keep state
// payloads manageable. Users who need all fields for a specific client should
// chain through `data.auth0_client` with the client_id from the list.
package auth0client

import (
	"context"

	mgmt "github.com/auth0/go-auth0/v2/management"
	mgmtclient "github.com/auth0/go-auth0/v2/management/client"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/auth0/terraform-provider-auth0/v2/internal/framework"
)

var (
	_ datasource.DataSource              = (*clientsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*clientsDataSource)(nil)
)

// NewListDataSource returns a fresh auth0_clients data source implementation.
func NewListDataSource() datasource.DataSource { return &clientsDataSource{} }

type clientsDataSource struct {
	mgmt *mgmtclient.Management
}

// clientsDSModel is the data-source view: filters + a flat list of summary
// objects.
type clientsDSModel struct {
	ID types.String `tfsdk:"id"`

	// Filters (all optional).
	AppType          types.String `tfsdk:"app_type"`
	IsFirstParty     types.Bool   `tfsdk:"is_first_party"`
	IsGlobal         types.Bool   `tfsdk:"is_global"`
	ExternalClientID types.String `tfsdk:"external_client_id"`
	Q                types.String `tfsdk:"q"`

	// Result.
	Clients types.List `tfsdk:"clients"`
}

// clientSummaryAttrTypes is the attribute schema for one entry in `clients`.
func clientSummaryAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"client_id":      types.StringType,
		"name":           types.StringType,
		"description":    types.StringType,
		"app_type":       types.StringType,
		"is_first_party": types.BoolType,
		"global":         types.BoolType,
		"logo_uri":       types.StringType,
		"tenant":         types.StringType,
	}
}

func (d *clientsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_clients"
}

func (d *clientsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if m, ok := framework.ManagementFromDataSource(req, resp); ok {
		d.mgmt = m
	}
}

func (d *clientsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "List Auth0 clients (applications) matching optional filters. " +
			"Each entry is a compact summary; use `data.auth0_client` to fetch the full record for a specific client.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Computed:    true,
				Description: "Synthetic identifier (always `clients`).",
			},
			"app_type": dsschema.StringAttribute{
				Optional:    true,
				Description: "Optional comma-separated list of application types to filter by (e.g. `regular_web,spa`).",
			},
			"is_first_party": dsschema.BoolAttribute{
				Optional:    true,
				Description: "Optional filter: only return first-party clients (true) or third-party (false).",
			},
			"is_global": dsschema.BoolAttribute{
				Optional:    true,
				Description: "Optional filter on the legacy `global` flag.",
			},
			"external_client_id": dsschema.StringAttribute{
				Optional:    true,
				Description: "Optional filter by Client ID Metadata Document URI for CIMD-registered clients.",
			},
			"q": dsschema.StringAttribute{
				Optional:    true,
				Description: "Optional advanced Lucene query (see Auth0 docs for permitted forms).",
			},
			"clients": dsschema.ListNestedAttribute{
				Computed:    true,
				Description: "Matching clients.",
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: map[string]dsschema.Attribute{
						"client_id":      dsschema.StringAttribute{Computed: true},
						"name":           dsschema.StringAttribute{Computed: true},
						"description":    dsschema.StringAttribute{Computed: true},
						"app_type":       dsschema.StringAttribute{Computed: true},
						"is_first_party": dsschema.BoolAttribute{Computed: true},
						"global":         dsschema.BoolAttribute{Computed: true},
						"logo_uri":       dsschema.StringAttribute{Computed: true},
						"tenant":         dsschema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *clientsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg clientsDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := &mgmt.ListClientsRequestParameters{}
	if !cfg.AppType.IsNull() && !cfg.AppType.IsUnknown() && cfg.AppType.ValueString() != "" {
		v := cfg.AppType.ValueString()
		params.AppType = &v
	}
	if !cfg.IsFirstParty.IsNull() && !cfg.IsFirstParty.IsUnknown() {
		v := cfg.IsFirstParty.ValueBool()
		params.IsFirstParty = &v
	}
	if !cfg.IsGlobal.IsNull() && !cfg.IsGlobal.IsUnknown() {
		v := cfg.IsGlobal.ValueBool()
		params.IsGlobal = &v
	}
	if !cfg.ExternalClientID.IsNull() && !cfg.ExternalClientID.IsUnknown() && cfg.ExternalClientID.ValueString() != "" {
		v := cfg.ExternalClientID.ValueString()
		params.ExternalClientID = &v
	}
	if !cfg.Q.IsNull() && !cfg.Q.IsUnknown() && cfg.Q.ValueString() != "" {
		v := cfg.Q.ValueString()
		params.Q = &v
	}

	page, err := d.mgmt.Clients.List(ctx, params)
	if err != nil {
		framework.AddAPIError(&resp.Diagnostics, "Failed to list clients", err)
		return
	}

	var entries []attr.Value
	iter := page.Iterator()
	for iter.Next(ctx) {
		c := iter.Current()
		obj, d := clientSummaryToObject(c)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		entries = append(entries, obj)
	}
	if err := iter.Err(); err != nil {
		framework.AddAPIError(&resp.Diagnostics, "Failed to iterate clients", err)
		return
	}

	listVal, dl := types.ListValue(types.ObjectType{AttrTypes: clientSummaryAttrTypes()}, entries)
	resp.Diagnostics.Append(dl...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := cfg
	state.ID = types.StringValue("clients")
	state.Clients = listVal
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// clientSummaryToObject projects a paginated *mgmt.Client into the
// data-source's summary object shape.
func clientSummaryToObject(c *mgmt.Client) (types.Object, diag.Diagnostics) {
	strOrNull := func(s *string) attr.Value {
		if s == nil {
			return types.StringNull()
		}
		return types.StringValue(*s)
	}
	enumOrNull := func(e *mgmt.ClientAppTypeEnum) attr.Value {
		if e == nil {
			return types.StringNull()
		}
		return types.StringValue(string(*e))
	}
	vals := map[string]attr.Value{
		"client_id":      strOrNull(c.ClientID),
		"name":           strOrNull(c.Name),
		"description":    strOrNull(c.Description),
		"app_type":       enumOrNull(c.AppType),
		"is_first_party": types.BoolPointerValue(c.IsFirstParty),
		"global":         types.BoolPointerValue(c.Global),
		"logo_uri":       strOrNull(c.LogoURI),
		"tenant":         strOrNull(c.Tenant),
	}
	return types.ObjectValue(clientSummaryAttrTypes(), vals)
}
