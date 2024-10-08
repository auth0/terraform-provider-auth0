# Modifying the key_rotation_id causes the keys to be rotated/rekeyed.
resource "auth0_encryption_key_manager" "my_key_manager_initial" {
  key_rotation_id = "da9f2f3b-1c7e-4245-8982-9a25da8407c4"
}

resource "auth0_encryption_key_manager" "my_key_manager_rekey" {
  key_rotation_id = "68feba2c-7768-40f3-9d71-4b91e0233abf"
}

# To initialize the process of providing root key by the customer, create a
# `customer_provided_root_key` block.
resource "auth0_encryption_key_manager" "my_key_manager" {
  customer_provided_root_key {
  }
}

# The public_wrapping_key and wrapping_algorithm should be available to
# be used to wrap the new key by the customer
output "key_manager" {
  depends_on = [auth0_encryption_key_manager.my_key_manager]
  value = {
    public_wrapping_key = auth0_encryption_key_manager.my_key_manager.customer_provided_root_key.*.public_wrapping_key
    wrapping_algorithm  = auth0_encryption_key_manager.my_key_manager.customer_provided_root_key.*.wrapping_algorithm
  }
}

# The root key should be wrapped using the specified algorithm by the customer and Base64 encoded.
resource "auth0_encryption_key_manager" "my_key_manager" {
  customer_provided_root_key {
    wrapped_key = "miw4MHtx9BriXv4FDNOT930z0+MaK8HXvLI8clu0bS7LgfeLmAW8e59QP2QD1VfNTB7uvD5lYgsK92G3X5G95qNWJjZ8euEk1fM1+vtONQptqQyBdTWW4ZcJadaodASsJrSMXfSD+xJ3Lh45yEmkeENSDi60ZxKu5qUYuZmPWpEXeohPakJSm5X1qNVNLCOzBhNNG+OMEp8FVXtXnZTZVNtjbG2peVRpLlNGQkGfCWSY2VjpJkMcqf7DTRTF+USv9G1GHirRYkdVmlAOLfn/iwAHhIJlOqWYEhwkglIctMzX8mxW6VHCS3gptvcRk2j3eYNcw7BBrumuF+DE0NgQmmKaz0nRkHFRlv9RMRhk0qweHWPrp5Y2gCv+6du/m9FVMsNOSR0+4eSWsgOQw5B8gRs+4NfHm2N5sK2CRfzJ3mVNJjysaaag6TrTPbQjwlmcg5+DzeSc87Af5lwUvWT/kXPOGzVUNv9cF0FX7JM06UBQv5vfuU5zL/6VvszqCyjdxvbLgtGU1j/Hev++gKCfTQ8UcpegYxM6Ea60y4Qb3OezfdFE8R8eZg=="
  }
}

