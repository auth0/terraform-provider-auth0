package auth0

import (
	"net/http"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	internalValidation "github.com/auth0/terraform-provider-auth0/auth0/internal/validation"
)

func newClient() *schema.Resource {
	return &schema.Resource{
		Create: createClient,
		Read:   readClient,
		Update: updateClient,
		Delete: deleteClient,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 140),
			},
			"client_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_secret": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"client_secret_rotation_trigger": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"app_type": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice(
					[]string{"native", "spa", "regular_web", "non_interactive"},
					false,
				),
			},
			"logo_uri": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"is_first_party": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"is_token_endpoint_ip_header_trusted": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"oidc_conformant": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"callbacks": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"allowed_logout_urls": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"grant_types": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Optional: true,
			},
			"organization_usage": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"deny", "allow", "require",
				}, false),
			},
			"organization_require_behavior": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"no_prompt", "pre_login_prompt",
				}, false),
			},
			"allowed_origins": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"allowed_clients": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"web_origins": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"jwt_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"lifetime_in_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"secret_encoded": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"scopes": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"alg": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"encryption_key": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"sso": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"sso_disabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"cross_origin_auth": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"cross_origin_loc": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"custom_login_page_on": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"custom_login_page": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"form_template": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"addons": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"samlp": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"audience": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"recipient": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"mappings": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     schema.TypeString,
									},
									"create_upn_claim": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
									},
									"passthrough_claims_with_no_mapping": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
									},
									"map_unknown_claims_as_is": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"map_identities": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
									},
									"signature_algorithm": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "rsa-sha1",
									},
									"digest_algorithm": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "sha1",
									},
									"destination": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"lifetime_in_seconds": {
										Type:     schema.TypeInt,
										Optional: true,
										Default:  3600,
									},
									"sign_response": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"name_identifier_format": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified",
									},
									"name_identifier_probes": {
										Type:     schema.TypeList,
										Elem:     &schema.Schema{Type: schema.TypeString},
										Optional: true,
									},
									"authn_context_class_ref": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"typed_attributes": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
									},
									"include_attribute_name_format": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
									},
									"logout": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"callback": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"slo_enabled": {
													Type:     schema.TypeBool,
													Optional: true,
												},
											},
										},
									},
									"binding": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"signing_cert": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"wsfed": {
							Type:     schema.TypeMap,
							Optional: true,
						},
					},
				},
			},
			"token_endpoint_auth_method": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"none",
					"client_secret_post",
					"client_secret_basic",
				}, false),
			},
			"client_metadata": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     schema.TypeString,
			},
			"mobile": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"android": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"app_package_name": {
										Type:     schema.TypeString,
										Optional: true,
										AtLeastOneOf: []string{
											"mobile.0.android.0.app_package_name",
											"mobile.0.android.0.sha256_cert_fingerprints",
										},
									},
									"sha256_cert_fingerprints": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
										AtLeastOneOf: []string{
											"mobile.0.android.0.app_package_name",
											"mobile.0.android.0.sha256_cert_fingerprints",
										},
									},
								},
							},
						},
						"ios": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"team_id": {
										Type:     schema.TypeString,
										Optional: true,
										AtLeastOneOf: []string{
											"mobile.0.ios.0.team_id",
											"mobile.0.ios.0.app_bundle_identifier",
										},
									},
									"app_bundle_identifier": {
										Type:     schema.TypeString,
										Optional: true,
										AtLeastOneOf: []string{
											"mobile.0.ios.0.team_id",
											"mobile.0.ios.0.app_bundle_identifier",
										},
									},
								},
							},
						},
					},
				},
			},
			"initiate_login_uri": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.IsURLWithScheme([]string{"https"}),
					internalValidation.IsURLWithNoFragment,
				),
			},
			"native_social_login": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"apple": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
						},
						"facebook": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"refresh_token": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"rotation_type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"rotating",
								"non-rotating",
							}, false),
						},
						"expiration_type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"expiring",
								"non-expiring",
							}, false),
						},
						"leeway": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"token_lifetime": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"infinite_token_lifetime": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"infinite_idle_token_lifetime": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"idle_token_lifetime": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
			"signing_keys": {
				Type:      schema.TypeList,
				Elem:      &schema.Schema{Type: schema.TypeMap},
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func createClient(d *schema.ResourceData, m interface{}) error {
	client := expandClient(d)
	api := m.(*management.Management)
	if err := api.Client.Create(client); err != nil {
		return err
	}
	d.SetId(auth0.StringValue(client.ClientID))
	return readClient(d, m)
}

func readClient(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)
	client, err := api.Client.Read(d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	err = d.Set("client_id", client.ClientID)
	err = multierror.Append(err, d.Set("client_secret", client.ClientSecret))
	err = multierror.Append(err, d.Set("name", client.Name))
	err = multierror.Append(err, d.Set("description", client.Description))
	err = multierror.Append(err, d.Set("app_type", client.AppType))
	err = multierror.Append(err, d.Set("logo_uri", client.LogoURI))
	err = multierror.Append(err, d.Set("is_first_party", client.IsFirstParty))
	err = multierror.Append(err, d.Set("is_token_endpoint_ip_header_trusted", client.IsTokenEndpointIPHeaderTrusted))
	err = multierror.Append(err, d.Set("oidc_conformant", client.OIDCConformant))
	err = multierror.Append(err, d.Set("callbacks", client.Callbacks))
	err = multierror.Append(err, d.Set("allowed_logout_urls", client.AllowedLogoutURLs))
	err = multierror.Append(err, d.Set("allowed_origins", client.AllowedOrigins))
	err = multierror.Append(err, d.Set("allowed_clients", client.AllowedClients))
	err = multierror.Append(err, d.Set("grant_types", client.GrantTypes))
	err = multierror.Append(err, d.Set("organization_usage", client.OrganizationUsage))
	err = multierror.Append(err, d.Set("organization_require_behavior", client.OrganizationRequireBehavior))
	err = multierror.Append(err, d.Set("web_origins", client.WebOrigins))
	err = multierror.Append(err, d.Set("sso", client.SSO))
	err = multierror.Append(err, d.Set("sso_disabled", client.SSODisabled))
	err = multierror.Append(err, d.Set("cross_origin_auth", client.CrossOriginAuth))
	err = multierror.Append(err, d.Set("cross_origin_loc", client.CrossOriginLocation))
	err = multierror.Append(err, d.Set("custom_login_page_on", client.CustomLoginPageOn))
	err = multierror.Append(err, d.Set("custom_login_page", client.CustomLoginPage))
	err = multierror.Append(err, d.Set("form_template", client.FormTemplate))
	err = multierror.Append(err, d.Set("token_endpoint_auth_method", client.TokenEndpointAuthMethod))
	err = multierror.Append(err, d.Set("native_social_login", flattenCustomSocialConfiguration(client.NativeSocialLogin)))
	err = multierror.Append(err, d.Set("jwt_configuration", flattenClientJwtConfiguration(client.JWTConfiguration)))
	err = multierror.Append(err, d.Set("refresh_token", flattenClientRefreshTokenConfiguration(client.RefreshToken)))
	err = multierror.Append(err, d.Set("encryption_key", client.EncryptionKey))
	err = multierror.Append(err, d.Set("addons", flattenClientAddons(client.Addons)))
	err = multierror.Append(err, d.Set("client_metadata", client.ClientMetadata))
	err = multierror.Append(err, d.Set("mobile", flattenClientMobile(client.Mobile)))
	err = multierror.Append(err, d.Set("initiate_login_uri", client.InitiateLoginURI))
	err = multierror.Append(err, d.Set("signing_keys", client.SigningKeys))

	return err
}

func updateClient(d *schema.ResourceData, m interface{}) error {
	client := expandClient(d)
	api := m.(*management.Management)
	if clientHasChange(client) {
		err := api.Client.Update(d.Id(), client)
		if err != nil {
			return err
		}
	}
	d.Partial(true)
	err := rotateClientSecret(d, m)
	if err != nil {
		return err
	}
	d.Partial(false)
	return readClient(d, m)
}

func deleteClient(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)
	err := api.Client.Delete(d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
	}
	return err
}

