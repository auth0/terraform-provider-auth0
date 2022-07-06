---
layout: "auth0"
page_title: "Auth0: auth0_tenant"
description: |-
  With this resource, you can manage Auth0 tenants, including setting logos and support contact information, 
  setting error pages, and configuring default tenant behaviors.
---

# auth0_tenant

With this resource, you can manage Auth0 tenants, including setting logos and support contact information, setting error
pages, and configuring default tenant behaviors.

~> Auth0 does not currently support creating tenants through the Management API. Therefore, this resource can only 
manage an existing tenant created through the Auth0 dashboard. 

Auth0 does not currently support adding/removing extensions on tenants through their API. The Auth0 dashboard must be 
used to add/remove extensions. 

## Example Usage

```hcl
resource "auth0_tenant" "tenant" {
  change_password {
    enabled = true
    html    = file("./password_reset.html")
  }

  guardian_mfa_page {
    enabled = true
    html    = file("./guardian_multifactor.html")
  }

  default_audience  = "<client_id>"
  default_directory = "Connection-Name"

  error_page {
    html          = file("./error.html")
    show_log_link = true
    url           = "http://mysite/errors"
  }

  friendly_name = "Tenant Name"
  picture_url   = "http://mysite/logo.png"
  support_email = "support@mysite"
  support_url   = "http://mysite/support"
  allowed_logout_urls = [
    "http://mysite/logout"
  ]
  session_lifetime = 46000
  sandbox_version  = "8"
}
```

## Argument Reference

Arguments accepted by this resource include:

