terraform {
  required_providers {
    auth0 = {
      source = "auth0/auth0"
    }
  }
}

provider "auth0" {}

resource "auth0_prompt" "prompt" {
  universal_login_experience = "classic"
  identifier_first           = false
}
