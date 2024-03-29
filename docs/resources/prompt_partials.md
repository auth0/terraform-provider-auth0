---
page_title: "Resource: auth0_prompt_partials"
description: |-
  With this resource, you can manage a customized sign up and login experience by adding custom content, form elements and css/javascript. You can read more about this here https://auth0.com/docs/customize/universal-login-pages/customize-signup-and-login-prompts.
---

# Resource: auth0_prompt_partials

With this resource, you can manage a customized sign up and login experience by adding custom content, form elements and css/javascript. You can read more about this [here](https://auth0.com/docs/customize/universal-login-pages/customize-signup-and-login-prompts).

## Example Usage

```terraform
resource "auth0_prompt_partials" "my_login_prompt_partials" {
  prompt = "login"

  form_content_start      = "<div>Updated Form Content Start</div>"
  form_content_end        = "<div>Updated Form Content End</div>"
  form_footer_start       = "<div>Updated Footer Start</div>"
  form_footer_end         = "<div>Updated Footer End</div>"
  secondary_actions_start = "<div>Updated Secondary Actions Start</div>"
  secondary_actions_end   = "<div>Updated Secondary Actions End</div>"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `prompt` (String) The prompt that you are adding partials for. Options are: `login-id`, `login`, `login-password`, `signup`, `signup-id`, `signup-password`.

### Optional

- `form_content_end` (String) Content that goes at the end of the form.
- `form_content_start` (String) Content that goes at the start of the form.
- `form_footer_end` (String) Footer content for the end of the footer.
- `form_footer_start` (String) Footer content for the start of the footer.
- `secondary_actions_end` (String) Actions that go at the end of secondary actions.
- `secondary_actions_start` (String) Actions that go at the start of secondary actions.

### Read-Only

- `id` (String) The ID of this resource.

## Import

Import is supported using the following syntax:

```shell
# This resource can be imported using the prompt name.
#
# Example:
terraform import auth0_prompt_partials.my_login_prompt_partials "login"
```
