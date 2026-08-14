# An empty data source block retrieves every organization of the tenant.
data "auth0_organizations" "my_organizations" {}

# Setting include_client_association_for additionally returns, for every organization
# associated with that client (application), the details of the association.
data "auth0_organizations" "my_organizations_with_client_association" {
  include_client_association_for = "AaiyAPdpYdesoKnqjj8HJqRn4T5titww"
}
