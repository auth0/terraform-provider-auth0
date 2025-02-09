package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"

	"github.com/auth0/terraform-provider-auth0/internal/config"
)

type providerDataSource struct {
	cfg *config.Config
}

// NewDataSource will return a new auth0_provider data source.
func NewDataSource() datasource.DataSource {
	return &providerDataSource{}
}

// Configure will be called by the framework to configure the auth0_provider data source.
func (r *providerDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if request.ProviderData != nil {
		r.cfg = request.ProviderData.(*config.Config)
	}
}

// Metadata will be called by the framework to get the type name for the auth0_encryption_key_manager datasource.
func (r *providerDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "auth0_provider"
}

// Schema will be called by the framework to get the schema name for the auth0_encryption_key_manager datasource.
func (r *providerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	if response != nil {
		response.Schema = schema.Schema{
			Description:         "A data source for retrieving basic information about the provider.",
			MarkdownDescription: "A data source for retrieving basic information about the provider.",
			Attributes: map[string]schema.Attribute{
				"provider_version": schema.StringAttribute{
					Computed:            true,
					Description:         "The version of the provider. ",
					MarkdownDescription: "The version of the provider. ",
				},
			},
		}
	}
}

// Read will be called by the framework to read an auth0_encryption_key_manager data source.
func (r *providerDataSource) Read(ctx context.Context, _ datasource.ReadRequest, response *datasource.ReadResponse) {
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("provider_version"), config.GetProviderVersion())...)
}
