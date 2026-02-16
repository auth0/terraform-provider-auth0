# Example: Retrieve all published versions of an action module

# Create and publish an action module
resource "auth0_action_module" "my_module" {
  name    = "My Shared Module"
  publish = true
  code    = <<-EOT
    module.exports = {
      greet: function(name) {
        return "Hello, " + name + "!";
      }
    };
  EOT
}

# Retrieve all published versions of the module
data "auth0_action_module_versions" "my_module_versions" {
  module_id = auth0_action_module.my_module.id
}

# Output the number of versions
output "total_versions" {
  value = length(data.auth0_action_module_versions.my_module_versions.versions)
}

# Output the latest version number
output "latest_version_number" {
  value = data.auth0_action_module_versions.my_module_versions.versions.0.version_number
}

# Output all version IDs
output "version_ids" {
  value = [for v in data.auth0_action_module_versions.my_module_versions.versions : v.id]
}

