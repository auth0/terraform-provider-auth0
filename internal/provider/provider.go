// Package provider wires the auth0 Terraform provider (Plugin Framework).
package provider

import (
	"context"
	"os"

	auth0client "github.com/auth0/terraform-provider-auth0/v2/internal/auth0"
	actionres "github.com/auth0/terraform-provider-auth0/v2/internal/services/action/auth0_action"
	triggerbindingsres "github.com/auth0/terraform-provider-auth0/v2/internal/services/action/auth0_trigger_bindings"
	clientres "github.com/auth0/terraform-provider-auth0/v2/internal/services/client"
	orgres "github.com/auth0/terraform-provider-auth0/v2/internal/services/organization"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interface.
var _ provider.Provider = (*auth0Provider)(nil)

// auth0Provider is the framework implementation of the Auth0 provider.
type auth0Provider struct {
	version string
}

// New returns a function that constructs a fresh provider instance.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &auth0Provider{version: version}
	}
}

// providerModel mirrors the provider HCL configuration.
type providerModel struct {
	Domain                    types.String `tfsdk:"domain"`
	Audience                  types.String `tfsdk:"audience"`
	ClientID                  types.String `tfsdk:"client_id"`
	ClientSecret              types.String `tfsdk:"client_secret"`
	APIToken                  types.String `tfsdk:"api_token"`
	ClientAssertionPrivateKey types.String `tfsdk:"client_assertion_private_key"`
	ClientAssertionAlgorithm  types.String `tfsdk:"client_assertion_signing_alg"`
	Debug                     types.Bool   `tfsdk:"debug"`
}

// Metadata sets the provider type name (`auth0`).
func (p *auth0Provider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "auth0"
	resp.Version = p.version
}

// Schema describes the provider-level configuration block.
func (p *auth0Provider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The Auth0 provider lets you manage Auth0 resources via the Management API.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Optional:    true,
				Description: "Your Auth0 tenant domain. May also be set via the AUTH0_DOMAIN environment variable.",
			},
			"audience": schema.StringAttribute{
				Optional:    true,
				Description: "Your Auth0 audience when using a custom domain. May also be set via AUTH0_AUDIENCE.",
			},
			"client_id": schema.StringAttribute{
				Optional:    true,
				Description: "Your Auth0 client ID. May also be set via AUTH0_CLIENT_ID.",
			},
			"client_secret": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Your Auth0 client secret. May also be set via AUTH0_CLIENT_SECRET.",
			},
			"api_token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "A pre-issued Management API token. May also be set via AUTH0_API_TOKEN. Mutually exclusive with the other auth methods; takes precedence when set.",
			},
			"client_assertion_private_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "PEM-encoded private key used to authenticate via Private Key JWT. May also be set via AUTH0_CLIENT_ASSERTION_PRIVATE_KEY.",
			},
			"client_assertion_signing_alg": schema.StringAttribute{
				Optional:    true,
				Description: "Signing algorithm used with the Private Key JWT (default: RS256). May also be set via AUTH0_CLIENT_ASSERTION_SIGNING_ALG.",
			},
			"debug": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether to log the underlying go-auth0 SDK HTTP traffic. May also be set via AUTH0_DEBUG.",
			},
		},
	}
}

// Configure builds the Auth0 Management client and stuffs it into the
// resp.ResourceData / resp.DataSourceData fields shared by every resource.
func (p *auth0Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := auth0client.Config{
		Domain:                    firstNonEmpty(data.Domain.ValueString(), os.Getenv("AUTH0_DOMAIN")),
		Audience:                  firstNonEmpty(data.Audience.ValueString(), os.Getenv("AUTH0_AUDIENCE")),
		ClientID:                  firstNonEmpty(data.ClientID.ValueString(), os.Getenv("AUTH0_CLIENT_ID")),
		ClientSecret:              firstNonEmpty(data.ClientSecret.ValueString(), os.Getenv("AUTH0_CLIENT_SECRET")),
		APIToken:                  firstNonEmpty(data.APIToken.ValueString(), os.Getenv("AUTH0_API_TOKEN")),
		ClientAssertionPrivateKey: firstNonEmpty(data.ClientAssertionPrivateKey.ValueString(), os.Getenv("AUTH0_CLIENT_ASSERTION_PRIVATE_KEY")),
		ClientAssertionAlgorithm:  firstNonEmpty(data.ClientAssertionAlgorithm.ValueString(), os.Getenv("AUTH0_CLIENT_ASSERTION_SIGNING_ALG")),
		Debug:                     data.Debug.ValueBool() || os.Getenv("AUTH0_DEBUG") == "true",
		ProviderVersion:           p.version,
		TerraformVersion:          req.TerraformVersion,
	}

	if cfg.Domain == "" {
		resp.Diagnostics.AddError(
			"Missing Auth0 domain",
			"The provider requires a `domain` (or AUTH0_DOMAIN env var) to be set.",
		)
		return
	}
	if cfg.Mode() == auth0client.AuthModeUnknown {
		resp.Diagnostics.AddError(
			"Missing Auth0 credentials",
			"Set one of: `api_token`, `client_id`+`client_secret`, or `client_id`+`client_assertion_private_key` (with the matching env vars).",
		)
		return
	}

	mgmt, err := auth0client.NewManagement(ctx, cfg)
	if err != nil {
		resp.Diagnostics.AddError("Failed to initialise Auth0 Management client", err.Error())
		return
	}

	resp.ResourceData = mgmt
	resp.DataSourceData = mgmt
}

// Resources returns every resource implemented by this provider.
func (p *auth0Provider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		clientres.NewResource,
		orgres.NewResource,
		actionres.NewResource,
		triggerbindingsres.NewResource,
	}
}

// DataSources returns every data source implemented by this provider.
func (p *auth0Provider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		actionres.NewDataSource,
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