func expandClient(d *schema.ResourceData) *management.Client {
	client := &management.Client{
		Name:                           String(d, "name"),
		Description:                    String(d, "description"),
		AppType:                        String(d, "app_type"),
		LogoURI:                        String(d, "logo_uri"),
		IsFirstParty:                   Bool(d, "is_first_party"),
		IsTokenEndpointIPHeaderTrusted: Bool(d, "is_token_endpoint_ip_header_trusted"),
		OIDCConformant:                 Bool(d, "oidc_conformant"),
		Callbacks:                      Slice(d, "callbacks"),
		AllowedLogoutURLs:              Slice(d, "allowed_logout_urls"),
		AllowedOrigins:                 Slice(d, "allowed_origins"),
		AllowedClients:                 Slice(d, "allowed_clients"),
		GrantTypes:                     Slice(d, "grant_types"),
		OrganizationUsage:              String(d, "organization_usage"),
		OrganizationRequireBehavior:    String(d, "organization_require_behavior"),
		WebOrigins:                     Slice(d, "web_origins"),
		SSO:                            Bool(d, "sso"),
		SSODisabled:                    Bool(d, "sso_disabled"),
		CrossOriginAuth:                Bool(d, "cross_origin_auth"),
		CrossOriginLocation:            String(d, "cross_origin_loc"),
		CustomLoginPageOn:              Bool(d, "custom_login_page_on"),
		CustomLoginPage:                String(d, "custom_login_page"),
		FormTemplate:                   String(d, "form_template"),
		TokenEndpointAuthMethod:        String(d, "token_endpoint_auth_method"),
		InitiateLoginURI:               String(d, "initiate_login_uri"),
	}

	List(d, "refresh_token", IsNewResource(), HasChange()).Elem(func(d ResourceData) {
		client.RefreshToken = &management.ClientRefreshToken{
			RotationType:              String(d, "rotation_type"),
			ExpirationType:            String(d, "expiration_type"),
			Leeway:                    Int(d, "leeway"),
			TokenLifetime:             Int(d, "token_lifetime"),
			InfiniteTokenLifetime:     Bool(d, "infinite_token_lifetime"),
			InfiniteIdleTokenLifetime: Bool(d, "infinite_idle_token_lifetime"),
			IdleTokenLifetime:         Int(d, "idle_token_lifetime"),
		}
	})

	List(d, "jwt_configuration").Elem(func(d ResourceData) {
		client.JWTConfiguration = &management.ClientJWTConfiguration{
			LifetimeInSeconds: Int(d, "lifetime_in_seconds"),
			SecretEncoded:     Bool(d, "secret_encoded", IsNewResource()),
			Algorithm:         String(d, "alg"),
			Scopes:            Map(d, "scopes"),
		}
	})

	if m := Map(d, "encryption_key"); m != nil {
		client.EncryptionKey = map[string]string{}
		for k, v := range m {
			client.EncryptionKey[k] = v.(string)
		}
	}

	List(d, "addons").Elem(func(d ResourceData) {
		client.Addons = &management.ClientAddons{}

		List(d, "samlp").Elem(func(d ResourceData) {
			client.Addons.SAML = &management.ClientAddonSAML{
				Audience:                       String(d, "audience"),
				Recipient:                      String(d, "recipient"),
				Mappings:                       Map(d, "mappings"),
				CreateUpnClaim:                 Bool(d, "create_upn_claim"),
				PassThroughClaimsWithNoMapping: Bool(d, "passthrough_claims_with_no_mapping"),
				MapUnknownClaimsAsIs:           Bool(d, "map_unknown_claims_as_is"),
				MapIdentities:                  Bool(d, "map_identities"),
				SignatureAlgorithm:             String(d, "signature_algorithm"),
				DigestAlgorithm:                String(d, "digest_algorithm"),
				Destination:                    String(d, "destination"),
				LifetimeInSeconds:              Int(d, "lifetime_in_seconds"),
				SignResponse:                   Bool(d, "sign_response"),
				NameIdentifierFormat:           String(d, "name_identifier_format"),
				NameIdentifierProbes:           StringSlice(d, "name_identifier_probes"),
				AuthnContextClassRef:           String(d, "authn_context_class_ref"),
				TypedAttributes:                Bool(d, "typed_attributes"),
				IncludeAttributeNameFormat:     Bool(d, "include_attribute_name_format"),
				Binding:                        String(d, "binding"),
				SigningCert:                    String(d, "signing_cert"),
			}

			List(d, "logout").Elem(func(d ResourceData) {
				client.Addons.SAML.Logout = &management.ClientAddonSAMLLogout{
					Callback:   String(d, "callback"),
					SLOEnabled: Bool(d, "slo_enabled"),
				}
			})
		})
	})

	if v, ok := d.GetOk("client_metadata"); ok {
		client.ClientMetadata = make(map[string]string)
		for key, value := range v.(map[string]interface{}) {
			client.ClientMetadata[key] = (value.(string))
		}
	}

	List(d, "native_social_login").Elem(func(d ResourceData) {
		client.NativeSocialLogin = &management.ClientNativeSocialLogin{}

		List(d, "apple").Elem(func(d ResourceData) {
			m := make(MapData)
			m.Set("enabled", Bool(d, "enabled"))

			client.NativeSocialLogin.Apple = m
		})

		List(d, "facebook").Elem(func(d ResourceData) {
			m := make(MapData)
			m.Set("enabled", Bool(d, "enabled"))

			client.NativeSocialLogin.Facebook = m
		})
	})

	List(d, "mobile").Elem(func(d ResourceData) {
		client.Mobile = &management.ClientMobile{}

		List(d, "android").Elem(func(d ResourceData) {
			client.Mobile.Android = &management.ClientMobileAndroid{
				AppPackageName:         String(d, "app_package_name"),
				SHA256CertFingerprints: StringSlice(d, "sha256_cert_fingerprints"),
			}
		})

		List(d, "ios").Elem(func(d ResourceData) {
			client.Mobile.IOS = &management.ClientMobileIOS{
				TeamID:              String(d, "team_id"),
				AppBundleIdentifier: String(d, "app_bundle_identifier"),
			}
		})
	})

	return client
}

