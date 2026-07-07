# Example: Retrieve a specific version of an action module

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



# Retrieve the latest published version using the module's version_id directly
data "auth0_action_module_version" "my_module_version" {
  module_id  = auth0_action_module.my_module.id
  version_id = auth0_action_module.my_module.version_id
}

# Output the version details
output "version_number" {
  value = data.auth0_action_module_version.my_module_version.version_number
}

output "version_code" {
  value = data.auth0_action_module_version.my_module_version.code
}

output "version_created_at" {
  value = data.auth0_action_module_version.my_module_version.created_at
}
