---
page_title: "Resource: auth0_connection"
description: |-
  With Auth0, you can define sources of users, otherwise known as connections, which may include identity providers (such as Google or LinkedIn), databases, or passwordless authentication methods. This resource allows you to configure and manage connections to be used with your clients and users.
---

# Resource: auth0_connection

With Auth0, you can define sources of users, otherwise known as connections, which may include identity providers (such as Google or LinkedIn), databases, or passwordless authentication methods. This resource allows you to configure and manage connections to be used with your clients and users.

~> The Auth0 dashboard displays only one connection per social provider. Although the Auth0 Management API allows the
creation of multiple connections per strategy, the additional connections may not be visible in the Auth0 dashboard.

~> When updating the `options` parameter, ensure that all nested fields within the `options` schema are explicitly defined. Failing to do so may result in the loss of existing configurations.

## Example Usage

### Auth0 Connection

```terraform
# This is an example of an Auth0 connection.

resource "auth0_connection" "my_connection" {
  name                 = "Example-Connection"
  is_domain_connection = true
  strategy             = "auth0"
  metadata = {
    key1 = "foo"
    key2 = "bar"
  }

  options {
    password_policy                = "excellent"
    brute_force_protection         = true
    strategy_version               = 2
    enabled_database_customization = true
    import_mode                    = false
    requires_username              = true
    disable_signup                 = false
    custom_scripts = {
      get_user = <<EOF
        function getByEmail(email, callback) {
          return callback(new Error("Whoops!"));
        }
      EOF
    }
    configuration = {
      foo = "bar"
      bar = "baz"
    }
    upstream_params = jsonencode({
      "screen_name" : {
        "alias" : "login_hint"
      }
    })

    password_history {
      enable = true
      size   = 3
    }

    password_no_personal_info {
      enable = true
    }

    password_dictionary {
      enable     = true
      dictionary = ["password", "admin", "1234"]
    }

    password_complexity_options {
      min_length = 12
    }

    validation {
      username {
        min = 10
        max = 40
      }
    }

    mfa {
      active                 = true
      return_enroll_settings = true
    }

    authentication_methods {
      passkey {
        enabled = true
      }
      password {
        enabled = true
      }
    }
    passkey_options {
      challenge_ui                   = "both"
      local_enrollment_enabled       = true
      progressive_enrollment_enabled = true
    }
  }
}
```

### Google OAuth2 Connection

~> Your Auth0 account may be pre-configured with a `google-oauth2` connection.

```terraform
# This is an example of a Google OAuth2 connection.

resource "auth0_connection" "google_oauth2" {
  name     = "Google-OAuth2-Connection"
  strategy = "google-oauth2"

  options {
    client_id                = "<client-id>"
    client_secret            = "<client-secret>"
    allowed_audiences        = ["example.com", "api.example.com"]
    scopes                   = ["email", "profile", "gmail", "youtube"]
    set_user_root_attributes = "on_each_login"
    non_persistent_attrs     = ["ethnicity", "gender"]
  }
}
```

### Google Apps

```terraform
resource "auth0_connection" "google_apps" {
  name                 = "connection-google-apps"
  is_domain_connection = false
  strategy             = "google-apps"
  show_as_button       = false
  options {
    client_id        = ""
    client_secret    = ""
    domain           = "example.com"
    tenant_domain    = "example.com"
    domain_aliases   = ["example.com", "api.example.com"]
    api_enable_users = true
    scopes           = ["ext_profile", "ext_groups"]
    icon_url         = "https://example.com/assets/logo.png"
    upstream_params = jsonencode({
      "screen_name" : {
        "alias" : "login_hint"
      }
    })
    set_user_root_attributes = "on_each_login"
    non_persistent_attrs     = ["ethnicity", "gender"]
  }
}
```

### Facebook Connection

```terraform
# This is an example of a Facebook connection.

resource "auth0_connection" "facebook" {
  name     = "Facebook-Connection"
  strategy = "facebook"

  options {
    client_id     = "<client-id>"
    client_secret = "<client-secret>"
    scopes = [
      "public_profile",
      "email",
      "groups_access_member_info",
      "user_birthday"
    ]
    set_user_root_attributes = "on_each_login"
    non_persistent_attrs     = ["ethnicity", "gender"]
  }
}
```

### Apple Connection

```terraform
# This is an example of an Apple connection.

resource "auth0_connection" "apple" {
  name     = "Apple-Connection"
  strategy = "apple"

  options {
    client_id                = "<client-id>"
    client_secret            = "-----BEGIN PRIVATE KEY-----\nMIHBAgEAMA0GCSqGSIb3DQEBAQUABIGsMIGpAgEAA\n-----END PRIVATE KEY-----"
    team_id                  = "<team-id>"
    key_id                   = "<key-id>"
    scopes                   = ["email", "name"]
    set_user_root_attributes = "on_first_login"
    non_persistent_attrs     = ["ethnicity", "gender"]
  }
}
```

### LinkedIn Connection

```terraform
# This is an example of an LinkedIn connection.

resource "auth0_connection" "linkedin" {
  name     = "Linkedin-Connection"
  strategy = "linkedin"

  options {
    client_id                = "<client-id>"
    client_secret            = "<client-secret>"
    strategy_version         = 2
    scopes                   = ["basic_profile", "profile", "email"]
    set_user_root_attributes = "on_each_login"
    non_persistent_attrs     = ["ethnicity", "gender"]
  }
}
```

### GitHub Connection

