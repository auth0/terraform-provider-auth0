terraform {
  required_providers {
    auth0 = {
      source = "auth0/auth0"
    }
  }
}

provider "auth0" {}

resource "auth0_organization" "organization" {
  name         = "auth0-inc"
  display_name = "Auth0 Inc."
  branding {
    logo_url = "https://example.com/assets/icons/icon.png"
    colors = {
      primary         = "#f2f2f2"
      page_background = "#e1e1e1"
    }
  }
  connections {
    connection_id = "con_X7iAWk1xB076gRi2"
  }
}

output "auth0_organization_id" {
  value = auth0_organization.organization.id
}
