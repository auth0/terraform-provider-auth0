# Data source to retrieve all prompt renderings
data "auth0_prompt_renderings" "all" {
}

# Data source to filter by prompt type
data "auth0_prompt_renderings" "login_passwordless" {
  prompt = "login-passwordless"
}

# Data source to filter by screen name
data "auth0_prompt_renderings" "login_screens" {
  screen = "login-id"
}

# Data source to filter by rendering mode
data "auth0_prompt_renderings" "advanced_only" {
  rendering_mode = "advanced"
}

# Data source with multiple filters
data "auth0_prompt_renderings" "filtered" {
  prompt         = "login-passwordless"
  rendering_mode = "advanced"
}

# Data source with wildcard syntax for prompt
data "auth0_prompt_renderings" "login_prompts" {
  prompt = "login"
}

# Data source with wildcard syntax for screen
data "auth0_prompt_renderings" "signup_screens" {
  screen = "signup-id"
}

# Access the renderings list
output "all_renderings" {
  value = data.auth0_prompt_renderings.all.renderings
}

# Access specific rendering properties
output "login_rendering_modes" {
  value = [for r in data.auth0_prompt_renderings.login_prompts.renderings : r.rendering_mode]
}