```terraform
# This is an example of an GitHub connection.

resource "auth0_connection" "github" {
  name     = "GitHub-Connection"
  strategy = "github"

  options {
    client_id                = "<client-id>"
    client_secret            = "<client-secret>"
    scopes                   = ["email", "profile", "public_repo", "repo"]
    set_user_root_attributes = "on_each_login"
    non_persistent_attrs     = ["ethnicity", "gender"]
  }
}
```

### SalesForce Connection

```terraform
# This is an example of an SalesForce connection.

resource "auth0_connection" "salesforce" {
  name     = "Salesforce-Connection"
  strategy = "salesforce"

  options {
    client_id                = "<client-id>"
    client_secret            = "<client-secret>"
    community_base_url       = "https://salesforce.example.com"
    scopes                   = ["openid", "email"]
    set_user_root_attributes = "on_first_login"
    non_persistent_attrs     = ["ethnicity", "gender"]
  }
}
```

### OAuth2 Connection

Also applies to following connection strategies: `dropbox`, `bitbucket`, `paypal`, `twitter`, `amazon`, `yahoo`, `box`, `wordpress`, `shopify`, `custom`

```terraform
# This is an example of an OAuth2 connection.

resource "auth0_connection" "oauth2" {
  name     = "OAuth2-Connection"
  strategy = "oauth2"

  options {
    client_id              = "<client-id>"
    client_secret          = "<client-secret>"
    strategy_version       = 2
    scopes                 = ["basic_profile", "profile", "email"]
    token_endpoint         = "https://auth.example.com/oauth2/token"
    authorization_endpoint = "https://auth.example.com/oauth2/authorize"
    pkce_enabled           = true
    icon_url               = "https://auth.example.com/assets/logo.png"
    custom_headers {
      header = "bar"
      value  = "foo"
    }
    custom_headers {
      header = "foo"
      value  = "bar"
    }
    scripts = {
      fetchUserProfile = <<EOF
        function fetchUserProfile(accessToken, context, callback) {
          return callback(new Error("Whoops!"));
        }
      EOF
    }
    set_user_root_attributes = "on_each_login"
    non_persistent_attrs     = ["ethnicity", "gender"]
  }
}
```

### Active Directory (AD)

```terraform
resource "auth0_connection" "ad" {
  name           = "connection-active-directory"
  display_name   = "Active Directory Connection"
  strategy       = "ad"
  show_as_button = true

  options {
    disable_self_service_change_password = true
    brute_force_protection               = true
    tenant_domain                        = "example.com"
    strategy_version                     = 2
    icon_url                             = "https://example.com/assets/logo.png"
    domain_aliases = [
      "example.com",
      "api.example.com"
    ]
    ips                      = ["192.168.1.1", "192.168.1.2"]
    set_user_root_attributes = "on_each_login"
    non_persistent_attrs     = ["ethnicity", "gender"]
    upstream_params = jsonencode({
      "screen_name" : {
        "alias" : "login_hint"
      }
    })
    use_cert_auth = false
    use_kerberos  = false
    disable_cache = false
  }
}
```

### Azure AD Connection

```terraform
resource "auth0_connection" "azure_ad" {
  name           = "connection-azure-ad"
  strategy       = "waad"
  show_as_button = true
  options {
    identity_api      = "azure-active-directory-v1.0"
    client_id         = "123456"
    client_secret     = "123456"
    strategy_version  = 2
    user_id_attribute = "oid"
    app_id            = "app-id-123"
    tenant_domain     = "example.onmicrosoft.com"
    domain            = "example.onmicrosoft.com"
    domain_aliases = [
      "example.com",
      "api.example.com"
    ]
    icon_url               = "https://example.onmicrosoft.com/assets/logo.png"
    use_wsfed              = false
    waad_protocol          = "openid-connect"
    waad_common_endpoint   = false
    max_groups_to_retrieve = 250
    api_enable_users       = true
    scopes = [
      "basic_profile",
      "ext_groups",
      "ext_profile"
    ]
    set_user_root_attributes               = "on_each_login"
    should_trust_email_verified_connection = "never_set_emails_as_verified"
    upstream_params = jsonencode({
      "screen_name" : {
        "alias" : "login_hint"
      }
    })
    non_persistent_attrs = ["ethnicity", "gender"]
  }
}
```

### SMS Connection

~> To be able to see this in the management dashboard as well, the name of the connection must be set to "sms".

```terraform
# This is an example of an SMS connection.

resource "auth0_connection" "sms" {
  name     = "SMS-Connection"
  strategy = "sms"

  options {
    name                   = "SMS OTP"
    twilio_sid             = "<twilio-sid>"
    twilio_token           = "<twilio-token>"
    from                   = "<phone-number>"
    syntax                 = "md_with_macros"
    template               = "Your one-time password is @@password@@"
    messaging_service_sid  = "<messaging-service-sid>"
    disable_signup         = false
    brute_force_protection = true
    forward_request_info   = true

    totp {
      time_step = 300
      length    = 6
    }

    provider    = "sms_gateway"
    gateway_url = "https://somewhere.com/sms-gateway"
    gateway_authentication {
      method                = "bearer"
      subject               = "test.us.auth0.com:sms"
      audience              = "https://somewhere.com/sms-gateway"
      secret                = "4e2680bb72ec2ae24836476dd37ed6c2"
      secret_base64_encoded = false
    }
  }
}

# This is an example of an SMS connection with a custom SMS gateway.

resource "auth0_connection" "sms" {
  name                 = "custom-sms-gateway"
  is_domain_connection = false
  strategy             = "sms"

  options {
    disable_signup         = false
    name                   = "sms"
    from                   = "+15555555555"
    syntax                 = "md_with_macros"
    template               = "@@password@@"
    brute_force_protection = true
    provider               = "sms_gateway"
    gateway_url            = "https://somewhere.com/sms-gateway"
    forward_request_info   = true

    totp {
      time_step = 300
      length    = 6
    }

    gateway_authentication {
      method                = "bearer"
      subject               = "test.us.auth0.com:sms"
      audience              = "https://somewhere.com/sms-gateway"
      secret                = "4e2680bb74ec2ae24736476dd37ed6c2"
      secret_base64_encoded = false
    }
  }
}
```

