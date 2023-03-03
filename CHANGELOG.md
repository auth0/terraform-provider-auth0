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
