package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"

	"github.com/auth0/terraform-provider-auth0/internal/config"
)

type auth0Provider struct {
}

type auth0ProviderModel struct {
	Domain types.String `tfsdk:"domain"`
	Audience types.String `tfsdk:"audience"`
	ClientID types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	ApiToken types.String `tfsdk:"api_token"`
	Debug types.Bool `tfsdk:"debug"`
}

func (p *auth0Provider) Metadata(_ context.Context, _ provider.MetadataRequest, _ *provider.MetadataResponse) {
}

func (p *auth0Provider) Schema(_ context.Context, _ provider.SchemaRequest, response *provider.SchemaResponse) {
	if response != nil {
		response.Schema = schema.Schema{
			Attributes: map[string]schema.Attribute{
				"domain": schema.StringAttribute{
					Optional: true,
					Description: "Your Auth0 domain name. " +
						"It can also be sourced from the AUTH0_DOMAIN environment variable.",
					MarkdownDescription: "Your Auth0 domain name. " +
						"It can also be sourced from the `AUTH0_DOMAIN` environment variable.",
				},
				"audience": schema.StringAttribute{
					Optional: true,
					Description: "Your Auth0 audience when using a custom domain. " +
						"It can also be sourced from the AUTH0_AUDIENCE environment variable.",
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
					Description: "Your Auth0 client ID. " +
						"It can also be sourced from the AUTH0_CLIENT_ID environment variable.",
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
					Description: "Your Auth0 client secret. " +
						"It can also be sourced from the AUTH0_CLIENT_SECRET environment variable.",
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
					Description: "Your Auth0 management api access token. " +
						"It can also be sourced from the AUTH0_API_TOKEN environment variable. " +
						"It can be used instead of client_id + client_secret. " +
						"If both are specified, api_token will be used over client_id + client_secret fields.",
					MarkdownDescription: "Your Auth0 [management api access token]" +
						"(https://auth0.com/docs/security/tokens/access-tokens/management-api-access-tokens). " +
						"It can also be sourced from the `AUTH0_API_TOKEN` environment variable. " +
						"It can be used instead of `client_id` + `client_secret`. " +
						"If both are specified, `api_token` will be used over `client_id` + `client_secret` fields.",
				},
				"debug": schema.BoolAttribute{
					Optional: true,
					Description:         "Indicates whether to turn on debug mode.",
					MarkdownDescription: "Indicates whether to turn on debug mode.",
				},
			},
		}
	}
}

func (p *auth0Provider) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
	domain := os.Getenv("AUTH0_DOMAIN")
	clientID := os.Getenv("AUTH0_CLIENT_ID")
	clientSecret := os.Getenv("AUTH0_CLIENT_SECRET")
	apiToken := os.Getenv("AUTH0_API_TOKEN")
	audience := os.Getenv("AUTH0_AUDIENCE")
	debugStr := os.Getenv("AUTH0_DEBUG")

	var debug bool
	switch debugStr {
	case "1", "true", "on":
		debug = true
	default:
		debug = false
	}

	var data auth0ProviderModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	if data.Domain.ValueString() != "" {
		domain = data.Domain.ValueString()
	}
	if data.ClientID.ValueString() != "" {
		clientID = data.ClientID.ValueString()
	}
	if data.ClientSecret.ValueString() != "" {
		clientSecret = data.ClientSecret.ValueString()
	}
	if data.ApiToken.ValueString() != "" {
		apiToken = data.ApiToken.ValueString()
	}
	if data.Audience.ValueString() != "" {
		audience = data.Audience.ValueString()
	}
	if !data.Debug.IsNull() && !data.Debug.IsUnknown() {
		debug = data.Debug.ValueBool()
	}

	config, diag := config.ConfigureFrameworkProvider(request.TerraformVersion, domain, clientID, clientSecret, apiToken, audience, debug)
	if config != nil {
		response.ResourceData = config
		response.DataSourceData = config
	}

	response.Diagnostics.Append(diag...)
}

func (p *auth0Provider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *auth0Provider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

// NewAuth0Provider returns a terraform Framework provider.Provider.
func NewAuth0Provider() provider.Provider {
	return &auth0Provider{}
}