### Email Connection

~> To be able to see this in the management dashboard as well, the name of the connection must be set to "email".

```terraform
# This is an example of an Email connection.

resource "auth0_connection" "passwordless_email" {
  strategy = "email"
  name     = "email"

  options {
    name                     = "email"
    from                     = "{{ application.name }} \u003croot@auth0.com\u003e"
    subject                  = "Welcome to {{ application.name }}"
    syntax                   = "liquid"
    template                 = "<html>This is the body of the email</html>"
    disable_signup           = false
    brute_force_protection   = true
    set_user_root_attributes = "on_each_login"
    non_persistent_attrs     = []
    auth_params = {
      scope         = "openid email profile offline_access"
      response_type = "code"
    }

    totp {
      time_step = 300
      length    = 6
    }
  }
}
```

### SAML Connection

```terraform
# This is an example of a SAML connection.

resource "auth0_connection" "samlp" {
  name     = "SAML-Connection"
  strategy = "samlp"

  options {
    debug                           = false
    signing_cert                    = "<signing-certificate>"
    sign_in_endpoint                = "https://saml.provider/sign_in"
    sign_out_endpoint               = "https://saml.provider/sign_out"
    global_token_revocation_jwt_iss = "issuer.example.com"
    global_token_revocation_jwt_sub = "user123"
    disable_sign_out                = true
    strategy_version                = 2
    tenant_domain                   = "example.com"
    domain_aliases                  = ["example.com", "alias.example.com"]
    protocol_binding                = "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
    request_template                = "<samlp:AuthnRequest xmlns:samlp=\"urn:oasis:names:tc:SAML:2.0:protocol\"\n@@AssertServiceURLAndDestination@@\n    ID=\"@@ID@@\"\n    IssueInstant=\"@@IssueInstant@@\"\n    ProtocolBinding=\"@@ProtocolBinding@@\" Version=\"2.0\">\n    <saml:Issuer xmlns:saml=\"urn:oasis:names:tc:SAML:2.0:assertion\">@@Issuer@@</saml:Issuer>\n</samlp:AuthnRequest>"
    user_id_attribute               = "https://saml.provider/imi/ns/identity-200810"
    signature_algorithm             = "rsa-sha256"
    digest_algorithm                = "sha256"
    icon_url                        = "https://saml.provider/assets/logo.png"
    entity_id                       = "<entity_id>"
    metadata_xml                    = <<EOF
    <?xml version="1.0"?>
    <md:EntityDescriptor xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata" xmlns:ds="http://www.w3.org/2000/09/xmldsig#" entityID="https://example.com">
      <md:IDPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
        <md:SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="https://saml.provider/sign_out"/>
        <md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="https://saml.provider/sign_in"/>
      </md:IDPSSODescriptor>
    </md:EntityDescriptor>
    EOF
    metadata_url                    = "https://saml.provider/imi/ns/FederationMetadata.xml" # Use either metadata_url or metadata_xml, but not both.

    fields_map = jsonencode({
      "name" : ["name", "nameidentifier"]
      "email" : ["emailaddress", "nameidentifier"]
      "family_name" : "surname"
    })

    signing_key {
      key  = "-----BEGIN PRIVATE KEY-----\n...{your private key here}...\n-----END PRIVATE KEY-----"
      cert = "-----BEGIN CERTIFICATE-----\n...{your public key cert here}...\n-----END CERTIFICATE-----"
    }

    decryption_key {
      key  = "-----BEGIN PRIVATE KEY-----\n...{your private key here}...\n-----END PRIVATE KEY-----"
      cert = "-----BEGIN CERTIFICATE-----\n...{your public key cert here}...\n-----END CERTIFICATE-----"
    }

    idp_initiated {
      client_id              = "client_id"
      client_protocol        = "samlp"
      client_authorize_query = "type=code&timeout=30"
    }
  }
}
```

### WindowsLive Connection

```terraform
# This is an example of a WindowsLive connection.

resource "auth0_connection" "windowslive" {
  name     = "Windowslive-Connection"
  strategy = "windowslive"

  options {
    client_id                = "<client-id>"
    client_secret            = "<client-secret>"
    strategy_version         = 2
    scopes                   = ["signin", "graph_user"]
    set_user_root_attributes = "on_first_login"
    non_persistent_attrs     = ["ethnicity", "gender"]
  }
}
```

### OIDC Connection

