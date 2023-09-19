data "auth0_signing_keys" "my_keys" {}

# Example on how to get the current key from the data source.
output "current_key" {
  value = try(
    element([for key in data.auth0_signing_keys.my_keys.signing_keys : key.kid if key.current], 0),
    "No current key found"
  )
}
