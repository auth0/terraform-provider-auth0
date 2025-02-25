## v1.13.1

FEATURES:
- `resource/auth0_prompt_screen_renderer`: Add support for new screens( `EA` Release) ([#1158](https://github.com/auth0/terraform-provider-auth0/pull/1158/))


## v1.13.0

ENHANCEMENTS:

- `resource/auth0_connection`: Add support to set `global_token_revocation_jwt_iss` and `global_token_revocation_jwt_sub` property for a connection ([#1142](https://github.com/auth0/terraform-provider-auth0/pull/1142/))
- `data-source/auth0_user`: Add support to retrieve a user via lucene query ([#1141](https://github.com/auth0/terraform-provider-auth0/pull/1141/))
- `resource/auth0_prompt_partials`: Add support to set prompt as `customized-consent` ([#1151](https://github.com/auth0/terraform-provider-auth0/pull/1151/))
- `resource/auth0_prompt_screen_partial`: Add support to set `form_content` in insertion_points ([#1151](https://github.com/auth0/terraform-provider-auth0/pull/1151/))
- `resource/auth0_prompt_screen_partials`: Add support to set `form_content` in insertion_points ([#1151](https://github.com/auth0/terraform-provider-auth0/pull/1151/))
- `resource/auth0_client`: Add support to set options for `google` as part of `native_social_login` ([#1150](https://github.com/auth0/terraform-provider-auth0/pull/1150/))
- `resource/auth0_branding_theme`: Add support to set `captcha_widget_theme` property ([#1154](https://github.com/auth0/terraform-provider-auth0/pull/1154/))
- `resource/auth0_trigger_action`: Add `custom-token-exchange` and `custom-email-provider` to list of supported triggers ([#1155](https://github.com/auth0/terraform-provider-auth0/pull/1155/))
- `resource/auth0_trigger_actions`: Add `custom-token-exchange` and `custom-email-provider` to list of supported triggers ([#1155](https://github.com/auth0/terraform-provider-auth0/pull/1155/))
- `resource/auth0_log_stream`: Add support to set `is_priority` property ([#1102](https://github.com/auth0/terraform-provider-auth0/pull/1102/))


## v1.12.0

FEATURES:
- `resource/auth0_prompt_screen_renderer`: Add support for new screens (`EA` Release) ([#1144](https://github.com/auth0/terraform-provider-auth0/pull/1144/))

## v1.11.1

BUG FIXES:
- `resource/auth0_client`: Add missing expand/flatten rules for `token_exchange` param ([#1145](https://github.com/auth0/terraform-provider-auth0/pull/1145/))
- `resource/auth0_action`: Add clause to support node18 action for `custom-token-exchange` trigger ([#1145](https://github.com/auth0/terraform-provider-auth0/pull/1145/))


## v1.11.0

FEATURES:
- `resource/auth0_token_exchange_profile`: Add a resource for managing Token Exchange Profile ([#1119](https://github.com/auth0/terraform-provider-auth0/pull/1119))
- `data-source/auth0_token_exchange_profile`: Add a data-source for retrieving Token Exchange Profile ([#1119](https://github.com/auth0/terraform-provider-auth0/pull/1119))

ENHANCEMENTS:
- `resource/auth0_client`: Add support to set `token_exchange` property for a client ([#1119](https://github.com/auth0/terraform-provider-auth0/pull/1119))
- `resource/auth0_connection`: Add support to set `authentication_methods` and `passkey_options` property for a connection ([#1099](https://github.com/auth0/terraform-provider-auth0/pull/1099))


## v1.10.0

ENHANCEMENTS:
- `resource/auth0_connection` - Add a Support `password reset by otp code` email template and `verification_method` field in connection schema. ([#1113](https://github.com/auth0/terraform-provider-auth0/pull/1113/))
- `resource/auth0_connection` - Support never_on_login as an allowed value for set_user_root_attributes ([#1123](https://github.com/auth0/terraform-provider-auth0/pull/1123/))

BUG FIXES:
- `resource/auth0_actions` - Fix: Dynamic block for action.Secrets and iterates via k/v ([#1115](https://github.com/auth0/terraform-provider-auth0/pull/1115/))
- `data-source/auth0_connection`- Fix: IdP-initiated SSO Behavior by adding the missing enabled field to options param connections ([#1105](https://github.com/auth0/terraform-provider-auth0/pull/1105/))
- `resource/auth0_flow` - Fix: Update example and docs for auth0_flow resource([#1129](https://github.com/auth0/terraform-provider-auth0/pull/1129/))
- `resource/auth0_prompt_screen_renderer` - Fix: Update example and docs for auth0_prompt_screen_renderer resource([#1127](https://github.com/auth0/terraform-provider-auth0/pull/1127/))

NOTES:
- Update workflow to install terraform manually since ubuntu image has removed it ([#1116](https://github.com/auth0/terraform-provider-auth0/pull/1116/))


## v1.9.1

ENHANCEMENTS:
- `resource/auth0_prompt_screen_renderer`: Update docs & unit tests related to auth0_prompt_screen_renderer  (`EA` Release) ([#1101](https://github.com/auth0/terraform-provider-auth0/pull/1101/))
- `resource/auth0_flow_vault_connection`: Update `setup` & `ready` attributes schema for auth0_flow_vault_connection ([#1103](https://github.com/auth0/terraform-provider-auth0/pull/1103/))


## v1.9.0

FEATURES:
- `resource/auth0_prompt_screen_renderer`: Add a resource for configuring the settings of prompt-screen ([#1077](https://github.com/auth0/terraform-provider-auth0/pull/1077))
- `data-source/auth0_prompt_screen_renderer`: Add a data-source for retrieving prompt-screen settings ([#1077](https://github.com/auth0/terraform-provider-auth0/pull/1077))


## v1.8.0

FEATURES:
- `resource/auth0_self_service_profile_custom_text`: Add new resource which allow to set custom text for SSO Profile ([#1075](https://github.com/auth0/terraform-provider-auth0/pull/1075/))
- `data-source/auth0_clients`: Add data-source which allows retrieving a list of clients with filters ([#1080](https://github.com/auth0/terraform-provider-auth0/pull/1080/))

ENHANCEMENTS:
- `resource/auth0_client`: Add support for setting `oidc_logout`, which includes `backchannel_logout_urls` and `backchannel_logout_initiators`. The `backchannel_logout_initiators` property supports `mode` and `selected_initiators` for more granular control ([#1045](https://github.com/auth0/terraform-provider-auth0/pull/1045/))
- `resource/auth0_self_service_profile`: Add support for setting `name`, `description`, `allowed_strategies` ([#1075](https://github.com/auth0/terraform-provider-auth0/pull/1075/))

BUG FIXES:
- `resource/auth0_form`: Update messages property with expand rule ([#1088](https://github.com/auth0/terraform-provider-auth0/pull/1088/))


## v1.7.3

NOTES:
This is a placeholder change to bump the version, as we are trying to resolve issues publishing to the Terraform Registry.

## v1.7.2

BUG FIXES:

- `resource/auth0_flow`: Fixed an issue with updating flows  ([#1058](https://github.com/auth0/terraform-provider-auth0/pull/1058/))
- `resource/auth0_form`: Fixed an issue with updating forms  ([#1058](https://github.com/auth0/terraform-provider-auth0/pull/1058/))

ENHANCEMENTS:
- `provider`: Added explicit check for handling missing env variables ([#1065](https://github.com/auth0/terraform-provider-auth0/pull/1065/))
- `resource/auth0_email_provider`: Added support for Custom Email Provider ([#1064](https://github.com/auth0/terraform-provider-auth0/pull/1064/))

NOTES:
- `resource/auth0_trigger_action`: Remove IGA-* triggers ([#1063](https://github.com/auth0/terraform-provider-auth0/pull/1063/))


## v1.7.1

BUG FIXES:

- `data-source/auth0_organization`: Implemented enhanced error handling to gracefully manage "Forbidden" errors when retrieving the list of client grants via the `auth0_organization` data source. This addresses cases where the feature is not enabled for the user, preventing the error from disrupting the process ([#1049](https://github.com/auth0/terraform-provider-auth0/pull/1049/))
- `resource/auth0_connection`: Updated documentation to clarify that `user_id_attribute` can be either `oid` or `sub` for Azure AD connections ([#1047](https://github.com/auth0/terraform-provider-auth0/pull/1047/))
- `resource/auth0_form`: Updated docs to use `auth0_form` in examples  ([#1046](https://github.com/auth0/terraform-provider-auth0/pull/1046/))

ENHANCEMENTS:
- `data-source/auth0_roles`: Updated from `Offset Pagination` to `Checkpoint Pagination` to retrieve more than 1,000 role users ([#1048](https://github.com/auth0/terraform-provider-auth0/pull/1048/))


## v1.7.0

FEATURES:

- `resource/auth0_encryption_key_manager`: Add new resource for re-keying of tenant master key ([#1031](https://github.com/auth0/terraform-provider-auth0/pull/1031/))
- `resource/auth0_encryption_key_manager`: Add support for `customer_provided_root_key` for BYOK ([#1041](https://github.com/auth0/terraform-provider-auth0/pull/1041/))
- `resource/auth0_organization_client_grant`: Add new resource for managing association of client-grant and organization ([#1027](https://github.com/auth0/terraform-provider-auth0/pull/1027/))
- `resource/auth0_form`: Add new resource for managing Forms ([#1039](https://github.com/auth0/terraform-provider-auth0/pull/1039/))
- `resource/auth0_flow`: Add new resource for managing Flows ([#1039](https://github.com/auth0/terraform-provider-auth0/pull/1039/))
- `resource/auth0_flow_vault_connection`: Add new resource for managing Flow Vault Connection ([#1039](https://github.com/auth0/terraform-provider-auth0/pull/1039/))
- `data-source/auth0_form`: Add a data-source for retrieving Form ([#1039](https://github.com/auth0/terraform-provider-auth0/pull/1039/))
- `data-source/auth0_flow`: Add a data-source for retrieving Flow ([#1039](https://github.com/auth0/terraform-provider-auth0/pull/1039/))
- `data-source/auth0_flow_vault_connection`: Add a data-source for retrieving Flow Vault Connection. ([#1039](https://github.com/auth0/terraform-provider-auth0/pull/1039/))

ENHANCEMENTS:

- `resource/auth0_connection`: Add support for `strategy_version` to be configurable ([#1024](https://github.com/auth0/terraform-provider-auth0/pull/1024/))  
- `resource/auth0_connection`: Add support for `user_id_attribute` in options attribute for AzureAD connections ([#1028](https://github.com/auth0/terraform-provider-auth0/pull/1028/))
- `data-source/auth0_organization`: Add support to fetch list of associated client grants ([#1027](https://github.com/auth0/terraform-provider-auth0/pull/1027/))
- `resource/auth0_tenant`: Add support for `acr_values_supported`, `pushed_authorization_requests_supported` and `remove_alg_from_jwks` configuration ([#1015](https://github.com/auth0/terraform-provider-auth0/pull/1015/))
- `resource/auth0_client_credentials`: Add support for setting `signed_request_object`, `tls_client_auth` and `self_signed_tls_client_auth` ([#1015](https://github.com/auth0/terraform-provider-auth0/pull/1015/))
- `resource/auth0_client`: Add support for setting `compliance_level` and `require_proof_of_possession` ([#1015](https://github.com/auth0/terraform-provider-auth0/pull/1015/))
- `resource/auth0_resource_server`: Add support for setting `consent_policy`, `authorization_details`, `token_encryption` and `proof_of_possession` ([#1015](https://github.com/auth0/terraform-provider-auth0/pull/1015/))
- `data-source/auth0_client`: Add support to retrieve `signed_request_object`, `tls_client_auth`, `compliance_level` and `require_proof_of_possession` ([#1015](https://github.com/auth0/terraform-provider-auth0/pull/1015/))

BUG FIXES:

- `resource/auth0_prompt_screen_partial`: Ensure removal of insertion points  ([#1043](https://github.com/auth0/terraform-provider-auth0/pull/1043/))


## v1.6.1

BUG FIXES:

- `resource/auth0_client`: Fixed an issue where the `default_organization` plan failed due to `ConflictsWith` and `RequiredWith` clauses in the schema ([#1021](https://github.com/auth0/terraform-provider-auth0/pull/1021))


## v1.6.0

FEATURES:

- `resource/auth0_prompt_screen_partial`: Add new resource to manage customized sign up and login experience. (1:1) ([#1013](https://github.com/auth0/terraform-provider-auth0/pull/1013))
- `resource/auth0_prompt_screen_partials`: Add new resource to manage customized sign up and login experience. (1:many) ([#1013](https://github.com/auth0/terraform-provider-auth0/pull/1013))
- `data_source/auth0_prompt_screen_partials`: Add new data source to retrieve prompt screen partials. ([#1013](https://github.com/auth0/terraform-provider-auth0/pull/1013))

ENHANCEMENTS:

- `resource/auth0_client`: Add Organizations for Client Credentials. ([#1009](https://github.com/auth0/terraform-provider-auth0/pull/1009))
- `resource/auth0_prompt_custom_text`: Add support for fetching the language list from a CDN for prompt custom text. ([#1006](https://github.com/auth0/terraform-provider-auth0/pull/1016))

BUG FIXES:

- `resource/auth0_connection`: Remove MinItems validation for precedence. ([#1017](https://github.com/auth0/terraform-provider-auth0/pull/1017))

NOTES:

- `resource/auth0_prompt_partials`: Deprecated in favor of `resource/auth0_prompt_screen_partial` and `resource/auth0_prompt_screen_partials`. ([#1013](https://github.com/auth0/terraform-provider-auth0/pull/1013))


## v1.5.0

FEATURES:

- `resource/auth0_connection`: Add new fields `attributes` and `precedence` to ConnectionOptions. ([#991](https://github.com/auth0/terraform-provider-auth0/pull/991))
- `resource/auth0_self_service_profile`: Add new resource for managing self-service profiles. ([#1008](https://github.com/auth0/terraform-provider-auth0/pull/1008))
- `data-source/auth0_self_service_profile`: Add a data source for retrieving self-service profiles. ([#1008](https://github.com/auth0/terraform-provider-auth0/pull/1008))


## v1.4.0

FEATURES:

- `resource/auth0_connection_scim_configuration`: Add a resource for managing SCIM(System for Cross-domain Identity Management) configuration. ([#980](https://github.com/auth0/terraform-provider-auth0/pull/980))
- `resource/auth0_prompt_custom_text`: Add new prompt values (`captcha`, `custom-form`, `customized-consent`, `passkeys`, `phone-identifier-challenge`, `phone-identifier-enrollment`) to the `auth0_prompt_custom_text` resource. ([#985](https://github.com/auth0/terraform-provider-auth0/pull/985))
- `data-source/auth0_connection_scim_configuration`: Add a data source for managing SCIM(System for Cross-domain Identity Management) configuration. ([#980](https://github.com/auth0/terraform-provider-auth0/pull/980))

ENHANCEMENTS:

- `resource/auth0_tenant`: Add support for `enable_sso` flag ([#972](https://github.com/auth0/terraform-provider-auth0/pull/972))

NOTES:

- `resource/auth0_tenant`: Deprecated the `require-pushed-authorization-requests` attribute. ([#986](https://github.com/auth0/terraform-provider-auth0/pull/986))

## v1.3.0

FEATURES:

- `resource/auth0_connection`: Add support for `is_signup_enabled` to support configuring sign-ups with Organization Membership ([#974](https://github.com/auth0/terraform-provider-auth0/pull/974))
- `resource/auth0_connection`: Add support for `show_as_button` to allow the button to be disabled in Organization Enterprise connections ([#974](https://github.com/auth0/terraform-provider-auth0/pull/974))
- `resource/auth0_resource_server`: Add Support for `rfc9068_profile` and `rfc9068_profile_authz` Token Dialects in Resource Server Configuration
- `data-source/auth0_organization`: Add `is_signup_enabled` to `connections` to indicate if sign-ups with Organization Membership are enabled ([#974](https://github.com/auth0/terraform-provider-auth0/pull/974))
- `data-source/auth0_organization`: Add `show_as_button` to `connections` to indicatate if button is disabled in Organization Enterprise connections ([#974](https://github.com/auth0/terraform-provider-auth0/pull/974))
 ([#973](https://github.com/auth0/terraform-provider-auth0/pull/973))

## v1.2.1

ENHANCEMENTS:

- Switched from `offset pagination` to `checkpoint pagination` when listing `organizations/members` to remove the 1000 result limit ([#965](https://github.com/auth0/terraform-provider-auth0/pull/965))

BUG FIXES:

- `resource/auth0_organization_member`: Resolve an issue where organization members were searched among a limited paginated result ([#964](https://github.com/auth0/terraform-provider-auth0/pull/964))

NOTES:

- `resource/auth0_role_permission`: Improved documentation by adding resource config and import examples ([#963](https://github.com/auth0/terraform-provider-auth0/pull/963))
- `resource/auth0_role_permissions`: Improved documentation by adding resource config and import examples ([#963](https://github.com/auth0/terraform-provider-auth0/pull/963))

## v1.2.0

FEATURES:

- `resource/auth0_prompt_partials`: Add new resource to manage prompt partials ([#918](https://github.com/auth0/terraform-provider-auth0/pull/918))


## v1.1.2

ENHANCEMENTS:

- `resource/auth0_action`: Prevent sending `secrets` and `dependencies` config if no changes triggered ([#903](https://github.com/auth0/terraform-provider-auth0/pull/903))

BUG FIXES:

- `resource/auth0_connection`: Fix PKCE values for OIDC connections ([#896](https://github.com/auth0/terraform-provider-auth0/pull/896))
- `resource/auth0_connection`: Allow `samlp` block to be defined as empty and inherit default values ([#905](https://github.com/auth0/terraform-provider-auth0/pull/905))


## v1.1.1

BUG FIXES:

- `resource/auth0_client_grant`: Allowing `scopes` to be set as empty ([#888](https://github.com/auth0/terraform-provider-auth0/pull/888))

## v1.1.0

FEATURES:

- `resource/auth0_tenant`: Add support for `customize_mfa_in_postlogin_action` setting ([#871](https://github.com/auth0/terraform-provider-auth0/pull/871))

ENHANCEMENTS:

- `resource/auth0_client`: Set default `token_endpoint_auth_method` based on `app_type` on creation ([#878](https://github.com/auth0/terraform-provider-auth0/pull/878))

BUG FIXES:

- `resource/auth0_tenant`: Set default `idle_session_lifetime` and `session_lifetime` on import ([#849](https://github.com/auth0/terraform-provider-auth0/pull/849))
- `resource/auth0_connection`: Prevent panic when checking for database config secrets ([#864](https://github.com/auth0/terraform-provider-auth0/pull/864))
- `resource/auth0_branding`: Allow deleting the resource even on free tenants ([#875](https://github.com/auth0/terraform-provider-auth0/pull/875))
- `data-source/auth0_organization`: Fix how we fetch organizations by name ([#877](https://github.com/auth0/terraform-provider-auth0/pull/877))
- `resource/auth0_connection`: Add support for `disable_self_service_change_password` on AD connection options ([#874](https://github.com/auth0/terraform-provider-auth0/pull/874))


## v1.0.0

NOTES:

- :warning: Check our [migration guide](https://github.com/auth0/terraform-provider-auth0/blob/main/MIGRATION_GUIDE.md) to navigate the breaking changes that were added in this release.
- This release supports auto generated terraform configuration. [Check our guide for more info](https://registry.terraform.io/providers/auth0/auth0/latest/docs/guides/generate_terraform_config).

FEATURES:

- `resource/auth0_connection`: Add support for Line strategy ([#818](https://github.com/auth0/terraform-provider-auth0/pull/818))
- `resource/auth0_connection`: Add `pkce` and `attribute_map` settings to OIDC and Okta Workforce connection options ([#815](https://github.com/auth0/terraform-provider-auth0/pull/815))
- `resource/auth0_client`: Add support for pushed authorization requests setting ([#756](https://github.com/auth0/terraform-provider-auth0/pull/756))
- `resource/auth0_tenant`: Add support for pushed authorization requests setting ([#756](https://github.com/auth0/terraform-provider-auth0/pull/756))
- `resource/auth0_tenant`: Add support for `allow_organization_name_in_authentication_api` toggle ([#832](https://github.com/auth0/terraform-provider-auth0/pull/832))
- `data-source/auth0_pages`: Add new data source to retrieve Auth0 pages ([#706](https://github.com/auth0/terraform-provider-auth0/pull/706))
- `data-source/auth0_signing_keys`: Add new data source to retrieve signing keys for applications ([#839](https://github.com/auth0/terraform-provider-auth0/pull/839))

ENHANCEMENTS:

- `resource/auth0_branding`: Improve validation for universal login template body ([#819](https://github.com/auth0/terraform-provider-auth0/pull/819))
- `resource/auth0_client`: Improve DX for managing `is_token_endpoint_ip_header_trusted` ([#796](https://github.com/auth0/terraform-provider-auth0/pull/796))
- `resource/auth0_log_stream`: Improve resource schema ([#798](https://github.com/auth0/terraform-provider-auth0/pull/798))
- `data-source/auth0_role`: Add ability to retrieve users assigned to a role ([#758](https://github.com/auth0/terraform-provider-auth0/pull/758))
- `resource/auth0_email_provider`: Add support for `azure_cs` and `ms365` email providers ([#752](https://github.com/auth0/terraform-provider-auth0/pull/752))
- `resource/auth0_connection`: Add support for `decryption_key` on SAML Connection Options ([#755](https://github.com/auth0/terraform-provider-auth0/pull/755))
- `resource/auth0_tenant`: Add support for `OIDCLogoutPrompt` toggle ([#754](https://github.com/auth0/terraform-provider-auth0/pull/754))
- `resource/auth0_action`: Add support for GA version of Node 18 within `runtime` field ([#709](https://github.com/auth0/terraform-provider-auth0/pull/709), [#722](https://github.com/auth0/terraform-provider-auth0/pull/722))
- `resource/auth0_trigger_actions`: Add `password-reset-post-challenge` to available trigger bindings ([#726](https://github.com/auth0/terraform-provider-auth0/pull/726))
- `resource/auth0_trigger_action`: Add `password-reset-post-challenge` to available trigger bindings ([#746](https://github.com/auth0/terraform-provider-auth0/pull/746))
- `resource/auth0_*`: Improve how we check for 404 errors ([#727](https://github.com/auth0/terraform-provider-auth0/pull/727))
- `resource/auth0_client`: Add validation to prevent empty `mobile` block ([#742](https://github.com/auth0/terraform-provider-auth0/pull/742))
- `resource/auth0_prompt_custom_text`: Add support for six new languages ([#732](https://github.com/auth0/terraform-provider-auth0/pull/732))
- `resource/auth0_connection`: Add support for `map_user_id_to_id` options field for Google Workspace connections ([#840](https://github.com/auth0/terraform-provider-auth0/pull/840))

BUG FIXES:

- `resource/auth0_prompt_custom_text`: Prevent `"null"` string literal when importing resource ([#821](https://github.com/auth0/terraform-provider-auth0/pull/821))
- `resource/auth0_guardian`: Remove minor `phone.message_types` validation ([#820](https://github.com/auth0/terraform-provider-auth0/pull/820))
- `resource/auth0_connection`: Allow sending `null` values for `from` and `messagingServiceSID` in SMS connection options ([#829](https://github.com/auth0/terraform-provider-auth0/pull/829))
- `resource/auth0_connection`: Passing-through Google Workspace admin tokens when managing the connection ([#830](https://github.com/auth0/terraform-provider-auth0/pull/830))
- `resource/auth0_tenant`: Allow creating native apps with device code grant ([#833](https://github.com/auth0/terraform-provider-auth0/pull/833))
- `resource/auth0_client_credentials`: Stop requiring `read:client_keys` permission when reading the resource ([#795](https://github.com/auth0/terraform-provider-auth0/pull/795))
- `resource/auth0_connection`: Passing-through critical connection options for following connection types: Ping Federate, AD, Azure AD, SAML ([#802](https://github.com/auth0/terraform-provider-auth0/pull/802))
- `resource/*`: Fix pagination issues when fetching resources ([#807](https://github.com/auth0/terraform-provider-auth0/pull/807))
- `data-source/*`: Fix pagination issues when fetching data sources ([#807](https://github.com/auth0/terraform-provider-auth0/pull/807))
- `resource/auth0_client_grant`: Add validation to prevent empty string scope values ([#793](https://github.com/auth0/terraform-provider-auth0/pull/793))
- `provider`: Fix issue with rate limit retry mechanism ([#788](https://github.com/auth0/terraform-provider-auth0/pull/788))
- `resource/auth0_client`: Prevent sending empty payloads for client addons ([#759](https://github.com/auth0/terraform-provider-auth0/pull/759))
- `resource/auth0_connection`: Correctly expand `shopify` connection strategy types ([#728](https://github.com/auth0/terraform-provider-auth0/pull/728))
- `data-source/auth0_*`: Raise 404 errors when retrieving data sources ([#698](https://github.com/auth0/terraform-provider-auth0/pull/698))

BREAKING CHANGES:

- `resource/auth0_client`: Formally type the `addons` sub-properties ([#655](https://github.com/auth0/terraform-provider-auth0/pull/655), [#656](https://github.com/auth0/terraform-provider-auth0/pull/656), [#657](https://github.com/auth0/terraform-provider-auth0/pull/657), [#658](https://github.com/auth0/terraform-provider-auth0/pull/658), [#661](https://github.com/auth0/terraform-provider-auth0/pull/661), [#662](https://github.com/auth0/terraform-provider-auth0/pull/662), [#663](https://github.com/auth0/terraform-provider-auth0/pull/663), [#664](https://github.com/auth0/terraform-provider-auth0/pull/664), [#665](https://github.com/auth0/terraform-provider-auth0/pull/665), [#666](https://github.com/auth0/terraform-provider-auth0/pull/666), [#667](https://github.com/auth0/terraform-provider-auth0/pull/667), [#668](https://github.com/auth0/terraform-provider-auth0/pull/668), [#669](https://github.com/auth0/terraform-provider-auth0/pull/669), [#670](https://github.com/auth0/terraform-provider-auth0/pull/670), [#671](https://github.com/auth0/terraform-provider-auth0/pull/671), [#672](https://github.com/auth0/terraform-provider-auth0/pull/672), [#673](https://github.com/auth0/terraform-provider-auth0/pull/673), [#674](https://github.com/auth0/terraform-provider-auth0/pull/674), [#675](https://github.com/auth0/terraform-provider-auth0/pull/675), [#676](https://github.com/auth0/terraform-provider-auth0/pull/676), [#677](https://github.com/auth0/terraform-provider-auth0/pull/677), [#678](https://github.com/auth0/terraform-provider-auth0/pull/678), [#681](https://github.com/auth0/terraform-provider-auth0/pull/681), [#682](https://github.com/auth0/terraform-provider-auth0/pull/682))
- `resource/auth0_user`: Remove `roles` and `permissions` fields ([#703](https://github.com/auth0/terraform-provider-auth0/pull/703))
- `resource/auth0_resource_server`: Remove `scopes` field ([#703](https://github.com/auth0/terraform-provider-auth0/pull/703))
- `resource/auth0_global_client`: Remove resource ([#704](https://github.com/auth0/terraform-provider-auth0/pull/704))
- `data-source/auth0_global_client`: Remove data source ([#704](https://github.com/auth0/terraform-provider-auth0/pull/704))
- `resource/auth0_tenant`: Remove `change_password`, `error_page` and `guardian_mfa_page` fields ([#711](https://github.com/auth0/terraform-provider-auth0/pull/711))
- `resource/auth0_tenant`: Remove `universal_login` block ([#712](https://github.com/auth0/terraform-provider-auth0/pull/712))
- `resource/auth0_client`: Remove `client_secret_rotation_trigger` field ([#711](https://github.com/auth0/terraform-provider-auth0/pull/711))
- `resource/auth0_role`: Remove `permissions` field ([#714](https://github.com/auth0/terraform-provider-auth0/pull/714))
- `resource/auth0_organization_member`: Remove `roles` field ([#715](https://github.com/auth0/terraform-provider-auth0/pull/715))
- `resource/auth0_client_grant`: Rename `scope` field to `scopes` ([#717](https://github.com/auth0/terraform-provider-auth0/pull/717))
- `resource/auth0_*`: Align resource import separators ([#718](https://github.com/auth0/terraform-provider-auth0/pull/718))
- `resource/auth0_client`: Remove `client_secret` and `token_endpoint_auth_method` fields ([#725](https://github.com/auth0/terraform-provider-auth0/pull/725))
- `resource/auth0_email`: Remove `api_user` field ([#730](https://github.com/auth0/terraform-provider-auth0/pull/730))
- `resource/auth0_connection`: Remove `enabled_clients` field ([#730](https://github.com/auth0/terraform-provider-auth0/pull/730))
- `resource/auth0_trigger_binding`: Remove resource ([#730](https://github.com/auth0/terraform-provider-auth0/pull/730))
- `resource/auth0_email`: Rename resource to `auth0_email_provider` ([#731](https://github.com/auth0/terraform-provider-auth0/pull/731))
  
## v1.0.0-beta.4

NOTES:

- This release supports auto generated terraform configuration. [Check our guide for more info](https://registry.terraform.io/providers/auth0/auth0/latest/docs/guides/generate_terraform_config).
- :warning: Check our [migration guide](https://github.com/auth0/terraform-provider-auth0/blob/main/MIGRATION_GUIDE.md) to navigate the breaking changes that were added in this release.

FEATURES:

- `resource/auth0_connection`: Add support for Line strategy ([#818](https://github.com/auth0/terraform-provider-auth0/pull/818))
- `resource/auth0_connection`: Add `pkce` and `attribute_map` settings to OIDC and Okta Workforce connection options ([#815](https://github.com/auth0/terraform-provider-auth0/pull/815))
- `resource/auth0_client`: Add support for pushed authorization requests setting ([#756](https://github.com/auth0/terraform-provider-auth0/pull/756))
- `resource/auth0_tenant`: Add support for pushed authorization requests setting ([#756](https://github.com/auth0/terraform-provider-auth0/pull/756))
- `resource/auth0_tenant`: Add support for `allow_organization_name_in_authentication_api` toggle ([#832](https://github.com/auth0/terraform-provider-auth0/pull/832))

ENHANCEMENTS:

- `resource/auth0_branding`: Improve validation for universal login template body ([#819](https://github.com/auth0/terraform-provider-auth0/pull/819))

BUG FIXES:

- `resource/auth0_prompt_custom_text`: Prevent `"null"` string literal when importing resource ([#821](https://github.com/auth0/terraform-provider-auth0/pull/821))
- `resource/auth0_guardian`: Remove minor `phone.message_types` validation  ([#820](https://github.com/auth0/terraform-provider-auth0/pull/820))
- `resource/auth0_connection`: Allow sending `null` values for `from` and `messagingServiceSID` in SMS connection options ([#829](https://github.com/auth0/terraform-provider-auth0/pull/829))
- `resource/auth0_connection`: Passing-through Google Workspace admin tokens when managing the connection ([#830](https://github.com/auth0/terraform-provider-auth0/pull/830))
- `resource/auth0_tenant`: Allow creating native apps with device code grant  ([#833](https://github.com/auth0/terraform-provider-auth0/pull/833))


## v1.0.0-beta.3

NOTES:

- :warning: Check our [migration guide](https://github.com/auth0/terraform-provider-auth0/blob/main/MIGRATION_GUIDE.md) to navigate the breaking changes that were added in this release.

ENHANCEMENTS:

- `resource/auth0_client`: Improve DX for managing `is_token_endpoint_ip_header_trusted` ([#796](https://github.com/auth0/terraform-provider-auth0/pull/796))
- `resource/auth0_log_stream`: Improve resource schema ([#798](https://github.com/auth0/terraform-provider-auth0/pull/798))

BUG FIXES:

- `resource/auth0_client_credentials`: Stop requiring `read:client_keys` permission when reading the resource ([#795](https://github.com/auth0/terraform-provider-auth0/pull/795))
- `resource/auth0_connection`: Passing-through critical connection options for following connection types: Ping Federate, AD, Azure AD, SAML ([#802](https://github.com/auth0/terraform-provider-auth0/pull/802))
- `resource/*`: Fix pagination issues when fetching resources ([#807](https://github.com/auth0/terraform-provider-auth0/pull/807))
- `data-source/*`: Fix pagination issues when fetching data sources ([#807](https://github.com/auth0/terraform-provider-auth0/pull/807))


## v1.0.0-beta.2

NOTES:

- :warning: Check our [migration guide](https://github.com/auth0/terraform-provider-auth0/blob/main/MIGRATION_GUIDE.md) to navigate the breaking changes that were added in this release.

BUG FIXES:

- `resource/auth0_client_grant`: Add validation to prevent empty string scope values ([#793](https://github.com/auth0/terraform-provider-auth0/pull/793))
- `provider`: Fix issue with rate limit retry mechanism ([#788](https://github.com/auth0/terraform-provider-auth0/pull/788))


## v1.0.0-beta.1

NOTES:

- :warning: Check our [migration guide](https://github.com/auth0/terraform-provider-auth0/blob/main/MIGRATION_GUIDE.md) to navigate the breaking changes that were added in this release.

ENHANCEMENTS:

- `data-source/auth0_role`: Add ability to retrieve users assigned to a role ([#758](https://github.com/auth0/terraform-provider-auth0/pull/758))
- `resource/auth0_email_provider`: Add support for `azure_cs` and `ms365` email providers ([#752](https://github.com/auth0/terraform-provider-auth0/pull/752))
- `resource/auth0_connection`: Add support for `decryption_key` on SAML Connection Options ([#755](https://github.com/auth0/terraform-provider-auth0/pull/755))
- `resource/auth0_tenant`: Add support for `OIDCLogoutPrompt` toggle ([#754](https://github.com/auth0/terraform-provider-auth0/pull/754))

BUG FIXES:

- `resource/auth0_client`: Prevent sending empty payloads for client addons ([#759](https://github.com/auth0/terraform-provider-auth0/pull/759))


## v1.0.0-beta.0

NOTES:

- :warning: Check our [migration guide](https://github.com/auth0/terraform-provider-auth0/blob/main/MIGRATION_GUIDE.md) to navigate the breaking changes that were added in this release.

FEATURES:

- `data-source/auth0_pages`: Add new data source to retrieve Auth0 pages ([#706](https://github.com/auth0/terraform-provider-auth0/pull/706))

ENHANCEMENTS:

- `resource/auth0_action`: Add support for GA version of Node 18 within `runtime` field ([#709](https://github.com/auth0/terraform-provider-auth0/pull/709), [#722](https://github.com/auth0/terraform-provider-auth0/pull/722))
- `resource/auth0_trigger_actions`: Add `password-reset-post-challenge` to available trigger bindings ([#726](https://github.com/auth0/terraform-provider-auth0/pull/726))
- `resource/auth0_trigger_action`: Add `password-reset-post-challenge` to available trigger bindings ([#746](https://github.com/auth0/terraform-provider-auth0/pull/746))
- `resource/auth0_*`: Improve how we check for 404 errors ([#727](https://github.com/auth0/terraform-provider-auth0/pull/727))
- `resource/auth0_client`: Add validation to prevent empty `mobile` block ([#742](https://github.com/auth0/terraform-provider-auth0/pull/742))
- `resource/auth0_prompt_custom_text`: Add support for six new languages ([#732](https://github.com/auth0/terraform-provider-auth0/pull/732))

BUG FIXES:

- `resource/auth0_connection`: Correctly expand `shopify` connection strategy types ([#728](https://github.com/auth0/terraform-provider-auth0/pull/728))
- `data-source/auth0_*`: Raise 404 errors when retrieving data sources ([#698](https://github.com/auth0/terraform-provider-auth0/pull/698))

BREAKING CHANGES:

- `resource/auth0_client`: Formally type the `addons` sub-properties ([#655](https://github.com/auth0/terraform-provider-auth0/pull/655), [#656](https://github.com/auth0/terraform-provider-auth0/pull/656), [#657](https://github.com/auth0/terraform-provider-auth0/pull/657), [#658](https://github.com/auth0/terraform-provider-auth0/pull/658), [#661](https://github.com/auth0/terraform-provider-auth0/pull/661), [#662](https://github.com/auth0/terraform-provider-auth0/pull/662), [#663](https://github.com/auth0/terraform-provider-auth0/pull/663), [#664](https://github.com/auth0/terraform-provider-auth0/pull/664), [#665](https://github.com/auth0/terraform-provider-auth0/pull/665), [#666](https://github.com/auth0/terraform-provider-auth0/pull/666), [#667](https://github.com/auth0/terraform-provider-auth0/pull/667), [#668](https://github.com/auth0/terraform-provider-auth0/pull/668), [#669](https://github.com/auth0/terraform-provider-auth0/pull/669), [#670](https://github.com/auth0/terraform-provider-auth0/pull/670), [#671](https://github.com/auth0/terraform-provider-auth0/pull/671), [#672](https://github.com/auth0/terraform-provider-auth0/pull/672), [#673](https://github.com/auth0/terraform-provider-auth0/pull/673), [#674](https://github.com/auth0/terraform-provider-auth0/pull/674), [#675](https://github.com/auth0/terraform-provider-auth0/pull/675), [#676](https://github.com/auth0/terraform-provider-auth0/pull/676), [#677](https://github.com/auth0/terraform-provider-auth0/pull/677), [#678](https://github.com/auth0/terraform-provider-auth0/pull/678), [#681](https://github.com/auth0/terraform-provider-auth0/pull/681), [#682](https://github.com/auth0/terraform-provider-auth0/pull/682))
- `resource/auth0_user`: Remove `roles` and `permissions` fields ([#703](https://github.com/auth0/terraform-provider-auth0/pull/703))
- `resource/auth0_resource_server`: Remove `scopes` field ([#703](https://github.com/auth0/terraform-provider-auth0/pull/703))
- `resource/auth0_global_client`: Remove resource ([#704](https://github.com/auth0/terraform-provider-auth0/pull/704))
- `data-source/auth0_global_client`: Remove data source ([#704](https://github.com/auth0/terraform-provider-auth0/pull/704))
- `resource/auth0_tenant`: Remove `change_password`, `error_page` and `guardian_mfa_page` fields ([#711](https://github.com/auth0/terraform-provider-auth0/pull/711))
- `resource/auth0_tenant`: Remove `universal_login` block ([#712](https://github.com/auth0/terraform-provider-auth0/pull/712))
- `resource/auth0_client`: Remove `client_secret_rotation_trigger` field ([#711](https://github.com/auth0/terraform-provider-auth0/pull/711))
- `resource/auth0_role`: Remove `permissions` field ([#714](https://github.com/auth0/terraform-provider-auth0/pull/714))
- `resource/auth0_organization_member`: Remove `roles` field ([#715](https://github.com/auth0/terraform-provider-auth0/pull/715))
- `resource/auth0_client_grant`: Rename `scope` field to `scopes` ([#717](https://github.com/auth0/terraform-provider-auth0/pull/717))
- `resource/auth0_*`: Align resource import separators ([#718](https://github.com/auth0/terraform-provider-auth0/pull/718))
- `resource/auth0_client`: Remove `client_secret` and `token_endpoint_auth_method` fields ([#725](https://github.com/auth0/terraform-provider-auth0/pull/725))
- `resource/auth0_email`: Remove `api_user` field ([#730](https://github.com/auth0/terraform-provider-auth0/pull/730))
- `resource/auth0_connection`: Remove `enabled_clients` field ([#730](https://github.com/auth0/terraform-provider-auth0/pull/730))
- `resource/auth0_trigger_binding`: Remove resource ([#730](https://github.com/auth0/terraform-provider-auth0/pull/730))
- `resource/auth0_email`: Rename resource to `auth0_email_provider` ([#731](https://github.com/auth0/terraform-provider-auth0/pull/731))


## 0.50.2

ENHANCEMENTS:

- `resource/auth0_action`: Add node18 GA (`node18-actions`) option to `runtime`([#803](https://github.com/auth0/terraform-provider-auth0/pull/803))


## 0.50.1

BUG FIXES:

- `resource/auth0_connection`: Passing-through critical connection options for following connection types: Ping Federate, AD, Azure AD, SAML([#786](https://github.com/auth0/terraform-provider-auth0/pull/786))


## 0.50.0

FEATURES:

- `resource/auth0_pages`: Add new resource to manage Auth0 pages (`change_password`, `error`, `guardian_mfa`, `login`) ([#691](https://github.com/auth0/terraform-provider-auth0/pull/691))

ENHANCEMENTS: 

- `resource/auth0_client`: Add `post_login_prompt` to available options for the `organization_require_behavior` attribute ([#680](https://github.com/auth0/terraform-provider-auth0/pull/680))
- `resource/auth0_connection`: Relax `metadata` validation by not requiring key length to be between 0 and 10 characters ([#685](https://github.com/auth0/terraform-provider-auth0/pull/685))

BUG FIXES:

- `resource/auth0_organization_connections`, `resource/auth0_organization_members`: Address a bug causing inconsistencies in the safeguarding process, ensuring reliable protection against erasing unintended modifications ([#645](https://github.com/auth0/terraform-provider-auth0/pull/645))
- `resource/auth0_organization_members`: Address a bug that prevented the creation of organization members when the member list was empty ([#646](https://github.com/auth0/terraform-provider-auth0/pull/646))
- `resource/auth0_connection`, `resource/auth0_organization_member`,`resource/auth0_user`: Update diffing algorithm to address a bug where the order of additions and removals was causing incorrect results ([#650](https://github.com/auth0/terraform-provider-auth0/pull/650))
- `resource/auth0_connection`: Remove invalid connection strategies ([#694](https://github.com/auth0/terraform-provider-auth0/pull/694))
- `resource/auth0_client`: Modify the behavior to only allow the update of the `is_token_endpoint_ip_header_trusted` setting after the client has been created successfully ([#696](https://github.com/auth0/terraform-provider-auth0/pull/696))
- `resource/auth0_branding`: Addressed a bug that prevented the deletion of the template when the universal login block was removed ([#695](https://github.com/auth0/terraform-provider-auth0/pull/695))

NOTES:

- :warning: Check our [migration guide](https://github.com/auth0/terraform-provider-auth0/blob/main/MIGRATION_GUIDE.md) to navigate the deprecations that were added in this release.


## 0.49.0

FEATURES:

- `data-source/auth0_organization`: Add members ([#615](https://github.com/auth0/terraform-provider-auth0/pull/615))
- `resource/auth0_organization_connections`: Add new resource to manage a 1:many relationship between an organization and its enabled connections ([#610](https://github.com/auth0/terraform-provider-auth0/pull/610))
- `resource/auth0_organization_members`: Add new resource to manage a 1:many relationship between an organization and its members ([#614](https://github.com/auth0/terraform-provider-auth0/pull/614))
- `resource/auth0_organization_member_role`: Add new resource to manage a 1:1 relationship between an organization member and its roles ([#622](https://github.com/auth0/terraform-provider-auth0/pull/622))
- `resource/auth0_organization_member_roles`: Add new resource to manage a 1:many relationship between an organization member and its roles ([#617](https://github.com/auth0/terraform-provider-auth0/pull/617))
- `resource/auth0_trigger_action`: Add new resource to manage a 1:1 relationship between a trigger binding and an action ([#612](https://github.com/auth0/terraform-provider-auth0/pull/612), [#621](https://github.com/auth0/terraform-provider-auth0/pull/621))
- `resource/auth0_trigger_actions`: Add new resource to manage a 1:many relationship between a trigger binding and actions ([#613](https://github.com/auth0/terraform-provider-auth0/pull/613))

BUG FIXES:

- `resource/auth0_client_credentials`: Correctly set ID when importing ([#608](https://github.com/auth0/terraform-provider-auth0/pull/608))
- `resource/auth0_connection`: More consistent `set_user_root_attributes` behavior for enterprise connections ([#619](https://github.com/auth0/terraform-provider-auth0/pull/619))
- `resource/auth0_user_role`: Enable importing of resource ([#629](https://github.com/auth0/terraform-provider-auth0/pull/629))
- `resource/auth0_user_permissions`: Update diffing algorithm to address a bug where the order of additions and removals was causing incorrect results ([#630](https://github.com/auth0/terraform-provider-auth0/pull/630))
- `resource/auth0_role_permissions`: Update diffing algorithm to address a bug where the order of additions and removals was causing incorrect results ([#632](https://github.com/auth0/terraform-provider-auth0/pull/632))
- `resource/auth0_trigger_action`: Fix delete logic ([#639](https://github.com/auth0/terraform-provider-auth0/pull/639))

NOTES: 

- :warning: Check our [migration guide](https://github.com/auth0/terraform-provider-auth0/blob/main/MIGRATION_GUIDE.md)
to navigate the deprecations that were added in this release.


## 0.48.0

FEATURES:

- `resource/auth0_client_credentials`: Add new resource to manage client credentials (`client_secret`, `private_key_jwt`, `authentication_methods`) ([#588](https://github.com/auth0/terraform-provider-auth0/pull/588))
- `resource/auth0_resource_server_scopes`: Add new resource to manage a 1:many relationship between the resource server (API) and its scopes (permissions) ([#600](https://github.com/auth0/terraform-provider-auth0/pull/600))
- `resource/auth0_resource_server_scope`: Add new resource to manage a 1:1 relationship between the resource server (API) and its scopes (permissions) ([#589](https://github.com/auth0/terraform-provider-auth0/pull/589))

BUG FIXES:

- `resource/auth0_connection`: Fix json tag for `forward_request_info` attribute ([#591](https://github.com/auth0/terraform-provider-auth0/pull/591))
- Fix import issue on several resources (`auth0_connection_clients`, `auth0_user_permissions`, `auth0_user_roles`, `auth0_role_permissions`) ([#594](https://github.com/auth0/terraform-provider-auth0/pull/594), [#595](https://github.com/auth0/terraform-provider-auth0/pull/595), [#596](https://github.com/auth0/terraform-provider-auth0/pull/596), [#597](https://github.com/auth0/terraform-provider-auth0/pull/597))
- `resource/auth0_connection`: Fix issue with setting `set_user_root_attributes` to `on_each_login` for Microsoft Azure AD Connections ([#602](https://github.com/auth0/terraform-provider-auth0/pull/602))

NOTES:

- New guides on how to achieve 0 downtime client credentials were added in this release ([#592](https://github.com/auth0/terraform-provider-auth0/pull/592))
- :warning: Check our [migration guide](https://github.com/auth0/terraform-provider-auth0/blob/main/MIGRATION_GUIDE.md)
to navigate the deprecations that were added in this release.


## 0.47.0

FEATURES:

- `resource/auth0_connection_clients`: Add new resource to manage a 1:many relationship between the connection and its enabled clients ([#568](https://github.com/auth0/terraform-provider-auth0/pull/568))
- `resource/auth0_user_permission`: Add new resource to manage a 1:1 relationship between the user and its permissions ([#574](https://github.com/auth0/terraform-provider-auth0/pull/574))
- `resource/auth0_user_permissions`: Add new resource to manage a 1:many relationship between the user and its permissions ([#578](https://github.com/auth0/terraform-provider-auth0/pull/578))
- `resource/auth0_user_role`: Add new resource to manage a 1:1 relationship between the user and its roles ([#580](https://github.com/auth0/terraform-provider-auth0/pull/580))
- `resource/auth0_user_roles`: Add new resource to manage a 1:many relationship between the user and its roles ([#579](https://github.com/auth0/terraform-provider-auth0/pull/579))
- `resource/auth0_role_permission`: Add new resource to manage a 1:1 relationship between the role and its permissions ([#582](https://github.com/auth0/terraform-provider-auth0/pull/582))
- `resource/auth0_role_permissions`: Add new resource to manage a 1:many relationship between the role and its permissions ([#583](https://github.com/auth0/terraform-provider-auth0/pull/583))
- `resource/auth0_user`: Add new readonly `permissions` attribute ([#572](https://github.com/auth0/terraform-provider-auth0/pull/572))
- `resource/auth0_client`: Add OIDC Back-Channel Logout support ([#581](https://github.com/auth0/terraform-provider-auth0/pull/581))
- `resource/auth0_role`: Add `description` and `resource_server_name` read-only fields to `permissions` ([#581](https://github.com/auth0/terraform-provider-auth0/pull/581))

BUG FIXES:

- Fix created import ID on association resource ([#569](https://github.com/auth0/terraform-provider-auth0/pull/569))

NOTES:

- :warning: Check our [migration guide](https://github.com/auth0/terraform-provider-auth0/blob/main/MIGRATION_GUIDE.md)
to navigate the deprecations that were added in this release.


## 0.46.0

BUG FIXES:

- `resource/auth0_resource_server`: Remove invalid `options` attribute from schema ([#551](https://github.com/auth0/terraform-provider-auth0/pull/551))
- `resource/auth0_trigger_binding`: Fix `trigger` import issue ([#554](https://github.com/auth0/terraform-provider-auth0/pull/554))
- `data-source/auth0_resource_server`: Fix auth0 management api data source not reading `scopes` ([#555](https://github.com/auth0/terraform-provider-auth0/pull/555))
- `resource/auth0_connection`: Fix faulty diffs when setting the `metadata_xml` on a SAML connection ([#559](https://github.com/auth0/terraform-provider-auth0/pull/559))
- `resource/auth0_connection_client`: Stop overriding internally the imported ID for this resource ([#562](https://github.com/auth0/terraform-provider-auth0/pull/562))
- `resource/auth0_organization_connection`: Stop overriding internally the imported ID for this resource ([#562](https://github.com/auth0/terraform-provider-auth0/pull/562))
- `resource/auth0_organization_member`: Stop overriding internally the imported ID for this resource ([#562](https://github.com/auth0/terraform-provider-auth0/pull/562))

FEATURES:

- `resource/auth0_guardian`: Add support for `direct` provider within `push` MFA ([#535](https://github.com/auth0/terraform-provider-auth0/pull/535))

ENHANCEMENTS:

- `resource/auth0_tenant`: Add support for `mfa_show_factor_list_on_enrollment` flag ([#561](https://github.com/auth0/terraform-provider-auth0/pull/561))

NOTES:

- :warning: The removal of the `options` attribute from the `auth0_resource_server` resource, while technically a breaking change,
should not cause any issues as the API wasn't accepting this parameter.


## 0.45.0

BUG FIXES:

- `resource/auth0_branding_theme`: Fixed typo in `fonts.links_style` validation ([#523](https://github.com/auth0/terraform-provider-auth0/pull/523))
- `data-source/auth0_resource_server`: Fixed data source to always return the id instead of the identifier ([#532](https://github.com/auth0/terraform-provider-auth0/pull/532))

FEATURES:

- `data-source/auth0_custom_domain`: Added data source to fetch custom domain ([#526](https://github.com/auth0/terraform-provider-auth0/pull/526))
- `resource/auth0_connection`: Added support for ping federate connections ([#527](https://github.com/auth0/terraform-provider-auth0/pull/527))

ENHANCEMENTS:

- `resource/auth0_client_grant`: Check if client grant already exists before creating ([#529](https://github.com/auth0/terraform-provider-auth0/pull/529))
- `resource/auth0_connection`: Added `disable_self_service_change_password` flag to database connection ([#525](https://github.com/auth0/terraform-provider-auth0/pull/525))

NOTES:

- Updated docs for `auth0_role` resource ([#524](https://github.com/auth0/terraform-provider-auth0/pull/524))


## 0.44.1

BUG FIXES:

- `resource/auth0_prompt_custom_text`: Added missing status prompt type ([#513](https://github.com/auth0/terraform-provider-auth0/pull/513))
- `data-source/auth0_connection`: Moved check for config secrets from the read to the update func ([#517](https://github.com/auth0/terraform-provider-auth0/pull/517))

ENHANCEMENTS:

- `resource/auth0_branding_theme`: Made fields optional ([#499](https://github.com/auth0/terraform-provider-auth0/pull/499))

NOTES:

- Added docs on available log stream types ([#462](https://github.com/auth0/terraform-provider-auth0/pull/462))
- Added docs on how to obtain the custom domain id for importing `auth0_custom_domain` resources ([#463](https://github.com/auth0/terraform-provider-auth0/pull/463))


## 0.44.0

BUG FIXES:

- `resource/auth0_prompt_custom_text`: Added missing prompt types ([#506](https://github.com/auth0/terraform-provider-auth0/pull/506))
- `resource/auth0_branding`: Fixed resource to allow managing only the universal login ([#506](https://github.com/auth0/terraform-provider-auth0/pull/506))

FEATURES:

- `data-source/auth0_connection`: Added data source to fetch connection information ([#470](https://github.com/auth0/terraform-provider-auth0/pull/470))
- `data-source/auth0_resource_server`: Added data source to fetch resource server information ([#477](https://github.com/auth0/terraform-provider-auth0/pull/477))
- `data-source/auth0_organization`: Added data source to fetch organization information ([#475](https://github.com/auth0/terraform-provider-auth0/pull/475))
- `data-source/auth0_tenant`: Expanded data source to fetch all the tenant information ([#479](https://github.com/auth0/terraform-provider-auth0/pull/479))
- `data-source/auth0_user`: Added data source to fetch user information ([#481](https://github.com/auth0/terraform-provider-auth0/pull/481))
- `data-source/auth0_role`: Added data source to fetch role information ([#483](https://github.com/auth0/terraform-provider-auth0/pull/483))
- `data-source/auth0_attack_protection`: Added data source to fetch attack protection information ([#485](https://github.com/auth0/terraform-provider-auth0/pull/485))
- `data-source/auth0_branding`: Added data source to fetch branding information ([#500](https://github.com/auth0/terraform-provider-auth0/pull/500))
- `data-source/auth0_branding_theme`: Added data source to fetch branding theme information ([#500](https://github.com/auth0/terraform-provider-auth0/pull/500))
- `resource/auth0_branding_theme`: Simplified management of this resource to no longer force a user to import the resource if already existing ([#504](https://github.com/auth0/terraform-provider-auth0/pull/504))

NOTES:

- `resource/auth0_connection`: Updated connection docs ([#471](https://github.com/auth0/terraform-provider-auth0/pull/471))
- `resource/auth0_tenant`: Added deprecation notice to `flags.universal_login` ([#503](https://github.com/auth0/terraform-provider-auth0/pull/503))


## 0.43.0

BUG FIXES:

- `resource/auth0_guardian`: Allow updating message templates for `phone-message-hook` sms provider ([#444](https://github.com/auth0/terraform-provider-auth0/pull/444))
- `resource/auth0_branding`: Manage universal-login body only if custom domains are set ([#446](https://github.com/auth0/terraform-provider-auth0/pull/446))
- `resource/auth0_connection`: Set `authorization_endpoint`, `issuer`, `jws_uri`, `token_endpoint`, `user_info_endpoint` as `Computed` to prevent diff issues ([#443](https://github.com/auth0/terraform-provider-auth0/pull/443))
- `resource/auth0_user`: Only send changed fields when updating a user ([#453](https://github.com/auth0/terraform-provider-auth0/pull/453))
- `resource/auth0_*`: Relax url schema validation for `initiate_login_uri`, `apple_app_link`, `google_app_link`, `default_redirection_uri` to allow to be set to empty ([#453](https://github.com/auth0/terraform-provider-auth0/pull/453))


FEATURES:

- `resource/auth0_log_stream`: Added support for segment log stream type ([#437](https://github.com/auth0/terraform-provider-auth0/pull/437))
- `resource/auth0_action`: Added `node18` to runtime options ([#442](https://github.com/auth0/terraform-provider-auth0/pull/442))
- `resource/auth0_attack_protection`: Added support for `breached_password_detection.stage` ([#445](https://github.com/auth0/terraform-provider-auth0/pull/445))
- `resource/auth0_connection`: Added support for `enable_script_context` to db connections ([#452](https://github.com/auth0/terraform-provider-auth0/pull/452))
- `resource/auth0_connection`: Added support for `fed_metadata_xml` to adfs connections ([#458](https://github.com/auth0/terraform-provider-auth0/pull/458))
- `resource/auth0_connection`: Added support for `icon_url` to okta connections ([#459](https://github.com/auth0/terraform-provider-auth0/pull/459))

NOTES:

- `resource/auth0_client`: Updated `client_secret_rotation_trigger` docs ([#439](https://github.com/auth0/terraform-provider-auth0/pull/439))
- `resource/auth0_client`: Updated `cross_origin_auth` docs ([#456](https://github.com/auth0/terraform-provider-auth0/pull/456))


## 0.42.0

BUG FIXES:

- `resource/auth0_connection_client`, `resource/auth0_organization_connection`, `resource/auth0_organization_member`:
Fixed concurrency issue ([#425](https://github.com/auth0/terraform-provider-auth0/pull/425))

BREAKING CHANGES:

- `resource/auth0_guardian`: Redesigned how mfa types get enabled ([#423](https://github.com/auth0/terraform-provider-auth0/pull/423))
  - It is now necessary to explicitly set the enabled attribute on all mfa types. Please check the [auth0_guardian](https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/guardian) docs for more info.


## 0.41.0

BREAKING CHANGES:

- `resource/auth0_connection`: Removed `enabled_clients` field ([#421](https://github.com/auth0/terraform-provider-auth0/pull/421))
  - It is necessary to migrate all `enabled_clients` on the `auth0_connection` resource to the [auth0_connection_client](https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/connection_client) resource.


## 0.40.1

FEATURES:

- `resource/auth0_log_stream`: Added support for Mixpanel log streams ([#408](https://github.com/auth0/terraform-provider-auth0/pull/408))

BUG FIXES:

- `resource/auth0_guardian`: Added `provider` to `push` mfa to fix `sns` settings not getting propagated ([#415](https://github.com/auth0/terraform-provider-auth0/pull/415))
- Prevent panic on null values when iterating over map elements ([#413](https://github.com/auth0/terraform-provider-auth0/pull/413))


## 0.40.0

FEATURES:

- `resource/auth0_connection_client`: Added new resource to manage enabled clients on a connection ([#379](https://github.com/auth0/terraform-provider-auth0/pull/379))
- `resource/auth0_connection`: Added support for `okta` connection type ([#395](https://github.com/auth0/terraform-provider-auth0/pull/395))
- `resource/auth0_email`: Added `settings` field ([#394](https://github.com/auth0/terraform-provider-auth0/pull/394))

ENHANCEMENTS:

- `resource/auth0_connection`: Added documentation for connection options for all strategy types ([#383](https://github.com/auth0/terraform-provider-auth0/pull/383))
- `resource/auth0_client`: Relaxed validation rules on initiate_login_uri to match API ([#389](https://github.com/auth0/terraform-provider-auth0/pull/389))
- `resource/auth0_email`: Update email provider if already existing when creating it ([#377](https://github.com/auth0/terraform-provider-auth0/pull/377))
- `resource/auth0_email`: Added validation to all attributes ([#391](https://github.com/auth0/terraform-provider-auth0/pull/391))

NOTES:

- `resource/auth0_email`: Deprecated `api_user` field ([#392](https://github.com/auth0/terraform-provider-auth0/pull/392))


## 0.39.0

FEATURES:

- `resource/auth0_client`: Added `addons.samlp.issuer` ([#334](https://github.com/auth0/terraform-provider-auth0/pull/334))
- `resource/auth0_client`: Added `client_aliases` ([#367](https://github.com/auth0/terraform-provider-auth0/pull/367))
- `resource/auth0_custom_domain`: Added `custom_client_ip_header` and `tls_policy` ([#335](https://github.com/auth0/terraform-provider-auth0/pull/335))

BUG FIXES:

- `resource/auth0_branding`: Ignored 404 error when fetching universal login content ([#359](https://github.com/auth0/terraform-provider-auth0/pull/359))
- `resource/auth0_branding_theme`: Improved precision of fields with `float64` instead of `int` ([#369](https://github.com/auth0/terraform-provider-auth0/pull/369))
- `resource/auth0_resource_server`: Fixed managing auth0 management api ([#374](https://github.com/auth0/terraform-provider-auth0/pull/374))
- `resource/auth0_client`: Fixed update behavior of `client_metadata` ([#362](https://github.com/auth0/terraform-provider-auth0/pull/362))
- `resource/auth0_connection`: Added validation on `identity_api` for `waad` connections ([#361](https://github.com/auth0/terraform-provider-auth0/pull/361))

NOTES:

- `resource/auth0_resource_server`: Improved RBAC docs ([#371](https://github.com/auth0/terraform-provider-auth0/pull/371))
- `resource/auth0_action`: Added guide on how to retrieve available action triggers ([#370](https://github.com/auth0/terraform-provider-auth0/pull/370))
- `resource/auth0_prompt_custom_text`: Escaped dollar sign references in docs ([#366](https://github.com/auth0/terraform-provider-auth0/pull/366))

## 0.38.0

This release focuses primarily on setting fields to empty values consistently across all resources.
For an in depth explanation please check: [#14](https://github.com/auth0/terraform-provider-auth0/issues/14#issuecomment-1271345897)

BUG FIXES:

- Allowed setting fields to empty consistently across all resources ([#354](https://github.com/auth0/terraform-provider-auth0/pull/354))
- Correctly destroy resources by setting the ID to blank ([#354](https://github.com/auth0/terraform-provider-auth0/pull/354))
- Stop ignoring non 404 errors when deleting resources ([#354](https://github.com/auth0/terraform-provider-auth0/pull/354))
- `resource/auth0_prompt`: Set `universal_login_experience` and `webauthn_platform_first_factor` to `Computed` ([#354](https://github.com/auth0/terraform-provider-auth0/pull/354))
- `resource/auth0_resource_server`: Set `skip_consent_for_verifiable_first_party_clients` and `enforce_policies` to `Computed` ([#354](https://github.com/auth0/terraform-provider-auth0/pull/354))
- `resource/auth0_rule`: Set `enabled` to `Computed` ([#354](https://github.com/auth0/terraform-provider-auth0/pull/354))

BREAKING CHANGES:

- `resource/auth0_organization`: Removed deprecated `connections` field ([#354](https://github.com/auth0/terraform-provider-auth0/pull/354))
  - Please migrate all managed `connections` through the `auth0_organization` resource to the `auth0_organization_connection` resource.

NOTES:

- `resource/auth0_resource_server`: Changed `identifier` from `Optional` to `Required` ([#354](https://github.com/auth0/terraform-provider-auth0/pull/354))


## 0.37.1

BUG FIXES:

- `resource/auth0_client`: Fix how we expand `addons.samlp` ([#322](https://github.com/auth0/terraform-provider-auth0/pull/322))

NOTES:

- `resource/auth0_client`: Improve description of `app_type` attribute ([#325](https://github.com/auth0/terraform-provider-auth0/pull/325))


## 0.37.0

FEATURES:

- `resource/auth0_connection`: Prevented erasing `options.configuration` by mistake ([#307](https://github.com/auth0/terraform-provider-auth0/pull/307))

BUG FIXES:

- `resource/auth0_organization_connection`: Fixed issue with importing ([#301](https://github.com/auth0/terraform-provider-auth0/pull/301))
- `resource/auth0_organization_member`: Fixed issue with importing ([#302](https://github.com/auth0/terraform-provider-auth0/pull/302))
- `resource/auth0_connection`: Added missing field `set_user_root_attributes` to the auth0 connection ([#303](https://github.com/auth0/terraform-provider-auth0/pull/303))
- `data-source/auth0_client`: Fixed search by name through all available clients ([#306](https://github.com/auth0/terraform-provider-auth0/pull/306))
- `resource/auth0_email`: Refactored and removed `ForceNew` on secret fields ([#304](https://github.com/auth0/terraform-provider-auth0/pull/304))

NOTES:

- `resource/auth0_prompt`: Refactored and added additional test cases ([#305](https://github.com/auth0/terraform-provider-auth0/pull/305))
- Upgraded test recordings to go-vcr v3 ([#309](https://github.com/auth0/terraform-provider-auth0/pull/309))
- Removed unnecessary `MapData` struct from resource data helpers ([#310](https://github.com/auth0/terraform-provider-auth0/pull/310))


## 0.36.0

FEATURES:

- `resource/auth0_branding_theme`: Add new resource to manage branding themes ([#292](https://github.com/auth0/terraform-provider-auth0/pull/292))
- `provider`: Add ability to pass a custom audience when using client credentials flow ([#295](https://github.com/auth0/terraform-provider-auth0/pull/295))

NOTES:

- `auth0_action`: Improve `supported_triggers.version` description ([#287](https://github.com/auth0/terraform-provider-auth0/pull/287))
- `auth0_connection`: Improve `options.scopes` description ([#297](https://github.com/auth0/terraform-provider-auth0/pull/297))


## 0.35.0

FEATURES:

- `resource/auth0_action`: Throw error when encountering untracked action secrets ([#248](https://github.com/auth0/terraform-provider-auth0/pull/248))

NOTES:

- Reorganized project layout ([#262](https://github.com/auth0/terraform-provider-auth0/pull/262))
- Updated documentation and examples


## 0.34.0

FEATURES:

- `resource/auth0_prompt`: Added `webauthn_platform_first_factor` field ([#237](https://github.com/auth0/terraform-provider-auth0/pull/237))
- `resource/auth0_connection`: Added `auth_params` for passwordless email connections ([#235](https://github.com/auth0/terraform-provider-auth0/pull/235), [#240](https://github.com/auth0/terraform-provider-auth0/pull/240), [#241](https://github.com/auth0/terraform-provider-auth0/pull/241))
- `resource/auth0_connection`: Added support for multiple OAuth2 compatible strategies ([#239](https://github.com/auth0/terraform-provider-auth0/pull/235))
- `resource/auth0_organization_member`: Added new resource to manage organization members and their roles ([#256](https://github.com/auth0/terraform-provider-auth0/pull/256))
- `resource/auth0_organization_connection`: Added new resource to manage organization connections ([#253](https://github.com/auth0/terraform-provider-auth0/pull/253))

BUG FIXES:

- `resource/auth0_organization`: Fixed issue with not being able to update `connections` ([#244](https://github.com/auth0/terraform-provider-auth0/pull/244))
- `resource/auth0_organization`: Fixed issue with `metadata` field not getting set to empty ([#245](https://github.com/auth0/terraform-provider-auth0/pull/245), [#254](https://github.com/auth0/terraform-provider-auth0/pull/254))
- `resource/auth0_action`: Fix issue with not being able to update `dependencies` ([#247](https://github.com/auth0/terraform-provider-auth0/pull/247))
- `resource/auth0_user`: Fix infinite plan on `user_metadata` ([#249](https://github.com/auth0/terraform-provider-auth0/pull/249), [#250](https://github.com/auth0/terraform-provider-auth0/pull/250))


## 0.33.0

FEATURES:

- `resource/auth0_guardian`: Added webauthn MFA ([#213](https://github.com/auth0/terraform-provider-auth0/pull/213))
- `resource/auth0_guardian`: Added duo MFA ([#214](https://github.com/auth0/terraform-provider-auth0/pull/214))
- `resource/auth0_guardian`: Added push (Amazon SNS, custom app) MFA ([#215](https://github.com/auth0/terraform-provider-auth0/pull/215))
- `resource/auth0_guardian`: Added recovery code ([#216](https://github.com/auth0/terraform-provider-auth0/pull/216))
- `resource/auth0_tenant`: Added `session_cookie` field ([#220](https://github.com/auth0/terraform-provider-auth0/pull/220))
- `resource/auth0_client`: Added `sso_integration` as valid app type ([#221](https://github.com/auth0/terraform-provider-auth0/pull/221))
- `resource/auth0_email_template`: Added `include_email_in_redirect` field ([#229](https://github.com/auth0/terraform-provider-auth0/pull/229))
- `resource/auth0_connection`: Added `upstream_params` field ([#223](https://github.com/auth0/terraform-provider-auth0/pull/223))
  

## 0.32.0

FEATURES:

- `resource/auth0_connection`: Added `disable_sign_out` field to samlp connections ([#204](https://github.com/auth0/terraform-provider-auth0/pull/204))
- `resource/auth0_connection`: Added `metadata_xml` and `metadata_url` to connection schema ([#204](https://github.com/auth0/terraform-provider-auth0/pull/204))
- `resource/auth0_connection`: Added `signing_key` to samlp connections ([#210](https://github.com/auth0/terraform-provider-auth0/pull/210))
- `resource/auth0_connection`: Adding `pkce_enabled` field to Oauth2 connection options ([#212](https://github.com/auth0/terraform-provider-auth0/pull/212))
- `resource/auth0_tenant`: Added several missing tenant flags ([#208](https://github.com/auth0/terraform-provider-auth0/pull/208))

BREAKING CHANGES:

- `resource/auth0_connection`: Removed deprecated `options.app_domain` in connection resource ([#202](https://github.com/auth0/terraform-provider-auth0/pull/202))
- `resource/auth0_connection`: Changed `options.fields_map` from Map to String ([#207](https://github.com/auth0/terraform-provider-auth0/pull/207))
  - Please make sure to do a `terraform state rm auth0_connection.<resource name> && terraform import auth0_connection.<resource name> <connID>` in order to prevent issues due to the breaking change after upgrading to this version.

## 0.31.0

FEATURES:

- `resource/auth0_hook`: Add warnings for untracked hook secrets ([#189](https://github.com/auth0/terraform-provider-auth0/pull/189))
- `resource/auth0_log_stream`: Add datadog_region validation ([#192](https://github.com/auth0/terraform-provider-auth0/pull/192))
- `resource/auth0_connection`: Add computed to password* fields in connection resource ([#191](https://github.com/auth0/terraform-provider-auth0/pull/191))
- `resource/auth0_connection`: Adding icon_url to OAuth2 connection types ([#196](https://github.com/auth0/terraform-provider-auth0/pull/196))

BUG FIXES:

- `resource/auth0_connection`: Fix crash with domain_aliases for ADFS ([#172](https://github.com/auth0/terraform-provider-auth0/pull/172))
- `resource/auth0_connection`: Fix subsequent updates to partial refresh_token object ([#187](https://github.com/auth0/terraform-provider-auth0/pull/187))
- `resource/auth0_tenant`: Setting session lifetime values as non-computed ([#193](https://github.com/auth0/terraform-provider-auth0/pull/193))
- `resource/auth0_user`: Preserve user ID casing in state ([#197](https://github.com/auth0/terraform-provider-auth0/pull/197))
- `resource/auth0_guardian`: Fix phone options issue#159 and refactor guardian resource implementation ([#195](https://github.com/auth0/terraform-provider-auth0/pull/195))


NOTES:

- Correct docs example typo binding_method to protocol_binding ([#179](https://github.com/auth0/terraform-provider-auth0/pull/179))
- Enabled http recordings with go-vcr to be used within tests for more reliable testing
- Adding documentation for passwordless email connection ([#179](https://github.com/auth0/terraform-provider-auth0/pull/179))
- Adding GitHub connection scopes documentation ([#199](https://github.com/auth0/terraform-provider-auth0/pull/199))


## 0.30.3

FEATURES:

- `resource/auth0_connection`: Added support for connection metadata field ([#158](https://github.com/auth0/terraform-provider-auth0/pull/158))


## 0.30.2

BUG FIXES:

- `resource/auth0_tenant`: Fixed auth0 tenant flags so it only sends set values ([#144](https://github.com/auth0/terraform-provider-auth0/pull/144))
- `resource/auth0_connection`: Fixed missing options when importing a connection object ([#142](https://github.com/auth0/terraform-provider-auth0/pull/142))

NOTES:

- `resource/auth0_custom_domain`: Removed deprecated verification_method from custom domain resource ([#143](https://github.com/auth0/terraform-provider-auth0/pull/143))


## 0.30.1

BUG FIXES:

- `resource/auth0_client`: Fix conversion issue flattenAddons func in client resource ([#140](https://github.com/auth0/terraform-provider-auth0/pull/140))


## 0.30.0

FEATURES:

- `resource/auth0_custom_domain*`: Added support for creating external resources associated with self-managed certificates ([#118](https://github.com/auth0/terraform-provider-auth0/pull/118))
- `resource/auth0_log_stream`: [DXCDT-104] Added filters to log_stream resource ([#133](https://github.com/auth0/terraform-provider-auth0/pull/133))


BUG FIXES:

- `resource/auth0_log_stream`: Fixed serialization of log streams with http sink custom headers ([#120](https://github.com/auth0/terraform-provider-auth0/pull/120))

NOTES:

- Upgraded to terraform-plugin-sdk@v2 ([#121](https://github.com/auth0/terraform-provider-auth0/pull/121), [#122](https://github.com/auth0/terraform-provider-auth0/pull/122), [#126](https://github.com/auth0/terraform-provider-auth0/pull/126))


## 0.29.0

FEATURES:

* `data-source/auth0_tenant`: Added a new data source for a subset of a tenant's info ([#95](https://github.com/auth0/terraform-provider-auth0/pull/95))

BUG FIXES:

* `resource/*`: [DXCDT-80] Fixed surfaced errors on all resources after not ignoring them when setting resource data (
  [#106](https://github.com/auth0/terraform-provider-auth0/pull/106),
  [#104](https://github.com/auth0/terraform-provider-auth0/pull/104),
  [#105](https://github.com/auth0/terraform-provider-auth0/pull/105),
  [#94](https://github.com/auth0/terraform-provider-auth0/pull/94),
  [#97](https://github.com/auth0/terraform-provider-auth0/pull/97),
  [#101](https://github.com/auth0/terraform-provider-auth0/pull/101),
  [#113](https://github.com/auth0/terraform-provider-auth0/pull/113),
  [#112](https://github.com/auth0/terraform-provider-auth0/pull/112),
  [#111](https://github.com/auth0/terraform-provider-auth0/pull/111),
  [#110](https://github.com/auth0/terraform-provider-auth0/pull/110),
  [#114](https://github.com/auth0/terraform-provider-auth0/pull/114),
  [#109](https://github.com/auth0/terraform-provider-auth0/pull/109)
)
* `resource/auth0_action`: Failed fast when action fails to build ([#107](https://github.com/auth0/terraform-provider-auth0/pull/107))


## 0.28.1

BUG FIXES:

* `resource/auth0_attack_protection`: Fix attack protection resource for PSaaS Tenants ([#86](https://github.com/auth0/terraform-provider-auth0/pull/86))


## 0.28.0

FEATURES:

* `resource/auth0_attack_protection`: Added Attack Protection Management Resource ([#77](https://github.com/auth0/terraform-provider-auth0/pull/77))

ENHANCEMENTS:

* `resource/auth0_connection`: Added ShowAsButton option for enterprise connections ([#80](https://github.com/auth0/terraform-provider-auth0/pull/80))

BUG FIXES:

* `resource/auth0_tenant`: Wiring `default_redirection_uri` parameter in 'auth0_tenant' into Auth0 API call ([#71](https://github.com/auth0/terraform-provider-auth0/pull/71))
* `resource/auth0_client`: Mark signing_keys as sensitive ([#72](https://github.com/auth0/terraform-provider-auth0/pull/72))


## 0.27.1

ENHANCEMENTS:

* Added Signing Keys to client resources ([#66](https://github.com/auth0/terraform-provider-auth0/pull/66))
* Update documentation to include missing resources and show how to import each resource ([#67](https://github.com/auth0/terraform-provider-auth0/pull/67))


## 0.27.0

ENHANCEMENTS:

* Added ability to authenticate with [management API tokens](https://auth0.com/docs/secure/tokens/access-tokens/management-api-access-tokens) ([#487](https://github.com/alexkappa/terraform-provider-auth0/pull/487))
* Added client data source ([#511](https://github.com/alexkappa/terraform-provider-auth0/pull/511))
* Added global client data source ([#512](https://github.com/alexkappa/terraform-provider-auth0/pull/512))

NOTES:

* Added reference to  `initiate_login_uri` property in client documentation ([#513](https://github.com/alexkappa/terraform-provider-auth0/pull/513))

## Previous History

This project is a continuation of [alexkappa/terraform-provider-auth0](https://github.com/alexkappa/terraform-provider-auth0), to view the previous change history, please see that [repo's changelog](https://github.com/alexkappa/terraform-provider-auth0/blob/master/CHANGELOG.md).