```terraform
# This is an example of an OIDC connection.

resource "auth0_connection" "oidc" {
  name           = "oidc-connection"
  display_name   = "OIDC Connection"
  strategy       = "oidc"
  show_as_button = false

  options {
    client_id                = "1234567"
    client_secret            = "1234567"
    domain_aliases           = ["example.com"]
    tenant_domain            = ""
    icon_url                 = "https://example.com/assets/logo.png"
    type                     = "back_channel"
    issuer                   = "https://www.paypalobjects.com"
    jwks_uri                 = "https://api.paypal.com/v1/oauth2/certs"
    discovery_url            = "https://www.paypalobjects.com/.well-known/openid-configuration"
    token_endpoint           = "https://api.paypal.com/v1/oauth2/token"
    userinfo_endpoint        = "https://api.paypal.com/v1/oauth2/token/userinfo"
    authorization_endpoint   = "https://www.paypal.com/signin/authorize"
    scopes                   = ["openid", "email"]
    set_user_root_attributes = "on_first_login"
    non_persistent_attrs     = ["ethnicity", "gender"]

    connection_settings {
      pkce = "auto"
    }

    attribute_map {
      mapping_mode   = "use_map"
      userinfo_scope = "openid email profile groups"
      attributes = jsonencode({
        "name" : "$${context.tokenset.name}",
        "email" : "$${context.tokenset.email}",
        "email_verified" : "$${context.tokenset.email_verified}",
        "nickname" : "$${context.tokenset.nickname}",
        "picture" : "$${context.tokenset.picture}",
        "given_name" : "$${context.tokenset.given_name}",
        "family_name" : "$${context.tokenset.family_name}"
      })
    }
  }
}
```

### Okta Connection

!> When configuring an Okta Workforce connection, the `scopes` attribute must be explicitly set. If omitted, the connection may not function correctly.
To ensure proper behavior, always specify:  `scopes = ["openid", "profile", "email"]`

