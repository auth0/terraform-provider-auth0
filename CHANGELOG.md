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