func rotateClientSecret(d *schema.ResourceData, m interface{}) error {
	if d.HasChange("client_secret_rotation_trigger") {
		api := m.(*management.Management)
		client, err := api.Client.RotateSecret(d.Id())
		if err != nil {
			return err
		}
		d.Set("client_secret", client.ClientSecret)
	}
	d.SetPartial("client_secret_rotation_trigger")
	return nil
}

func clientHasChange(c *management.Client) bool {
	return c.String() != "{}"
}

func flattenCustomSocialConfiguration(customSocial *management.ClientNativeSocialLogin) []interface{} {
	if customSocial != nil {
		m := make(map[string]interface{})

		if customSocial.Apple != nil {
			m["apple"] = map[string]interface{}{
				"enabled": customSocial.Apple["enabled"],
			}
		}
		if customSocial.Facebook != nil {
			m["facebook"] = map[string]interface{}{
				"enabled": customSocial.Facebook["enabled"],
			}
		}

		return []interface{}{m}
	}

	return nil
}

func flattenClientJwtConfiguration(jwt *management.ClientJWTConfiguration) []interface{} {
	m := make(map[string]interface{})
	if jwt != nil {
		m["lifetime_in_seconds"] = jwt.LifetimeInSeconds
		m["secret_encoded"] = jwt.SecretEncoded
		m["scopes"] = jwt.Scopes
		m["alg"] = jwt.Algorithm
	}
	return []interface{}{m}
}

