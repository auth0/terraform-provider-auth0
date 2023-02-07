package client

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// NewResource will return a new auth0_client resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createClient,
		ReadContext:   readClient,
		UpdateContext: updateClient,
		DeleteContext: deleteClient,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can set up applications that use Auth0 for authentication " +
			"and configure allowed callback URLs and secrets for these applications.",
		Schema: resourceSchema,
	}
}

func createClient(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	client := expandClient(d)
	if err := api.Client.Create(client); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(client.GetClientID())

	return readClient(ctx, d, m)
}

func readClient(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	client, err := api.Client.Read(d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	result := multierror.Append(
		d.Set("client_id", client.GetClientID()),
		d.Set("client_secret", client.GetClientSecret()),
		d.Set("client_aliases", client.GetClientAliases()),
		d.Set("name", client.GetName()),
		d.Set("description", client.GetDescription()),
		d.Set("app_type", client.GetAppType()),
		d.Set("logo_uri", client.GetLogoURI()),
		d.Set("is_first_party", client.GetIsFirstParty()),
		d.Set("is_token_endpoint_ip_header_trusted", client.GetIsTokenEndpointIPHeaderTrusted()),
		d.Set("oidc_conformant", client.GetOIDCConformant()),
		d.Set("callbacks", client.GetCallbacks()),
		d.Set("allowed_logout_urls", client.GetAllowedLogoutURLs()),
		d.Set("allowed_origins", client.GetAllowedOrigins()),
		d.Set("allowed_clients", client.GetAllowedClients()),
		d.Set("grant_types", client.GetGrantTypes()),
		d.Set("organization_usage", client.GetOrganizationUsage()),
		d.Set("organization_require_behavior", client.GetOrganizationRequireBehavior()),
		d.Set("web_origins", client.GetWebOrigins()),
		d.Set("sso", client.GetSSO()),
		d.Set("sso_disabled", client.GetSSODisabled()),
		d.Set("cross_origin_auth", client.GetCrossOriginAuth()),
		d.Set("cross_origin_loc", client.GetCrossOriginLocation()),
		d.Set("custom_login_page_on", client.GetCustomLoginPageOn()),
		d.Set("custom_login_page", client.GetCustomLoginPage()),
		d.Set("form_template", client.GetFormTemplate()),
		d.Set("token_endpoint_auth_method", client.GetTokenEndpointAuthMethod()),
		d.Set("native_social_login", flattenCustomSocialConfiguration(client.GetNativeSocialLogin())),
		d.Set("jwt_configuration", flattenClientJwtConfiguration(client.GetJWTConfiguration())),
		d.Set("refresh_token", flattenClientRefreshTokenConfiguration(client.GetRefreshToken())),
		d.Set("encryption_key", client.GetEncryptionKey()),
		d.Set("addons", flattenClientAddons(client.Addons)),
		d.Set("mobile", flattenClientMobile(client.GetMobile())),
		d.Set("initiate_login_uri", client.GetInitiateLoginURI()),
		d.Set("signing_keys", client.SigningKeys),
		d.Set("client_metadata", client.ClientMetadata),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updateClient(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	client := expandClient(d)
	if clientHasChange(client) {
		if err := api.Client.Update(d.Id(), client); err != nil {
			return diag.FromErr(err)
		}
	}

	d.Partial(true)
	if err := rotateClientSecret(d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return readClient(ctx, d, m)
}

func deleteClient(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	if err := api.Client.Delete(d.Id()); err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
	}

	d.SetId("")
	return nil
}

func rotateClientSecret(d *schema.ResourceData, m interface{}) error {
	if !d.HasChange("client_secret_rotation_trigger") {
		return nil
	}

	api := m.(*management.Management)

	client, err := api.Client.RotateSecret(d.Id())
	if err != nil {
		return err
	}

	return d.Set("client_secret", client.GetClientSecret())
}
