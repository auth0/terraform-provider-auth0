---
page_title: Auto-generating Terraform config from Auth0 tenant
description: |-
  How to auto-generate Terraform configuration files from your Auth0 tenant
---

# Auto-generating Terraform config files from Auth0 tenant

~> This guide refers to Auth0 CLI functionality that is _experimental_ and may change in future versions.

Adopting Terraform for a mature Auth0 tenant can be daunting. Developers often face the challenge of manually importing numerous resources, necessitating the retrieval of multiple IDs. At times, they resort to manual Terraform configuration to match the tenant's resources.

Fortunately, the [Auth0 CLI](https://auth0.github.io/auth0-cli/) simplifies this process by auto-generating Terraform configuration files from an Auth0 tenant. This guide instructs developers on using the Auth0 CLI to auto-generate these files, enabling rapid transition to Terraform in minutes, not days.

## Pre-requisites:

- **Auth0 CLI v1.1.0+** – Auth0's official CLI. This tool will be performing the heavy lifting. Specifically requires versions 1.1.0 and up. See: [Auth0 CLI installation instructions](https://auth0.github.io/auth0-cli/).

## 1. Create Dedicated M2M Application for TF Provider

Establish an authenticated link between the Auth0 Terraform provider and the Auth0 tenant you wish to generate config for. This can be done by creating a dedicated machine-to-machine (M2M) application (client).

Follow the [Terraform Quickstart Guide](https://registry.terraform.io/providers/auth0/auth0/latest/docs/guides/quickstart#create-a-machine-to-machine-application) for instructions. Note the **domain**, **client ID**, and **client secret** values, as they are required in the next step.

## 2. Set Environment Variables for TF Provider

In your terminal, set the following environment variables, replacing `AUTH0_DOMAIN`, `AUTH0_CLIENT_ID`, and `AUTH0_CLIENT_SECRET` with the values noted in step 1:

```sh
export AUTH0_DOMAIN=***********
export AUTH0_CLIENT_ID=***********
export AUTH0_CLIENT_SECRET=***********
```

**Note:** Environment variables are the simplest and most secure way to pass credentials to the provider. Refer to the [related documentation](https://registry.terraform.io/providers/auth0/auth0/latest/docs#example-usage) for alternatives.

## 3. Authenticate with Auth0 CLI

Like the Terraform Provider, the Auth0 CLI requires an authentication link to the Auth0 tenant you wish to generate config for. To begin the authentication step, run:

```sh
auth0 login
```

Follow the interactive prompts to complete the authentication process.

Authenticating as a user is the simplest and quickest way to authenticate with the CLI but authenticating as a machine is also a valid option. However, it is recommended to use a separate machine-to-machine client than the one created in step 1.

~> It is required to authenticate the Auth0 CLI and TF provider to the same domain. The resulting auto-generated configuration is portable thereafter.

## 4. Run `tf generate` Command

With the Auth0 CLI authenticated to your tenant, initiate Terraform configuration auto-generation by running:

```
auth0 tf generate --output-dir tmp-auth0-tf
```

This command fetches relevant data from Auth0 and facilitates resource import using the Auth0 Terraform provider on the developer's behalf. You may want to isolate the auto-generated config from your current directory. In the above example, the output directory `tmp-auth0-tf` is used.

Follow the command's prompts and instructions. A successful run will produce an `auth0_generated.tf` file that produces no errors when `./terraform plan` is run. In certain cases it may be necessary to troubleshoot minor Terraform issues.

Some files are expected to be created during this process:

- `auth0_main.tf` – Establishes the Auth0 Terraform provider with specific versions for auto-generated config.
- `auth0_import.tf` – Contains all resources' import blocks, including names and IDs. Related: [Hashicorp Import Blocks](https://developer.hashicorp.com/terraform/language/import).
- `terraform` binary – A local Terraform binary instance for auto-generation. Pinned to a specific version.
- `auth0_generated.tf` – The final Terraform resource artifact representing your Auth0 tenant.

## 5. Review and Apply Terraform Config

Once you've executed the `auth0 tf generate` command and have created the `auth0_generated.tf` artifact, it is advised to spot-check it to ensure that all expected resources are present and no egregious errors exist.

It is at the developer's discretion to decide whether to immediately apply this generated configuration. Applying the configuration instructs Terraform to align your Auth0 tenant with the contents of these configuration files.

Alternatively, you can choose not to apply the configuration immediately. Instead, retain the generated Terraform configuration files for reference or modification. These files offer a valuable snapshot of your Auth0 tenant's configuration at the time of generation. Using the generated configuration files in this manner grants you the flexibility to adjust them to your needs or apply them to different Auth0 tenants as necessary.

## Note About Sensitive Values

While the generated Terraform config appears complete, it cannot export sensitive values like secrets and keys.

If configuring the same tenant as the one exported, immediate alterations may not be necessary. However, when applying this configuration to other tenants, you might need to supplement those values after the fact to ensure proper operation. Review the `auth0_generated.tf` file for properties commented with `# sensitive`. These are fields that may require replacements.

**Example:**
In the below example, both `credentials.access_key_id` and `credentials.api_key` properties are marked as sensitive.

```hcl
# __generated__ by Terraform from "52745e4d-278c-4b6b-8cac-a27e457215d6"
resource "auth0_email_provider" "email_provider" {
  default_from_address = "mailing-daemon@travel0.com"
  enabled              = true
  name                 = "smtp"
  credentials {
    access_key_id              = null # sensitive
    api_key                    = null # sensitive
  }
}
```
