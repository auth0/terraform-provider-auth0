# This is an example of an SalesForce connection.

resource "auth0_connection" "salesforce" {
  name     = "Salesforce-Connection"
  strategy = "salesforce"

  options {
    client_id          = "<client-id>"
    client_secret      = "<client-secret>"
    community_base_url = "https://salesforce.example.com"
  }
}