* `change_password` - (Optional) List(Resource). Configuration settings for change passsword page. For details, see [Change Password Page](#change-password-page).
* `guardian_mfa_page` - (Optional) List(Resource). Configuration settings for the Guardian MFA page. For details, see [Guardian MFA Page](#guardian-mfa-page).
* `default_audience` - (Optional) String. API Audience to use by default for API Authorization flows. This setting is equivalent to appending the audience to every authorization request made to the tenant for every application.
* `default_directory` - (Optional) String. Name of the connection to be used for Password Grant exchanges. Options include `auth0-adldap`, `ad`, `auth0`, `email`, `sms`, `waad`, and `adfs`.
* `error_page` - (Optional) List(Resource). Configuration settings for error pages. For details, see [Error Page](#error-page).
* `friendly_name` - (Optional) String. Friendly name for the tenant.
* `picture_url` - (Optional). String URL of logo to be shown for the tenant. Recommended size is 150px x 150px. If no URL is provided, the Auth0 logo will be used. 
* `support_email` - (Optional) String. Support email address for authenticating users.
* `support_url` - (Optional) String. Support URL for authenticating users.
* `allowed_logout_urls` - (Optional) List(String). URLs that Auth0 may redirect to after logout.
* `session_lifetime` - (Optional) Integer. Number of hours during which a session will stay valid.
* `sandbox_version` - (Optional) String. Selected sandbox version for the extensibility environment, which allows you to use custom scripts to extend parts of Auth0's functionality.
* `idle_session_lifetime` - (Optional) Integer. Number of hours during which a session can be inactive before the user must log in again.
* `enabled_locales`- (Optional) List(String). Supported locales for the user interface. The first locale in the list will be used to set the default locale.
* `flags` - (Optional) List(Resource). Configuration settings for tenant flags. For details, see [Flags](#flags).
* `universal_login` - (Optional) List(Resource). Configuration settings for Universal Login. For details, see [Universal Login](#universal-login).
* `default_redirection_uri` - (Optional) String. The default absolute redirection uri, must be https and cannot contain a fragment.

### Change Password Page

`change_password_page` supports the following arguments:

* `enabled` - (Required) Boolean. Indicates whether to use the custom change password page.
* `html` - (Required) String, HTML format with supported Liquid syntax. Customized content of the change password page.

### Guardian MFA Page

`guardian_mfa_page` supports the following arguments:

* `enabled` - (Required) Boolean. Indicates whether to use the custom Guardian page.
* `html` - (Required) String, HTML format with supported Liquid syntax. Customized content of the Guardian page.

### Error Page

`error_page` supports the following arguments:

* `html` - (Required) String, HTML format with supported Liquid syntax. Customized content of the error page.
* `show_log_link` - (Required) Boolean. Indicates whether to show the link to logs as part of the default error page.
* `url` - (Required) String. URL to redirect to when an error occurs rather than showing the default error page.

### Flags

`flags` supports the following arguments:

* `enable_client_connections` - (Optional) Boolean. Indicates whether all current connections should be enabled when a new client is created.
* `enable_apis_section` - (Optional) Boolean. Indicates whether the APIs section is enabled for the tenant.
* `enable_pipeline2` - (Optional) Boolean. Indicates whether advanced API Authorization scenarios are enabled.
* `enable_dynamic_client_registration` - (Optional) Boolean. Indicates whether the tenant allows dynamic client registration.
* `enable_custom_domain_in_emails` - (Optional) Boolean. Indicates whether the tenant allows custom domains in emails.
* `universal_login` - (Optional) Boolean. Indicates whether the tenant uses universal login.
* `enable_legacy_logs_search_v2` - (Optional) Boolean. Indicates whether to use the older v2 legacy logs search.
* `disable_clickjack_protection_headers` - (Optional) Boolean. Indicated whether classic Universal Login prompts include additional security headers to prevent clickjacking.
* `enable_public_signup_user_exists_error` - (Optional) Boolean. Indicates whether the public sign up process shows a user_exists error if the user already exists.
* `allow_legacy_delegation_grant_types` - (Optional) Boolean. Whether the legacy delegation endpoint will be enabled for your account (true) or not available (false).
* `allow_legacy_ro_grant_types` - (Optional) Boolean. Whether the legacy `auth/ro` endpoint (used with resource owner password and passwordless features) will be enabled for your account (true) or not available (false).
* `allow_legacy_tokeninfo_endpoint` - (Optional) Boolean. If enabled, customers can use Tokeninfo Endpoint, otherwise they can not use it.
* `enable_legacy_profile` - (Optional) Boolean. Whether ID tokens and the userinfo endpoint includes a complete user profile (true) or only OpenID Connect claims (false).
* `enable_idtoken_api2` - (Optional) Boolean. Whether ID tokens can be used to authorize some types of requests to API v2 (true) not not (false).
* `no_disclose_enterprise_connections` - (Optional) Boolean. Do not Publish Enterprise Connections Information with IdP domains on the lock configuration file.
* `disable_management_api_sms_obfuscation` - (Optional) Boolean. If true, SMS phone numbers will not be obfuscated in Management API GET calls.
* `enable_adfs_waad_email_verification` - (Optional) Boolean. If enabled, users will be presented with an email verification prompt during their first login when using Azure AD or ADFS connections.
* `revoke_refresh_token_grant` - (Optional) Boolean. Delete underlying grant when a Refresh Token is revoked via the Authentication API.
* `dashboard_log_streams_next` - (Optional) Boolean. Enables beta access to log streaming changes.
* `dashboard_insights_view` - (Optional) Boolean. Enables new insights activity page view.
* `disable_fields_map_fix` - (Optional) Boolean. Disables SAML fields map fix for bad mappings with repeated attributes.

### Universal Login

`universal_login` supports the following arguments:

* `colors` - (Optional) List(Resource). Configuration settings for Universal Login colors. See [Universal Login - Colors](#colors).

#### Colors 

`colors` supports the following arguments:

* `primary` - (Optional) String, Hexadecimal. Primary button background color.
* `page_background` - (Optional) String, Hexadecimal. Background color of login pages.

## Attribute Reference

Attributes exported by this resource include:

* `sandbox_version` - String. Selected sandbox version for the extensibility environment, which allows you to use custom scripts to extend parts of Auth0's functionality.

## Import

As this is not a resource identifiable by an ID within the Auth0 Management API, tenant can be imported using a random
string. We recommend [Version 4 UUID](https://www.uuidgenerator.net/version4) e.g.

```shell
$ terraform import auth0_tenant.tenant 82f4f21b-017a-319d-92e7-2291c1ca36c4
```
