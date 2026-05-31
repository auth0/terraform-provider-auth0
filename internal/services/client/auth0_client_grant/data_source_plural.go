// data_source_plural.go — implements the auth0_client_grants (plural)
// data source.
//
// Returns all client grants matching optional filters: client_id, audience,
// subject_type, allow_any_organization, default_for.
package auth0clientgrant

import (
	"context"

	mgmt "github.com/auth0/go-auth0/v2/management"
	mgmtclient "github.com/auth0/go-auth0/v2/management/client"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/auth0/terraform-provider-auth0/v2/internal/framework"
)

var (
	_ datasource.DataSource              = (*clientGrantsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*clientGrantsDataSource)(nil)
)

// NewListDataSource returns a fresh auth0_client_grants data source implementation.
func NewListDataSource() datasource.DataSource { return &clientGrantsDataSource{} }

type clientGrantsDataSource struct {
	mgmt *mgmtclient.Management
}

// grantsDSModel holds the optional filters and the list result.
type grantsDSModel struct {
	ID types.String `tfsdk:"id"`

	// Filters.
	ClientID             types.String `tfsdk:"client_id"`
	Audience             types.String `tfsdk:"audience"`
	SubjectType          types.String `tfsdk:"subject_type"`
	AllowAnyOrganization types.Bool   `tfsdk:"allow_any_organization"`
	DefaultFor           types.String `tfsdk:"default_for"`

	// Result.
	ClientGrants types.List `tfsdk:"client_grants"`
}

// grantEntryAttrTypes returns the attribute types for one entry in the
// `client_grants` list.
func grantEntryAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                          types.StringType,
		"client_id":                   types.StringType,
		"audience":                    types.StringType,
		"scopes":                      types.ListType{ElemType: types.StringType},
		"organization_usage":          types.StringType,
		"allow_any_organization":      types.BoolType,
		"default_for":                 types.StringType,
		"is_system":                   types.BoolType,
		"subject_type":                types.StringType,
		"authorization_details_types": types.ListType{ElemType: types.StringType},
		"allow_all_scopes":            types.BoolType,
	}
}

func (d *clientGrantsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_client_grants"
}

func (d *clientGrantsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if m, ok := framework.ManagementFromDataSource(req, resp); ok {
		d.mgmt = m
	}
}

