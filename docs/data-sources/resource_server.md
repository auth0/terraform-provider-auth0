---
page_title: "Data Source: auth0_resource_server"
description: |-
  Data source to retrieve a specific Auth0 resource server by resource_server_id or identifier.
---

# Data Source: auth0_resource_server

Data source to retrieve a specific Auth0 resource server by `resource_server_id` or `identifier`.

## Example Usage

```terraform
# An Auth0 Resource Server loaded using its identifier.
data "auth0_resource_server" "some-resource-server-by-identifier" {
  identifier = "https://my-api.com/v1"
}

# An Auth0 Resource Server loaded using its ID.
data "auth0_resource_server" "some-resource-server-by-id" {
  resource_server_id = "abcdefghkijklmnopqrstuvwxyz0123456789"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `identifier` (String) Unique identifier for the resource server. Used as the audience parameter for authorization calls. If not provided, `resource_server_id` must be set.
- `resource_server_id` (String) The ID of the resource server. If not provided, `identifier` must be set.

### Read-Only

- `allow_offline_access` (Boolean) Indicates whether refresh tokens can be issued for this resource server.
- `authorization_details` (List of Object) Authorization details for this resource server. (see [below for nested schema](#nestedatt--authorization_details))
- `consent_policy` (String) Consent policy for this resource server. Options include `transactional-authorization-with-mfa`, or `null` to disable.
- `enforce_policies` (Boolean) If this setting is enabled, RBAC authorization policies will be enforced for this API. Role and permission assignments will be evaluated during the login transaction.
- `id` (String) The ID of this resource.
- `name` (String) Friendly name for the resource server. Cannot include `<` or `>` characters.
- `proof_of_possession` (List of Object) Configuration settings for proof-of-possession for this resource server. (see [below for nested schema](#nestedatt--proof_of_possession))
- `scopes` (Set of Object) List of permissions (scopes) used by this resource server. (see [below for nested schema](#nestedatt--scopes))
- `signing_alg` (String) Algorithm used to sign JWTs. Options include `HS256`, `RS256`, and `PS256`.
- `signing_secret` (String) Secret used to sign tokens when using symmetric algorithms (HS256).
- `skip_consent_for_verifiable_first_party_clients` (Boolean) Indicates whether to skip user consent for applications flagged as first party.
- `token_dialect` (String) Dialect of access tokens that should be issued for this resource server. Options include `access_token`, `rfc9068_profile`, `access_token_authz`, and `rfc9068_profile_authz`. `access_token` is a JWT containing standard Auth0 claims. `rfc9068_profile` is a JWT conforming to the IETF JWT Access Token Profile. `access_token_authz` is a JWT containing standard Auth0 claims, including RBAC permissions claims. `rfc9068_profile_authz` is a JWT conforming to the IETF JWT Access Token Profile, including RBAC permissions claims. RBAC permissions claims are available if RBAC (`enforce_policies`) is enabled for this API. For more details, refer to [Access Token Profiles](https://auth0.com/docs/secure/tokens/access-tokens/access-token-profiles).
- `token_encryption` (List of Object) Configuration for JSON Web Encryption(JWE) of tokens for this resource server. (see [below for nested schema](#nestedatt--token_encryption))
- `token_lifetime` (Number) Number of seconds during which access tokens issued for this resource server from the token endpoint remain valid.
- `token_lifetime_for_web` (Number) Number of seconds during which access tokens issued for this resource server via implicit or hybrid flows remain valid. Cannot be greater than the `token_lifetime` value.
- `verification_location` (String) URL from which to retrieve JWKs for this resource server. Used for verifying the JWT sent to Auth0 for token introspection.

<a id="nestedatt--authorization_details"></a>
### Nested Schema for `authorization_details`

Read-Only:

- `disable` (Boolean)
- `type` (String)


<a id="nestedatt--proof_of_possession"></a>
### Nested Schema for `proof_of_possession`

Read-Only:

- `disable` (Boolean)
- `mechanism` (String)
- `required` (Boolean)


<a id="nestedatt--scopes"></a>
### Nested Schema for `scopes`

Read-Only:

- `description` (String)
- `name` (String)


<a id="nestedatt--token_encryption"></a>
### Nested Schema for `token_encryption`

Read-Only:

- `disable` (Boolean)
- `encryption_key` (List of Object) (see [below for nested schema](#nestedobjatt--token_encryption--encryption_key))
- `format` (String)

<a id="nestedobjatt--token_encryption--encryption_key"></a>
### Nested Schema for `token_encryption.encryption_key`

Read-Only:

- `algorithm` (String)
- `kid` (String)
- `name` (String)
- `pem` (String)


