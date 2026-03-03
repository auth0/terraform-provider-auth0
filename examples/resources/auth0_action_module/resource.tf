resource "auth0_action_module" "my_module" {
  name = "My Shared Module"
  code = <<-EOT
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

