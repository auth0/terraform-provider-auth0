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
