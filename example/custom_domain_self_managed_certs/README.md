# Custom Domain with Self-managed Certificates on GCP

This example sets up an Auth0 tenant with a custom domain that uses self-managed certificates. It also configures the appropriate resources in Google Cloud Platform to forward requests to Auth0 over HTTPS.

To use this example, in addition to setting the usual Auth0 environment variables, you will also need to set `GOOGLE_PROJECT` and `GOOGLE_CREDENTIALS` (or equivalent; see [the provider reference](https://registry.terraform.io/providers/hashicorp/google/latest/docs/guides/provider_reference#authentication) for more information).

Note that Google-managed certificates take some time to provision. If everything in the configuration looks right but you're getting TLS errors trying to load your custom domain, you should wait 5-10 minutes and then refresh the page.
