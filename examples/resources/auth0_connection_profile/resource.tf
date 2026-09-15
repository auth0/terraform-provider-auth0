resource "auth0_connection_profile" "my_profile" {
  name = "My-Profile"

  organization {
    show_as_button             = "optional"
    assign_membership_on_login = "required"
  }

  connection_name_prefix_template = "template1"

  enabled_features = [
    "scim",
    "universal_logout"
  ]

  # Requires the `my_orgs_cross_app_access_resource_app` tenant flag (EA only).
  cross_app_access_resource_app {
    status {
      default_value  = "enabled"
      allowed_values = ["enabled", "disabled"]
    }
  }
}