```terraform
# This is an example of an Okta Workforce connection.

resource "auth0_connection" "okta" {
  name           = "okta-connection"
  display_name   = "Okta Workforce Connection"
  strategy       = "okta"
  show_as_button = false

  options {
    client_id                = "1234567"
    client_secret            = "1234567"
    domain                   = "example.okta.com"
    domain_aliases           = ["example.com"]
    issuer                   = "https://example.okta.com"
    jwks_uri                 = "https://example.okta.com/oauth2/v1/keys"
    token_endpoint           = "https://example.okta.com/oauth2/v1/token"
    userinfo_endpoint        = "https://example.okta.com/oauth2/v1/userinfo"
    authorization_endpoint   = "https://example.okta.com/oauth2/v1/authorize"
    scopes                   = ["openid", "profile", "email"]
    set_user_root_attributes = "on_first_login"
    non_persistent_attrs     = ["ethnicity", "gender"]
    upstream_params = jsonencode({
      "screen_name" : {
        "alias" : "login_hint"
      }
    })

    connection_settings {
      pkce = "auto"
    }

    attribute_map {
      mapping_mode   = "basic_profile"
      userinfo_scope = "openid email profile groups"
      attributes = jsonencode({
        "name" : "$${context.tokenset.name}",
        "email" : "$${context.tokenset.email}",
        "email_verified" : "$${context.tokenset.email_verified}",
        "nickname" : "$${context.tokenset.nickname}",
        "picture" : "$${context.tokenset.picture}",
        "given_name" : "$${context.tokenset.given_name}",
        "family_name" : "$${context.tokenset.family_name}"
      })
    }
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name of the connection. This value is immutable and changing it requires the creation of a new resource.
- `strategy` (String) Type of the connection, which indicates the identity provider.

### Optional

- `authentication` (Block List, Max: 1) Configure the purpose of a connection to be used for authentication during login. (see [below for nested schema](#nestedblock--authentication))
- `connected_accounts` (Block List, Max: 1) Configure the purpose of a connection to be used for connected accounts and Token Vault. (see [below for nested schema](#nestedblock--connected_accounts))
- `display_name` (String) Name used in login screen.
- `is_domain_connection` (Boolean) Indicates whether the connection is domain level.
- `metadata` (Map of String) Metadata associated with the connection, in the form of a map of string values (max 255 chars).
- `options` (Block List, Max: 1) Configuration settings for connection options. (see [below for nested schema](#nestedblock--options))
- `realms` (List of String) Defines the realms for which the connection will be used (e.g., email domains). If not specified, the connection name is added as the realm.
- `show_as_button` (Boolean) Display connection as a button. Only available on enterprise connections.

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--authentication"></a>
### Nested Schema for `authentication`

Required:

- `active` (Boolean)


<a id="nestedblock--connected_accounts"></a>
### Nested Schema for `connected_accounts`

Required:

- `active` (Boolean)


<a id="nestedblock--options"></a>
### Nested Schema for `options`

Optional:

- `access_token_url` (String) URL used to exchange a user-authorized request token for an access token.
- `adfs_server` (String) ADFS URL where to fetch the metadata source.
- `allowed_audiences` (Set of String) List of allowed audiences.
- `api_enable_users` (Boolean) Enable API Access to users.
- `app_id` (String) App ID.
- `attribute_map` (Block List, Max: 1) OpenID Connect and Okta Workforce connections can automatically map claims received from the identity provider (IdP). You can configure this mapping through a library template provided by Auth0 or by entering your own template directly. Click [here](https://auth0.com/docs/authenticate/identity-providers/enterprise-identity-providers/configure-pkce-claim-mapping-for-oidc#map-claims-for-oidc-connections) for more info. (see [below for nested schema](#nestedblock--options--attribute_map))
- `attributes` (Block List) Order of attributes for precedence in identification.Valid values: email, phone_number, username. If Precedence is set, it must contain all values (email, phone_number, username) in specific order (see [below for nested schema](#nestedblock--options--attributes))
- `auth_params` (Map of String) Query string parameters to be included as part of the generated passwordless email link.
- `authentication_methods` (Block List) Specifies the authentication methods and their configuration (enabled or disabled) (see [below for nested schema](#nestedblock--options--authentication_methods))
- `authorization_endpoint` (String) Authorization endpoint.
- `brute_force_protection` (Boolean) Indicates whether to enable brute force protection, which will limit the number of signups and failed logins from a suspicious IP address.
- `client_id` (String) The strategy's client ID.
- `client_secret` (String, Sensitive) The strategy's client secret.
- `community_base_url` (String) Salesforce community base URL.
- `configuration` (Map of String, Sensitive) A case-sensitive map of key value pairs used as configuration variables for the `custom_script`.
- `connection_settings` (Block List, Max: 1) Proof Key for Code Exchange (PKCE) configuration settings for an OIDC or Okta Workforce connection. (see [below for nested schema](#nestedblock--options--connection_settings))
- `consumer_key` (String) Identifies the client to the service provider
- `consumer_secret` (String) Secret used to establish ownership of the consumer key.
- `custom_headers` (Block Set) Configure extra headers to the Token endpoint of an OAuth 2.0 provider (see [below for nested schema](#nestedblock--options--custom_headers))
- `custom_scripts` (Map of String) A map of scripts used to integrate with a custom database.
- `debug` (Boolean) When enabled, additional debug information will be generated.
- `decryption_key` (Block List, Max: 1) The key used to decrypt encrypted responses from the connection. Uses the `key` and `cert` properties to provide the private key and certificate respectively. (see [below for nested schema](#nestedblock--options--decryption_key))
- `digest_algorithm` (String) Sign Request Algorithm Digest.
- `disable_cache` (Boolean) Indicates whether to disable the cache or not.
- `disable_self_service_change_password` (Boolean) Indicates whether to remove the forgot password link within the New Universal Login.
- `disable_sign_out` (Boolean) When enabled, will disable sign out.
- `disable_signup` (Boolean) Indicates whether to allow user sign-ups to your application.
- `discovery_url` (String) OpenID discovery URL, e.g. `https://auth.example.com/.well-known/openid-configuration`.
- `domain` (String) Domain name.
- `domain_aliases` (Set of String) List of the domains that can be authenticated using the identity provider. Only needed for Identifier First authentication flows.
- `enable_script_context` (Boolean) Set to `true` to inject context into custom DB scripts (warning: cannot be disabled once enabled).
- `enabled_database_customization` (Boolean) Set to `true` to use a legacy user store.
- `entity_id` (String) Custom Entity ID for the connection.
- `fed_metadata_xml` (String) Federation Metadata for the ADFS connection.
- `fields_map` (String) If you're configuring a SAML enterprise connection for a non-standard PingFederate Server, you must update the attribute mappings.
- `forward_request_info` (Boolean) Specifies whether or not request info should be forwarded to sms gateway.
- `from` (String) Address to use as the sender.
- `gateway_authentication` (Block List, Max: 1) Defines the parameters used to generate the auth token for the custom gateway. (see [below for nested schema](#nestedblock--options--gateway_authentication))
- `gateway_url` (String) Defines a custom sms gateway to use instead of Twilio.
- `global_token_revocation_jwt_iss` (String) Specifies the issuer of the JWT used for global token revocation for the SAML connection.
- `global_token_revocation_jwt_sub` (String) Specifies the subject of the JWT used for global token revocation for the SAML connection.
- `icon_url` (String) Icon URL.
- `identity_api` (String) Azure AD Identity API. Available options are: `microsoft-identity-platform-v2.0` or `azure-active-directory-v1.0`.
- `idp_initiated` (Block List, Max: 1) Configuration options for IDP Initiated Authentication. This is an object with the properties: `client_id`, `client_protocol`, and `client_authorize_query`. (see [below for nested schema](#nestedblock--options--idp_initiated))
- `import_mode` (Boolean) Indicates whether you have a legacy user store and want to gradually migrate those users to the Auth0 user store.
- `ips` (Set of String) A list of IPs.
- `issuer` (String) Issuer URL, e.g. `https://auth.example.com`.
- `jwks_uri` (String) JWKS URI.
- `key_id` (String) Apple Key ID.
- `map_user_id_to_id` (Boolean) By default Auth0 maps `user_id` to `email`. Enabling this setting changes the behavior to map `user_id` to 'id' instead. This can only be defined on a new Google Workspace connection and can not be changed once set.
- `max_groups_to_retrieve` (String) Maximum number of groups to retrieve.
- `messaging_service_sid` (String) SID for Copilot. Used when SMS Source is Copilot.
- `metadata_url` (String) The URL of the SAML metadata document.
- `metadata_xml` (String) The XML content for the SAML metadata document. Values within the xml will take precedence over other attributes set on the options block.
- `mfa` (Block List, Max: 1) Configuration options for multifactor authentication. (see [below for nested schema](#nestedblock--options--mfa))
- `name` (String) The public name of the email or SMS Connection. In most cases this is the same name as the connection name.
- `non_persistent_attrs` (Set of String) If there are user fields that should not be stored in Auth0 databases due to privacy reasons, you can add them to the DenyList here.
- `passkey_options` (Block List, Max: 1) Defines options for the passkey authentication method (see [below for nested schema](#nestedblock--options--passkey_options))
- `password_complexity_options` (Block List, Max: 1) Configuration settings for password complexity. (see [below for nested schema](#nestedblock--options--password_complexity_options))
- `password_dictionary` (Block List, Max: 1) Configuration settings for the password dictionary check, which does not allow passwords that are part of the password dictionary. (see [below for nested schema](#nestedblock--options--password_dictionary))
- `password_history` (Block List) Configuration settings for the password history that is maintained for each user to prevent the reuse of passwords. (see [below for nested schema](#nestedblock--options--password_history))
- `password_no_personal_info` (Block List, Max: 1) Configuration settings for the password personal info check, which does not allow passwords that contain any part of the user's personal data, including user's `name`, `username`, `nickname`, `user_metadata.name`, `user_metadata.first`, `user_metadata.last`, user's `email`, or first part of the user's `email`. (see [below for nested schema](#nestedblock--options--password_no_personal_info))
- `password_policy` (String) Indicates level of password strength to enforce during authentication. A strong password policy will make it difficult, if not improbable, for someone to guess a password through either manual or automated means. Options include `none`, `low`, `fair`, `good`, `excellent`.
- `ping_federate_base_url` (String) Ping Federate Server URL.
- `pkce_enabled` (Boolean) Enables Proof Key for Code Exchange (PKCE) functionality for OAuth2 connections.
- `precedence` (List of String) Order of attributes for precedence in identification.Valid values: email, phone_number, username. If Precedence is set, it must contain all values (email, phone_number, username) in specific order
- `protocol_binding` (String) The SAML Response Binding: how the SAML token is received by Auth0 from the IdP.
- `provider` (String) Defines the custom `sms_gateway` provider.
- `realm_fallback` (Boolean) Allows configuration if connections_realm_fallback flag is enabled for the tenant
- `request_template` (String) Template that formats the SAML request.
- `request_token_url` (String) URL used to obtain an unauthorized request token.
- `requires_username` (Boolean) Indicates whether the user is required to provide a username in addition to an email address.
- `scopes` (Set of String) Permissions to grant to the connection. Within the Auth0 dashboard these appear under the "Attributes" and "Extended Attributes" sections. Some examples: `basic_profile`, `ext_profile`, `ext_nested_groups`, etc.
- `scripts` (Map of String) A map of scripts used for an OAuth connection. Only accepts a `fetchUserProfile` script.
- `session_key` (String) Session Key for storing the request token.
- `set_user_root_attributes` (String) Determines whether to sync user profile attributes (`name`, `given_name`, `family_name`, `nickname`, `picture`) at each login or only on the first login. Options include: `on_each_login`, `on_first_login`, `never_on_login`. Default value: `on_each_login`.
- `should_trust_email_verified_connection` (String) Choose how Auth0 sets the email_verified field in the user profile.
- `sign_in_endpoint` (String) SAML single login URL for the connection.
- `sign_out_endpoint` (String) SAML single logout URL for the connection.
- `sign_saml_request` (Boolean) When enabled, the SAML authentication request will be signed.
- `signature_algorithm` (String) Sign Request Algorithm.
- `signature_method` (String) Signature method used to sign the request
- `signing_cert` (String) X.509 signing certificate (encoded in PEM or CER) you retrieved from the IdP, Base64-encoded.
- `signing_key` (Block List, Max: 1) The key used to sign requests in the connection. Uses the `key` and `cert` properties to provide the private key and certificate respectively. (see [below for nested schema](#nestedblock--options--signing_key))
- `strategy_version` (Number) Version 1 is deprecated, use version 2.
- `subject` (String) Subject line of the email.
- `syntax` (String) Syntax of the template body.
- `team_id` (String) Apple Team ID.
- `template` (String) Body of the template.
- `tenant_domain` (String) Tenant domain name.
- `token_endpoint` (String) Token endpoint.
- `token_endpoint_auth_method` (String) Specifies the authentication method for the token endpoint. (Okta/OIDC Connections)
- `token_endpoint_auth_signing_alg` (String) Specifies the signing algorithm for the token endpoint. (Okta/OIDC Connections)
- `totp` (Block List, Max: 1) Configuration options for one-time passwords. (see [below for nested schema](#nestedblock--options--totp))
- `twilio_sid` (String) SID for your Twilio account.
- `twilio_token` (String, Sensitive) AuthToken for your Twilio account.
- `type` (String) Value can be `back_channel` or `front_channel`. Front Channel will use OIDC protocol with `response_mode=form_post` and `response_type=id_token`. Back Channel will use `response_type=code`.
- `upstream_params` (String) You can pass provider-specific parameters to an identity provider during authentication. The values can either be static per connection or dynamic per user.
- `use_cert_auth` (Boolean) Indicates whether to use cert auth or not.
- `use_kerberos` (Boolean) Indicates whether to use Kerberos or not.
- `use_wsfed` (Boolean) Whether to use WS-Fed.
- `user_authorization_url` (String) URL used to obtain user authorization.
- `user_id_attribute` (String) Attribute in the token that will be mapped to the user_id property in Auth0.
- `userinfo_endpoint` (String) User info endpoint.
- `validation` (Block List, Max: 1) Validation of the minimum and maximum values allowed for a user to have as username. (see [below for nested schema](#nestedblock--options--validation))
- `waad_common_endpoint` (Boolean) Indicates whether to use the common endpoint rather than the default endpoint. Typically enabled if you're using this for a multi-tenant application in Azure AD.
- `waad_protocol` (String) Protocol to use.

<a id="nestedblock--options--attribute_map"></a>
### Nested Schema for `options.attribute_map`

Required:

- `mapping_mode` (String) Method used to map incoming claims. Possible values: `use_map` (Okta or OIDC), `bind_all` (OIDC) or `basic_profile` (Okta).

Optional:

- `attributes` (String) This property is an object containing mapping information that allows Auth0 to interpret incoming claims from the IdP. Mapping information must be provided as key/value pairs.
- `userinfo_scope` (String) This property defines the scopes that Auth0 sends to the IdP’s UserInfo endpoint when requested.


<a id="nestedblock--options--attributes"></a>
### Nested Schema for `options.attributes`

Optional:

- `email` (Block List) Connection Options for Email Attribute (see [below for nested schema](#nestedblock--options--attributes--email))
- `phone_number` (Block List) Connection Options for Phone Number Attribute (see [below for nested schema](#nestedblock--options--attributes--phone_number))
- `username` (Block List) Connection Options for User Name Attribute (see [below for nested schema](#nestedblock--options--attributes--username))

<a id="nestedblock--options--attributes--email"></a>
### Nested Schema for `options.attributes.email`

Optional:

- `identifier` (Block List) Connection Options Email Attribute Identifier (see [below for nested schema](#nestedblock--options--attributes--email--identifier))
- `profile_required` (Boolean) Defines whether Profile is required
- `signup` (Block List) Defines signup settings for Email attribute (see [below for nested schema](#nestedblock--options--attributes--email--signup))
- `unique` (Boolean) If set to false, it allow multiple accounts with the same email address
- `verification_method` (String) Defines whether whether user will receive a link or an OTP during user signup for email verification and password reset for email verification

<a id="nestedblock--options--attributes--email--identifier"></a>
### Nested Schema for `options.attributes.email.identifier`

Optional:

- `active` (Boolean) Defines whether email attribute is active as an identifier
- `default_method` (String) Gets and Sets the default authentication method for the email identifier type. Valid values: `password`, `email_otp`


<a id="nestedblock--options--attributes--email--signup"></a>
### Nested Schema for `options.attributes.email.signup`

Optional:

- `status` (String) Defines signup status for Email Attribute
- `verification` (Block List) Defines settings for Verification under Email attribute (see [below for nested schema](#nestedblock--options--attributes--email--signup--verification))

<a id="nestedblock--options--attributes--email--signup--verification"></a>
### Nested Schema for `options.attributes.email.signup.verification`

Optional:

- `active` (Boolean) Defines verification settings for signup attribute




<a id="nestedblock--options--attributes--phone_number"></a>
### Nested Schema for `options.attributes.phone_number`

Optional:

- `identifier` (Block List) Connection Options Phone Number Attribute Identifier (see [below for nested schema](#nestedblock--options--attributes--phone_number--identifier))
- `profile_required` (Boolean) Defines whether Profile is required
- `signup` (Block List) Defines signup settings for Phone Number attribute (see [below for nested schema](#nestedblock--options--attributes--phone_number--signup))

<a id="nestedblock--options--attributes--phone_number--identifier"></a>
### Nested Schema for `options.attributes.phone_number.identifier`

Optional:

- `active` (Boolean) Defines whether Phone Number attribute is active as an identifier
- `default_method` (String) Gets and Sets the default authentication method for the phone_number identifier type. Valid values: `password`, `phone_otp`


<a id="nestedblock--options--attributes--phone_number--signup"></a>
### Nested Schema for `options.attributes.phone_number.signup`

Optional:

- `status` (String) Defines status of signup for Phone Number attribute
- `verification` (Block List) Defines verification settings for Phone Number attribute (see [below for nested schema](#nestedblock--options--attributes--phone_number--signup--verification))

<a id="nestedblock--options--attributes--phone_number--signup--verification"></a>
### Nested Schema for `options.attributes.phone_number.signup.verification`

Optional:

- `active` (Boolean) Defines verification settings for Phone Number attribute




<a id="nestedblock--options--attributes--username"></a>
### Nested Schema for `options.attributes.username`

Optional:

- `identifier` (Block List) Connection options for User Name Attribute Identifier (see [below for nested schema](#nestedblock--options--attributes--username--identifier))
- `profile_required` (Boolean) Defines whether Profile is required
- `signup` (Block List) Defines signup settings for User Name attribute (see [below for nested schema](#nestedblock--options--attributes--username--signup))
- `validation` (Block List) Defines validation settings for User Name attribute (see [below for nested schema](#nestedblock--options--attributes--username--validation))

<a id="nestedblock--options--attributes--username--identifier"></a>
### Nested Schema for `options.attributes.username.identifier`

Optional:

- `active` (Boolean) Defines whether UserName attribute is active as an identifier
- `default_method` (String) Gets and Sets the default authentication method for the username identifier type. Valid value: `password`


<a id="nestedblock--options--attributes--username--signup"></a>
### Nested Schema for `options.attributes.username.signup`

Optional:

- `status` (String) Defines whether User Name attribute is active as an identifier


<a id="nestedblock--options--attributes--username--validation"></a>
### Nested Schema for `options.attributes.username.validation`

Optional:

- `allowed_types` (Block List) Defines allowed types for for UserName attribute (see [below for nested schema](#nestedblock--options--attributes--username--validation--allowed_types))
- `max_length` (Number) Defines Max Length for User Name attribute
- `min_length` (Number) Defines Min Length for User Name attribute

<a id="nestedblock--options--attributes--username--validation--allowed_types"></a>
### Nested Schema for `options.attributes.username.validation.allowed_types`

Optional:

- `email` (Boolean) One of the allowed types for UserName signup attribute
- `phone_number` (Boolean) One of the allowed types for UserName signup attribute





<a id="nestedblock--options--authentication_methods"></a>
### Nested Schema for `options.authentication_methods`

Optional:

- `email_otp` (Block List, Max: 1) Configures Email OTP authentication (see [below for nested schema](#nestedblock--options--authentication_methods--email_otp))
- `passkey` (Block List, Max: 1) Configures passkey authentication (see [below for nested schema](#nestedblock--options--authentication_methods--passkey))
- `password` (Block List, Max: 1) Configures password authentication (see [below for nested schema](#nestedblock--options--authentication_methods--password))
- `phone_otp` (Block List, Max: 1) Configures Phone OTP authentication (see [below for nested schema](#nestedblock--options--authentication_methods--phone_otp))

<a id="nestedblock--options--authentication_methods--email_otp"></a>
### Nested Schema for `options.authentication_methods.email_otp`

Optional:

- `enabled` (Boolean) Enables Email OTP authentication


<a id="nestedblock--options--authentication_methods--passkey"></a>
### Nested Schema for `options.authentication_methods.passkey`

Optional:

- `enabled` (Boolean) Enables passkey authentication


<a id="nestedblock--options--authentication_methods--password"></a>
### Nested Schema for `options.authentication_methods.password`

Optional:

- `enabled` (Boolean) Enables password authentication


<a id="nestedblock--options--authentication_methods--phone_otp"></a>
### Nested Schema for `options.authentication_methods.phone_otp`

Optional:

- `enabled` (Boolean) Enables Phone OTP authentication



<a id="nestedblock--options--connection_settings"></a>
### Nested Schema for `options.connection_settings`

Required:

- `pkce` (String) PKCE configuration. Possible values: `auto` (uses the strongest algorithm available), `S256` (uses the SHA-256 algorithm), `plain` (uses plaintext as described in the PKCE specification) or `disabled` (disables support for PKCE).


<a id="nestedblock--options--custom_headers"></a>
### Nested Schema for `options.custom_headers`

Required:

- `header` (String)
- `value` (String)


<a id="nestedblock--options--decryption_key"></a>
### Nested Schema for `options.decryption_key`

Required:

- `cert` (String)
- `key` (String)


<a id="nestedblock--options--gateway_authentication"></a>
### Nested Schema for `options.gateway_authentication`

Optional:

- `audience` (String) Audience claim for the HS256 token sent to `gateway_url`.
- `method` (String) Authentication method (default is `bearer` token).
- `secret` (String, Sensitive) Secret used to sign the HS256 token sent to `gateway_url`.
- `secret_base64_encoded` (Boolean) Specifies whether or not the secret is Base64-encoded.
- `subject` (String) Subject claim for the HS256 token sent to `gateway_url`.


<a id="nestedblock--options--idp_initiated"></a>
### Nested Schema for `options.idp_initiated`

Optional:

- `client_authorize_query` (String)
- `client_id` (String)
- `client_protocol` (String)
- `enabled` (Boolean)


<a id="nestedblock--options--mfa"></a>
### Nested Schema for `options.mfa`

Optional:

- `active` (Boolean) Indicates whether multifactor authentication is enabled for this connection.
- `return_enroll_settings` (Boolean) Indicates whether multifactor authentication enrollment settings will be returned.


<a id="nestedblock--options--passkey_options"></a>
### Nested Schema for `options.passkey_options`

Optional:

- `challenge_ui` (String) Controls the UI used to challenge the user for their passkey
- `local_enrollment_enabled` (Boolean) Enables or disables enrollment prompt for local passkey when user authenticates using a cross-device passkey for the connection
- `progressive_enrollment_enabled` (Boolean) Enables or disables progressive enrollment of passkeys for the connection


<a id="nestedblock--options--password_complexity_options"></a>
### Nested Schema for `options.password_complexity_options`

Optional:

- `min_length` (Number) Minimum number of characters allowed in passwords.


<a id="nestedblock--options--password_dictionary"></a>
### Nested Schema for `options.password_dictionary`

Optional:

- `dictionary` (Set of String) Customized contents of the password dictionary. By default, the password dictionary contains a list of the [10,000 most common passwords](https://github.com/danielmiessler/SecLists/blob/master/Passwords/Common-Credentials/10k-most-common.txt); your customized content is used in addition to the default password dictionary. Matching is not case-sensitive.
- `enable` (Boolean) Indicates whether the password dictionary check is enabled for this connection.


<a id="nestedblock--options--password_history"></a>
### Nested Schema for `options.password_history`

Optional:

- `enable` (Boolean)
- `size` (Number)


<a id="nestedblock--options--password_no_personal_info"></a>
### Nested Schema for `options.password_no_personal_info`

Optional:

- `enable` (Boolean)


<a id="nestedblock--options--signing_key"></a>
### Nested Schema for `options.signing_key`

Required:

- `cert` (String)
- `key` (String)


<a id="nestedblock--options--totp"></a>
### Nested Schema for `options.totp`

Optional:

- `length` (Number) Length of the one-time password.
- `time_step` (Number) Seconds between allowed generation of new passwords.


<a id="nestedblock--options--validation"></a>
### Nested Schema for `options.validation`

Optional:

- `username` (Block List, Max: 1) Specifies the `min` and `max` values of username length. (see [below for nested schema](#nestedblock--options--validation--username))

<a id="nestedblock--options--validation--username"></a>
### Nested Schema for `options.validation.username`

Optional:

- `max` (Number)
- `min` (Number)

## Import

Import is supported using the following syntax:

```shell
# This resource can be imported by specifying the connection ID.
#
# Example:
terraform import auth0_connection.google "con_a17f21fdb24d48a0"
```