func flattenClientRefreshTokenConfiguration(refreshToken *management.ClientRefreshToken) []interface{} {
	m := make(map[string]interface{})
	if refreshToken != nil {
		m["rotation_type"] = refreshToken.RotationType
		m["expiration_type"] = refreshToken.ExpirationType
		m["leeway"] = refreshToken.Leeway
		m["token_lifetime"] = refreshToken.TokenLifetime
		m["infinite_token_lifetime"] = refreshToken.InfiniteTokenLifetime
		m["infinite_idle_token_lifetime"] = refreshToken.InfiniteIdleTokenLifetime
		m["idle_token_lifetime"] = refreshToken.IdleTokenLifetime
	}
	return []interface{}{m}
}

func flattenClientAddons(addons *management.ClientAddons) []interface{} {
	m := make(map[string]interface{})

	if addons == nil || addons.SAML == nil {
		return nil
	}

	samlpMap := map[string]interface{}{
		"audience":                           addons.SAML.Audience,
		"recipient":                          addons.SAML.Recipient,
		"mappings":                           addons.SAML.Mappings,
		"create_upn_claim":                   addons.SAML.CreateUpnClaim,
		"passthrough_claims_with_no_mapping": addons.SAML.PassThroughClaimsWithNoMapping,
		"map_unknown_claims_as_is":           addons.SAML.MapUnknownClaimsAsIs,
		"map_identities":                     addons.SAML.MapIdentities,
		"signature_algorithm":                addons.SAML.SignatureAlgorithm,
		"digest_algorithm":                   addons.SAML.DigestAlgorithm,
		"destination":                        addons.SAML.Destination,
		"lifetime_in_seconds":                addons.SAML.LifetimeInSeconds,
		"sign_response":                      addons.SAML.SignResponse,
		"name_identifier_format":             addons.SAML.NameIdentifierFormat,
		"name_identifier_probes":             addons.SAML.NameIdentifierProbes,
		"authn_context_class_ref":            addons.SAML.AuthnContextClassRef,
		"typed_attributes":                   addons.SAML.TypedAttributes,
		"include_attribute_name_format":      addons.SAML.IncludeAttributeNameFormat,
		"binding":                            addons.SAML.Binding,
		"signing_cert":                       addons.SAML.SigningCert,
	}

	if addons.SAML.Logout != nil {
		logoutMap := map[string]interface{}{
			"callback":    addons.SAML.Logout.Callback,
			"slo_enabled": addons.SAML.Logout.SLOEnabled,
		}
		samlpMap["logout"] = []interface{}{logoutMap}
	}

	m["samlp"] = []interface{}{samlpMap}

	return []interface{}{m}
}

func flattenClientMobile(mobile *management.ClientMobile) []interface{} {
	m := make(map[string]interface{})

	if mobile == nil {
		return nil
	}

	if mobile.Android != nil {
		m["android"] = []interface{}{
			map[string]interface{}{
				"app_package_name":         mobile.Android.AppPackageName,
				"sha256_cert_fingerprints": mobile.Android.SHA256CertFingerprints,
			},
		}
	}
	if mobile.IOS != nil {
		m["ios"] = []interface{}{
			map[string]interface{}{
				"team_id":               mobile.IOS.TeamID,
				"app_bundle_identifier": mobile.IOS.AppBundleIdentifier,
			},
		}
	}

	return []interface{}{m}
}
