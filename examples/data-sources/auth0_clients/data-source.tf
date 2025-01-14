# Auth0 clients with "External" in the name
data "auth0_clients" "external_apps" {
  name_filter = "External"
}

# Auth0 clients filtered by non_interactive or spa app type
data "auth0_clients" "m2m_apps" {
  app_types = ["non_interactive", "spa"]
}

# Auth0 clients filtered by is_first_party equal to true
data "auth0_clients" "first_party_apps" {
  is_first_party = true
}