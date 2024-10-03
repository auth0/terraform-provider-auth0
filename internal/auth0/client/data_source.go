package client

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewDataSource will return a new auth0_client data source.
func NewDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readClientForDataSource,
		Description: "Data source to retrieve a specific Auth0 application client by `client_id` or `name`.",
		Schema:      dataSourceSchema(),
	}
}

func dataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(internalSchema.Clone(NewResource().Schema))

	internalSchema.SetExistingAttributesAsOptional(dataSourceSchema, "name", "client_id")

	dataSourceSchema["name"].Description = "The name of the client. If not provided, `client_id` must be set."
	dataSourceSchema["client_id"].Description = "The ID of the client. If not provided, `name` must be set."

	dataSourceSchema["client_secret"] = &schema.Schema{
		Type:      schema.TypeString,
		Computed:  true,
		Sensitive: true,
		Description: "Secret for the client. Keep this private. To access this attribute you need to add the " +
			"`read:client_keys` scope to the Terraform client. Otherwise, the attribute will contain an empty string.",
	}

	dataSourceSchema["token_endpoint_auth_method"] = &schema.Schema{
		Type:     schema.TypeString,
		Computed: true,
		Description: "The authentication method for the token endpoint. " +
			"Results include `none` (public client without a client secret), " +
			"`client_secret_post` (client uses HTTP POST parameters), " +
			"`client_secret_basic` (client uses HTTP Basic), " +
			"Managing a client's authentication method can be done via the " +
			"`auth0_client_credentials` resource.",
	}

	dataSourceSchema["signed_request_object"] = &schema.Schema{
		Type:        schema.TypeSet,
		Computed:    true,
		Description: "Configuration for JWT-secured Authorization Requests(JAR).",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"required": {
					Type:        schema.TypeBool,
					Computed:    true,
					Description: "Require JWT-secured authorization requests.",
				},
				"credentials": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "Credentials that will be enabled on the client for JWT-secured authorization requests.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"id": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The ID of the client credential.",
							},
							"name": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "Friendly name for a credential.",
							},
							"key_id": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The key identifier of the credential, generated on creation.",
							},
							"credential_type": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "Credential type. Supported types: `public_key`.",
							},
							"algorithm": {
								Type:     schema.TypeString,
								Computed: true,
								Description: "Algorithm which will be used with the credential. " +
									"Can be one of `RS256`, `RS384`, `PS256`. If not specified, " +
									"`RS256` will be used.",
							},
							"created_at": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The ISO 8601 formatted date the credential was created.",
							},
							"updated_at": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The ISO 8601 formatted date the credential was updated.",
							},
							"expires_at": {
								Type:     schema.TypeString,
								Computed: true,
								Description: "The ISO 8601 formatted date representing " +
									"the expiration of the credential.",
							},
						},
					},
				},
			},
		},
	}

	dataSourceSchema["client_authentication_methods"] = &schema.Schema{
		Type:        schema.TypeSet,
		Computed:    true,
		Description: "Defines client authentication methods.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"private_key_jwt": {
					Type:        schema.TypeSet,
					Computed:    true,
					Description: "If this is defined, the client is enabled to use the Private Key JWT authentication method.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"credentials": {
								Type:        schema.TypeList,
								Computed:    true,
								Description: "Credentials that will be enabled on the client for Private Key JWT authentication.",
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"id": {
											Type:        schema.TypeString,
											Computed:    true,
											Description: "The ID of the client credential.",
										},
										"name": {
											Type:        schema.TypeString,
											Computed:    true,
											Description: "Friendly name for a credential.",
										},
										"key_id": {
											Type:        schema.TypeString,
											Computed:    true,
											Description: "The key identifier of the credential, generated on creation.",
										},
										"credential_type": {
											Type:        schema.TypeString,
											Computed:    true,
											Description: "Credential type. Supported types: `public_key`.",
										},
										"algorithm": {
											Type:     schema.TypeString,
											Computed: true,
											Description: "Algorithm which will be used with the credential. " +
												"Can be one of `RS256`, `RS384`, `PS256`. If not specified, " +
												"`RS256` will be used.",
										},
										"created_at": {
											Type:        schema.TypeString,
											Computed:    true,
											Description: "The ISO 8601 formatted date the credential was created.",
										},
										"updated_at": {
											Type:        schema.TypeString,
											Computed:    true,
											Description: "The ISO 8601 formatted date the credential was updated.",
										},
										"expires_at": {
											Type:     schema.TypeString,
											Computed: true,
											Description: "The ISO 8601 formatted date representing " +
												"the expiration of the credential.",
										},
									},
								},
							},
						},
					},
				},
				"tls_client_auth": {
					Type:        schema.TypeSet,
					Computed:    true,
					Description: "If this is defined, the client is enabled to use the CA-based mTLS authentication method.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"credentials": {
								Type:        schema.TypeList,
								Computed:    true,
								Description: "Credentials that will be enabled on the client for CA-based mTLS authentication.",
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"id": {
											Type:        schema.TypeString,
											Computed:    true,
											Description: "The ID of the client credential.",
										},
										"name": {
											Type:        schema.TypeString,
											Computed:    true,
											Description: "Friendly name for a credential.",
										},
										"credential_type": {
											Type:        schema.TypeString,
											Computed:    true,
											Description: "Credential type. Supported types: `cert_subject_dn`.",
										},
										"subject_dn": {
											Type:        schema.TypeString,
											Computed:    true,
											Description: "Subject Distinguished Name. Mutually exlusive with `pem` property.",
										},
										"created_at": {
											Type:        schema.TypeString,
											Computed:    true,
											Description: "The ISO 8601 formatted date the credential was created.",
										},
										"updated_at": {
											Type:        schema.TypeString,
											Computed:    true,
											Description: "The ISO 8601 formatted date the credential was updated.",
										},
									},
								},
							},
						},
					},
				},
				"self_signed_tls_client_auth": {
					Type:     schema.TypeSet,
					Computed: true,
					Description: "If this is defined, the client is enabled to use the mTLS authentication " +
						"method utilizing a self-signed certificate.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"credentials": {
								Type:     schema.TypeList,
								Computed: true,
								Description: "Credentials that will be enabled on the client for mTLS " +
									"authentication utilizing self-signed certificates.",
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"id": {
											Type:        schema.TypeString,
											Computed:    true,
											Description: "The ID of the client credential.",
										},
										"name": {
											Type:        schema.TypeString,
											Computed:    true,
											Description: "Friendly name for a credential.",
										},
										"credential_type": {
											Type:        schema.TypeString,
											Computed:    true,
											Description: "Credential type. Supported types: `x509_cert`.",
										},
										"created_at": {
											Type:        schema.TypeString,
											Computed:    true,
											Description: "The ISO 8601 formatted date the credential was created.",
										},
										"updated_at": {
											Type:        schema.TypeString,
											Computed:    true,
											Description: "The ISO 8601 formatted date the credential was updated.",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return dataSourceSchema
}

func readClientForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	clientID := data.Get("client_id").(string)
	if clientID != "" {
		data.SetId(clientID)

		client, err := api.Client.Read(ctx, data.Id())
		if err != nil {
			return diag.FromErr(err)
		}

		err = flattenClientForDataSource(ctx, api, data, client)

		return diag.FromErr(err)
	}

	name := data.Get("name").(string)
	if name == "" {
		return diag.Errorf("One of 'client_id' or 'name' is required.")
	}

	var page int
	for {
		clients, err := api.Client.List(
			ctx,
			management.Page(page),
			management.PerPage(100),
		)
		if err != nil {
			return diag.FromErr(err)
		}

		for _, client := range clients.Clients {
			if client.GetName() == name {
				data.SetId(client.GetClientID())
				err = flattenClientForDataSource(ctx, api, data, client)
				return diag.FromErr(err)
			}
		}

		if !clients.HasNext() {
			break
		}

		page++
	}

	return diag.Errorf("No client found with \"name\" = %q", name)
}
