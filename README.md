# terraform-provider-auth0 v2

> Plugin Framework rewrite of [`auth0/terraform-provider-auth0`](https://github.com/auth0/terraform-provider-auth0).
> This is the **v2.x** line, built on top of [`go-auth0`](https://github.com/auth0/go-auth0) **v2**.
> The legacy SDKv2 provider continues to live on the `v1` branch and is published as `auth0/auth0` `<= 1.x`.

## Status

Early scaffolding. Currently implemented:

- [x] Provider configuration (domain, audience, client_id/secret, api_token, private-key JWT, debug)
- [x] `auth0_client` resource (subset of fields — name, description, app_type, callbacks, allowed_logout_urls, allowed_origins, web_origins, grant_types, token_endpoint_auth_method, client_metadata, etc.)
- [x] `auth0_organization` resource (name, display_name, branding, metadata)

More resources will follow incrementally.

## Local development & testing

### 1. Build & install the binary

```sh
make install
```

This compiles the provider into `$GOPATH/bin/terraform-provider-auth0`.

### 2. Configure a Terraform CLI dev override

Add this to `~/.terraformrc` so Terraform picks up your locally-built binary
instead of downloading from the registry:

```hcl
provider_installation {
  dev_overrides {
    "auth0/auth0" = "/Users/<you>/go/bin"
  }
  direct {}
}
```

### 3. Run the example

```sh
cd examples/basic
export AUTH0_DOMAIN="your-tenant.auth0.com"
export AUTH0_CLIENT_ID="..."
export AUTH0_CLIENT_SECRET="..."
terraform plan
terraform apply
```

> With dev overrides you do **not** run `terraform init`; Terraform will print a
> warning about the override and use your local binary directly.

## Authentication

The provider supports three authentication modes (auto-detected from config):

| Mode               | Required attributes / env vars                                                                |
| ------------------ | --------------------------------------------------------------------------------------------- |
| Static token       | `api_token` / `AUTH0_API_TOKEN`                                                               |
| Client credentials | `client_id` + `client_secret` (`AUTH0_CLIENT_ID`, `AUTH0_CLIENT_SECRET`)                      |
| Private Key JWT    | `client_id` + `client_assertion_private_key` (+ optional `client_assertion_signing_alg`)      |

`api_token` takes precedence, then Private Key JWT, then client secret.

## License

MIT — see [LICENSE](./LICENSE).

