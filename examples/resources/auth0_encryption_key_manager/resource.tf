resource "auth0_encryption_key_manager" "my_encryption_key_manager_initial" {
  key_rotation_id = "da9f2f3b-1c7e-4245-8982-9a25da8407c4"
}

resource "auth0_encryption_key_manager" "my_encryption_key_manager_rekey" {
  key_rotation_id = "68feba2c-7768-40f3-9d71-4b91e0233abf"
}

