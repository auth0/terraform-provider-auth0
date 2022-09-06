---
page_title: "Getting Started"
description: |-
Quickly get started with the Auth0 Provider.
---

# Getting started

In this guide we'll go through setting up an Auth0 Application for our provider to authenticate against and manage
resources.

## Create a Machine to Machine Application

For Terraform to be able to create resources in Auth0, you'll need to manually create an Auth0 Machine-to-Machine
Application that allows Terraform to communicate with Auth0.

Head to the Applications section of your [Auth0 Dashboard](https://manage.auth0.com/#/applications) and click the
"Create Application" button on the top right.

<img alt="create_app1" src="https://user-images.githubusercontent.com/28300158/183633275-88a5ae17-64e4-4352-8b9c-f8f62ba50a97.png">

In the form that pops up, give your app a name like "Terraform Provider Auth0" and select 
"Machine to Machine Application" as the type. Click the "Create" button to be taken to the next screen.

<img alt="create_app2" src="https://user-images.githubusercontent.com/28300158/183634949-cabdfe6e-93cf-42f1-bfdb-b0f216c2642c.png">

You'll need to authorize your new app to call the Auth0 Management API. Select it in the dropdown and then authorize all
scopes by clicking "All" in the top right of the scopes selection area. Click the "Authorize" button to continue.

<img alt="create_app3" src="https://user-images.githubusercontent.com/28300158/183635167-724ea60e-117d-47a5-a18c-746f402ee52a.png">

You'll be taken to the details page for your new application. Open the "Settings" tab and copy the Domain, Client ID,
and Client Secret values - you'll need them in the next step for configuring the Auth0 Provider.

<img alt="create_app4" src="https://user-images.githubusercontent.com/28300158/183635366-bee78296-cb7f-4586-b0a5-067aaa3ea578.png">


## Configure the Provider

Although you can put passwords, secrets, and other credentials directly into Terraform configuration files, hard-coding
credentials into any Terraform configuration is not recommended, and risks secret leakage should this file ever be 
committed to a public version control system. Because of this you'll set your Auth0 Application credentials as
environment variables instead.

In the terminal window where you're running Terraform, run the following commands, substituting `AUTHO_DOMAIN`,
`CLIENT_ID`, and `CLIENT_SECRET` for your M2M app's values:

```shell
export AUTH0_DOMAIN=***********
export AUTH0_CLIENT_ID=***********
export AUTH0_CLIENT_SECRET=***********
```

After you've set your environment variables, head back to your text editor, and add the following in `main.tf`:

```terraform
terraform {
  required_providers {
    auth0 = {
      source  = "auth0/auth0"
      version = "~> 0.34.0"
    }
  }
}

provider "auth0" {}
```

The Auth0 Provider will communicate with the Auth0 Management API using the M2M credentials you've provided. 
Moreover, we specify the version range that we want to allow for the provider, to prevent an uncontrolled update.

Now run the following to initialize your terraform configuration:

```shell
terraform init
```

## Manage resources through the Provider

Now you can start adding the Auth0 resources you want to manage through terraform. As an example let's create a new
Web Application in our Auth0 Tenant.

In our `main.tf` from above, let's append the following:

```terraform
// We are appending on main.tf

resource "auth0_client" "my_client" {
  name            = "WebAppExample"
  description     = "My Web App Created Through Terraform"
  app_type        = "regular_web"
  callbacks       = ["http://localhost:3000/callback"]
  oidc_conformant = true

  jwt_configuration {
    alg = "RS256"
  }
}
```

With the new resource in place, you can run the following terminal commands to apply your configuration:

```shell
terraform apply
```

After apply finishes, you can verify that the application was created by going to the
[Auth0 Dashboard Applications page](https://manage.auth0.com/#/applications). You should see a new application called
"WebAppExample", as specified in the name argument passed to the resource.
