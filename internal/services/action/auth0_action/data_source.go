package auth0action

import (
	"context"

	mgmtclient "github.com/auth0/go-auth0/v2/management/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/auth0/terraform-provider-auth0/v2/internal/framework"
)

var (
	_ datasource.DataSource              = (*actionDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*actionDataSource)(nil)
)

// NewDataSource returns a fresh auth0_action data source implementation.
func NewDataSource() datasource.DataSource { return &actionDataSource{} }

type actionDataSource struct {
	mgmt *mgmtclient.Management
}

func (d *actionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_action"
}

func (d *actionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if m, ok := framework.ManagementFromDataSource(req, resp); ok {
		d.mgmt = m
	}
}

func (d *actionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves an Auth0 Action by its ID.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Required:    true,
				Description: "The ID of the action to retrieve.",
			},
			"name": dsschema.StringAttribute{
				Computed:    true,
				Description: "The name of the action.",
			},
			"supported_triggers": dsschema.SingleNestedAttribute{
				Computed:    true,
				Description: "The trigger this action targets.",
				Attributes: map[string]dsschema.Attribute{
					"id":      dsschema.StringAttribute{Computed: true, Description: "The trigger identifier."},
					"version": dsschema.StringAttribute{Computed: true, Description: "The trigger version."},
				},
			},
			"code": dsschema.StringAttribute{
				Computed:    true,
				Description: "The source code of the action.",
			},
			"runtime": dsschema.StringAttribute{
				Computed:    true,
				Description: "The Node runtime.",
			},
			"dependencies": dsschema.SetNestedAttribute{
				Computed:    true,
				Description: "The list of third-party npm modules this action depends on.",
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: map[string]dsschema.Attribute{
						"name":    dsschema.StringAttribute{Computed: true, Description: "The npm module name."},
						"version": dsschema.StringAttribute{Computed: true, Description: "The npm module version."},
					},
				},
			},
			"secret_names": dsschema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "The names of the secrets configured on the action (values are never returned).",
			},
			"version_id": dsschema.StringAttribute{
				Computed:    true,
				Description: "The ID of the currently deployed action version.",
			},
		},
	}
}

// dsModel mirrors the data source schema.
type dsModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	SupportedTriggers types.Object `tfsdk:"supported_triggers"`
	Code              types.String `tfsdk:"code"`
	Runtime           types.String `tfsdk:"runtime"`
	Dependencies      types.Set    `tfsdk:"dependencies"`
	SecretNames       types.List   `tfsdk:"secret_names"`
	VersionID         types.String `tfsdk:"version_id"`
}

func (d *actionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	got, err := d.mgmt.Actions.Get(ctx, data.ID.ValueString())
	if err != nil {
		framework.AddAPIError(&resp.Diagnostics, "Failed to read action", err)
		return
	}

	data.Name = types.StringValue(got.GetName())
	data.Code = types.StringValue(got.GetCode())
	data.Runtime = types.StringValue(got.GetRuntime())
	data.SupportedTriggers = flattenTriggers(got.GetSupportedTriggers(), &resp.Diagnostics)
	data.Dependencies = flattenDependencies(got.GetDependencies(), &resp.Diagnostics)
	data.SecretNames = flattenSecretNames(got.GetSecrets(), &resp.Diagnostics)
	if dv := got.DeployedVersion; dv != nil {
		data.VersionID = types.StringValue(dv.GetID())
	} else {
		data.VersionID = types.StringValue("")
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
