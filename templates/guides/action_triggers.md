---
page_title: Retrieve triggers available within actions
description: |-
Retrieve the set of triggers currently available within actions.
---

# Retrieving the set of triggers available within actions

In this guide we'll show how to retrieve the set of triggers currently available within actions.
A trigger is an extensibility point to which actions can be bound.

## Get an API Explorer Token

Head to the APIs section of your [Auth0 Dashboard](https://manage.auth0.com/#/apis) and click on the 
**Auth0 Management API**.

<img alt="get_api_explorer_token_1" src="https://user-images.githubusercontent.com/28300158/196163108-1a871ae9-7d98-4c4b-bb9f-a1fbf2e3fe8f.png">

Click on the **Create & Authorize Test Application** within the **API Explorer** tab.

<img alt="get_api_explorer_token_2" src="https://user-images.githubusercontent.com/28300158/196164274-785a39bf-e774-4d3d-a56b-c9b2b5c9149c.png">

Copy the **Token** contents and go to the [Management API Explorer](https://auth0.com/docs/api/management/v2).

<img alt="get_api_explorer_token_3" src="https://user-images.githubusercontent.com/28300158/196165167-e0e4b86d-8536-4613-9c61-7faec6dec8f9.png">

Click the **Set API Token** button at the top left.

Set the **API Token** by pasting the **Token** that you copied above.

Click the **Set Token** button.

<img alt="get_api_explorer_token_4" src="https://user-images.githubusercontent.com/28300158/196165357-6a2c3b69-2219-46eb-af2b-66665945bc75.png">

Retrieve the set of triggers available within actions by clicking on the **Try** button at 
[https://auth0.com/docs/api/management/v2#!/Actions/get_triggers](https://auth0.com/docs/api/management/v2#!/Actions/get_triggers).

<img alt="get_api_explorer_token_5" src="https://user-images.githubusercontent.com/28300158/196166349-8e4414ab-a110-4cf6-9343-dcc75a46146d.png">

At the time of writing (_2022-10-17_) the available triggers are the following:

```json
{
  "triggers": [
    {
      "id": "post-login",
      "version": "v2",
      "status": "DEPRECATED",
      "runtimes": [
        "node12",
        "node16"
      ],
      "default_runtime": "node16",
      "compatible_triggers": []
    },
    {
      "id": "post-login",
      "version": "v3",
      "status": "CURRENT",
      "runtimes": [
        "node12",
        "node16"
      ],
      "default_runtime": "node16",
      "compatible_triggers": [
        {
          "id": "post-login",
          "version": "v2"
        }
      ]
    },
    {
      "id": "post-login",
      "version": "v1",
      "status": "DEPRECATED",
      "runtimes": [
        "node12"
      ],
      "default_runtime": "node12",
      "compatible_triggers": []
    },
    {
      "id": "credentials-exchange",
      "version": "v1",
      "status": "DEPRECATED",
      "runtimes": [
        "node12"
      ],
      "default_runtime": "node12",
      "compatible_triggers": []
    },
    {
      "id": "credentials-exchange",
      "version": "v2",
      "status": "CURRENT",
      "runtimes": [
        "node12",
        "node16"
      ],
      "default_runtime": "node16",
      "compatible_triggers": []
    },
    {
      "id": "pre-user-registration",
      "version": "v2",
      "status": "CURRENT",
      "runtimes": [
        "node12",
        "node16"
      ],
      "default_runtime": "node16",
      "compatible_triggers": []
    },
    {
      "id": "pre-user-registration",
      "version": "v1",
      "status": "DEPRECATED",
      "runtimes": [
        "node12"
      ],
      "default_runtime": "node12",
      "compatible_triggers": []
    },
    {
      "id": "post-user-registration",
      "version": "v2",
      "status": "CURRENT",
      "runtimes": [
        "node12",
        "node16"
      ],
      "default_runtime": "node16",
      "compatible_triggers": []
    },
    {
      "id": "post-user-registration",
      "version": "v1",
      "status": "DEPRECATED",
      "runtimes": [
        "node12"
      ],
      "default_runtime": "node12",
      "compatible_triggers": []
    },
    {
      "id": "post-change-password",
      "version": "v2",
      "status": "CURRENT",
      "runtimes": [
        "node12",
        "node16"
      ],
      "default_runtime": "node16",
      "compatible_triggers": []
    },
    {
      "id": "post-change-password",
      "version": "v1",
      "status": "DEPRECATED",
      "runtimes": [
        "node12"
      ],
      "default_runtime": "node12",
      "compatible_triggers": []
    },
    {
      "id": "send-phone-message",
      "version": "v2",
      "status": "CURRENT",
      "runtimes": [
        "node12",
        "node16"
      ],
      "default_runtime": "node16",
      "compatible_triggers": []
    },
    {
      "id": "send-phone-message",
      "version": "v1",
      "status": "DEPRECATED",
      "runtimes": [
        "node12"
      ],
      "compatible_triggers": []
    }
  ]
}
```

Use these to set up your `supported_triggers` block within the `auth0_action` resource:

```terraform
resource "auth0_action" "my_action" {
  name    = format("Test Action %s", timestamp())
  runtime = "node16"
  code    = <<-EOT
   exports.onExecutePostLogin = async (event, api) => {
     console.log(event);
   };
  EOT

  supported_triggers {
    id      = "post-login"
    version = "v3"
  }
}
```
