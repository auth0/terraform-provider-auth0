resource "auth0_action_module" "my_module" {
  name    = "My Shared Module"
  publish = true
  code    = <<-EOT
  /**
   * A shared utility function that can be used across multiple actions.
   */
  module.exports = {
    greet: function(name) {
      return "Hello, " + name + "!";
    },
    formatDate: function(date) {
      return date.toISOString();
    }
  };
  EOT

  dependencies {
    name    = "lodash"
    version = "4.17.21"
  }

  secrets {
    name  = "API_KEY"
    value = "my-secret-api-key"
  }
}

# Use the module in an action by referencing its id and version_id.
resource "auth0_action" "my_action" {
  name    = "My Action"
  runtime = "node22"
  deploy  = true
  code    = <<-EOT
  const myModule = require('My Shared Module');

  exports.onExecutePostLogin = async (event, api) => {
    console.log(myModule.greet(event.user.name));
  };
  EOT

  modules {
    module_id         = auth0_action_module.my_module.id
    module_version_id = auth0_action_module.my_module.version_id
  }

  supported_triggers {
    id      = "post-login"
    version = "v3"
  }
}
