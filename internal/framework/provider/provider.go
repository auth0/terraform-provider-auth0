package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	"github.com/auth0/terraform-provider-auth0/internal/framework/auth0/encryptionkeymanager"
	"github.com/auth0/terraform-provider-auth0/internal/framework/auth0/resourceserver"
)

// Auth0Provider is the type we use for implementing a framework provider.
type Auth0Provider struct {
	configureFunc func(context.Context, provider.ConfigureRequest, *provider.ConfigureResponse)
}

// Metadata will be called by the framework to get the type name and version for the auth0 provider.
func (p *Auth0Provider) Metadata(_ context.Context, _ provider.MetadataRequest, response *provider.MetadataResponse) {
	response.TypeName = "auth0"
	response.Version = config.GetProviderVersion()
}

// Schema will be called by the framework to get the schema name for the auth0 provider.
func (p *Auth0Provider) Schema(_ context.Context, _ provider.SchemaRequest, response *provider.SchemaResponse) {
	if response != nil {
		response.Schema = schema.Schema{
			Attributes: map[string]schema.Attribute{
				"domain": schema.StringAttribute{
					Optional: true,
					MarkdownDescription: "Your Auth0 domain name. " +
						"It can also be sourced from the `AUTH0_DOMAIN` environment variable.",
				},
				"audience": schema.StringAttribute{
					Optional: true,
					MarkdownDescription: "Your Auth0 audience when using a custom domain. " +
						"It can also be sourced from the `AUTH0_AUDIENCE` environment variable.",
				},
				"client_id": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(path.Expressions{
							path.MatchRoot("api_token"),
						}...),
						stringvalidator.AlsoRequires(path.Expressions{
							path.MatchRoot("client_secret"),
						}...),
					},
					MarkdownDescription: "Your Auth0 client ID. " +
						"It can also be sourced from the `AUTH0_CLIENT_ID` environment variable.",
				},
				"client_secret": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(path.Expressions{
							path.MatchRoot("api_token"),
						}...),
						stringvalidator.AlsoRequires(path.Expressions{
							path.MatchRoot("client_id"),
						}...),
					},
					MarkdownDescription: "Your Auth0 client secret. " +
						"It can also be sourced from the `AUTH0_CLIENT_SECRET` environment variable.",
				},
				"api_token": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(path.Expressions{
							path.MatchRoot("client_id"),
						}...),
						stringvalidator.ConflictsWith(path.Expressions{
							path.MatchRoot("client_secret"),
						}...),
					},
					MarkdownDescription: "Your Auth0 [management api access token]" +
						"(https://auth0.com/docs/security/tokens/access-tokens/management-api-access-tokens). " +
						"It can also be sourced from the `AUTH0_API_TOKEN` environment variable. " +
						"It can be used instead of `client_id` + `client_secret`. " +
						"If both are specified, `api_token` will be used over `client_id` + `client_secret` fields.",
				},
				"debug": schema.BoolAttribute{
					Optional:            true,
					MarkdownDescription: "Indicates whether to turn on debug mode.",
				},
			},
		}
	}
}

// SetConfigureFunc is used by our testing code to change the Configure func in the auth0 provider.
func (p *Auth0Provider) SetConfigureFunc(cfgFunc func(context.Context, provider.ConfigureRequest, *provider.ConfigureResponse)) {
	p.configureFunc = cfgFunc
}

// Configure will be called by the framework to configure the auth0 provider.
func (p *Auth0Provider) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
	p.configureFunc(ctx, request, response)
}

// DataSources will be called by the framework to configure the auth0 provider data sources.
func (p *Auth0Provider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDataSource,
		resourceserver.NewDataSource,
	}
}

// Resources will be called by the framework to configure the auth0 provider resources.
func (p *Auth0Provider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		encryptionkeymanager.NewResource,
		resourceserver.NewResource,
		resourceserver.NewScopeResource,
		resourceserver.NewScopesResource,
	}
}

// New returns a terraform Framework provider.Provider.
func New() *Auth0Provider {
	return &Auth0Provider{
		configureFunc: config.ConfigureFrameworkProvider(),
	}
}