func (d *clientGrantsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "List Auth0 client grants matching optional filters.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Computed:    true,
				Description: "Synthetic identifier (always `client_grants`).",
			},

			// Filters.
			"client_id":              dsschema.StringAttribute{Optional: true, Description: "Optional filter on client_id."},
			"audience":               dsschema.StringAttribute{Optional: true, Description: "Optional filter on audience."},
			"subject_type":           dsschema.StringAttribute{Optional: true, Description: "Optional filter on subject_type (`client` or `user`)."},
			"allow_any_organization": dsschema.BoolAttribute{Optional: true, Description: "Optional filter on allow_any_organization."},
			"default_for":            dsschema.StringAttribute{Optional: true, Description: "Optional filter on default_for. Currently only `third_party_clients` is supported by the API."},

			"client_grants": dsschema.ListNestedAttribute{
				Computed:    true,
				Description: "Matching client grants.",
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: map[string]dsschema.Attribute{
						"id":                          dsschema.StringAttribute{Computed: true},
						"client_id":                   dsschema.StringAttribute{Computed: true},
						"audience":                    dsschema.StringAttribute{Computed: true},
						"scopes":                      dsschema.ListAttribute{Computed: true, ElementType: types.StringType},
						"organization_usage":          dsschema.StringAttribute{Computed: true},
						"allow_any_organization":      dsschema.BoolAttribute{Computed: true},
						"default_for":                 dsschema.StringAttribute{Computed: true},
						"is_system":                   dsschema.BoolAttribute{Computed: true},
						"subject_type":                dsschema.StringAttribute{Computed: true},
						"authorization_details_types": dsschema.ListAttribute{Computed: true, ElementType: types.StringType},
						"allow_all_scopes":            dsschema.BoolAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *clientGrantsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg grantsDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := &mgmt.ListClientGrantsRequestParameters{}
	if !cfg.ClientID.IsNull() && !cfg.ClientID.IsUnknown() && cfg.ClientID.ValueString() != "" {
		v := cfg.ClientID.ValueString()
		params.ClientID = &v
	}
	if !cfg.Audience.IsNull() && !cfg.Audience.IsUnknown() && cfg.Audience.ValueString() != "" {
		v := cfg.Audience.ValueString()
		params.Audience = &v
	}
	if !cfg.SubjectType.IsNull() && !cfg.SubjectType.IsUnknown() && cfg.SubjectType.ValueString() != "" {
		enum, err := mgmt.NewClientGrantSubjectTypeEnumFromString(cfg.SubjectType.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("subject_type"), "Invalid subject_type", err.Error())
			return
		}
		params.SubjectType = &enum
	}
	if !cfg.AllowAnyOrganization.IsNull() && !cfg.AllowAnyOrganization.IsUnknown() {
		v := cfg.AllowAnyOrganization.ValueBool()
		// ClientGrantAllowAnyOrganizationEnum is a type alias for bool.
		var b mgmt.ClientGrantAllowAnyOrganizationEnum = v
		params.AllowAnyOrganization = &b
	}
	if !cfg.DefaultFor.IsNull() && !cfg.DefaultFor.IsUnknown() && cfg.DefaultFor.ValueString() != "" {
		enum, err := mgmt.NewClientGrantDefaultForEnumFromString(cfg.DefaultFor.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("default_for"), "Invalid default_for", err.Error())
			return
		}
		params.DefaultFor = &enum
	}

	page, err := d.mgmt.ClientGrants.List(ctx, params)
	if err != nil {
		framework.AddAPIError(&resp.Diagnostics, "Failed to list client grants", err)
		return
	}

	var entries []attr.Value
	iter := page.Iterator()
	for iter.Next(ctx) {
		c := iter.Current()
		obj, dd := grantEntryToObject(c)
		resp.Diagnostics.Append(dd...)
		if resp.Diagnostics.HasError() {
			return
		}
		entries = append(entries, obj)
	}
	if err := iter.Err(); err != nil {
		framework.AddAPIError(&resp.Diagnostics, "Failed to iterate client grants", err)
		return
	}

	listVal, dl := types.ListValue(types.ObjectType{AttrTypes: grantEntryAttrTypes()}, entries)
	resp.Diagnostics.Append(dl...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := cfg
	state.ID = types.StringValue("client_grants")
	state.ClientGrants = listVal
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// grantEntryToObject projects a *mgmt.ClientGrantResponseContent into the
// data-source's list-entry shape.
func grantEntryToObject(c *mgmt.ClientGrantResponseContent) (types.Object, diag.Diagnostics) {
	strOrNull := func(s *string) attr.Value {
		if s == nil {
			return types.StringNull()
		}
		return types.StringValue(*s)
	}

	orgUsage := types.StringNull()
	if c.OrganizationUsage != nil {
		orgUsage = types.StringValue(string(*c.OrganizationUsage))
	}
	defaultFor := types.StringNull()
	if c.DefaultFor != nil {
		defaultFor = types.StringValue(string(*c.DefaultFor))
	}
	subjectType := types.StringNull()
	if c.SubjectType != nil {
		subjectType = types.StringValue(string(*c.SubjectType))
	}

	vals := map[string]attr.Value{
		"id":                          strOrNull(c.ID),
		"client_id":                   strOrNull(c.ClientID),
		"audience":                    strOrNull(c.Audience),
		"scopes":                      framework.StringSliceToList(c.Scope),
		"organization_usage":          orgUsage,
		"allow_any_organization":      types.BoolPointerValue(c.AllowAnyOrganization),
		"default_for":                 defaultFor,
		"is_system":                   types.BoolPointerValue(c.IsSystem),
		"subject_type":                subjectType,
		"authorization_details_types": framework.StringSliceToList(c.AuthorizationDetailsTypes),
		"allow_all_scopes":            types.BoolPointerValue(c.AllowAllScopes),
	}
	return types.ObjectValue(grantEntryAttrTypes(), vals)
}
